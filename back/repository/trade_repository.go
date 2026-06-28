package repository

import (
	"encoding/json"
	"time"

	"gorm.io/gorm"

	"github.com/yanatoritakuma/trade-analyzer/back/domain/trade"
	"github.com/yanatoritakuma/trade-analyzer/back/model"
)

// adminSubQuery は admin の user_id を絞り込むサブクエリ条件。
const adminSubQuery = "user_id IN (SELECT id FROM users WHERE role = 'admin' AND deleted_at IS NULL)"

// TradeRepositoryImpl は trade.TradeRepository のGORM実装。
type TradeRepositoryImpl struct {
	db *gorm.DB
}

func NewTradeRepositoryImpl(db *gorm.DB) trade.TradeRepository {
	return &TradeRepositoryImpl{db: db}
}

func ptrFloat(v float64) *float64 {
	if v == 0 {
		return nil
	}
	return &v
}

func ptrStr(v string) *string {
	if v == "" {
		return nil
	}
	return &v
}

func toTradeEntity(m *model.Trade) *trade.Trade {
	t := &trade.Trade{
		ID:          m.ID,
		UserID:      m.UserID,
		Ticker:      m.Ticker,
		Name:        m.Name,
		Mode:        trade.Mode(m.Mode),
		Action:      trade.Action(m.Action),
		Price:       m.Price,
		Quantity:    m.Quantity,
		Confidence:  ptrFloat(m.Confidence),
		Reason:      ptrStr(m.Reason),
		TargetPrice: ptrFloat(m.TargetPrice),
		StopLoss:    ptrFloat(m.StopLoss),
		ClosedAt:    m.ClosedAt,
		CreatedAt:   m.CreatedAt,
	}
	if m.ClosedAt != nil {
		pnl := m.ResultPnl
		t.ResultPnl = &pnl
	}
	return t
}

func (r *TradeRepositoryImpl) FindByFilter(f trade.Filter) ([]*trade.Trade, *trade.Summary, error) {
	q := r.db.Model(&model.Trade{}).Where(adminSubQuery).Where("mode = ?", string(f.Mode))
	if f.Ticker != nil && *f.Ticker != "" {
		q = q.Where("ticker = ?", *f.Ticker)
	}
	if f.Action != nil && *f.Action != "" {
		q = q.Where("action = ?", string(*f.Action))
	}
	if f.From != nil {
		q = q.Where("created_at >= ?", *f.From)
	}
	if f.To != nil {
		// to は終端日の終わりまで含める
		q = q.Where("created_at < ?", f.To.Add(24*time.Hour))
	}

	var ms []model.Trade
	if err := q.Order("created_at DESC").Find(&ms).Error; err != nil {
		return nil, nil, err
	}

	trades := make([]*trade.Trade, 0, len(ms))
	var totalPnl float64
	closed, win := 0, 0
	for i := range ms {
		t := toTradeEntity(&ms[i])
		trades = append(trades, t)
		if ms[i].ClosedAt != nil {
			closed++
			totalPnl += ms[i].ResultPnl
			if ms[i].ResultPnl > 0 {
				win++
			}
		}
	}
	summary := &trade.Summary{Count: len(ms), TotalPnl: totalPnl}
	if closed > 0 {
		summary.WinRate = float64(win) / float64(closed) * 100
	}

	if err := r.attachAnalysis(trades); err != nil {
		return nil, nil, err
	}
	return trades, summary, nil
}

// attachAnalysis は analysis_logs の buy/no_buy理由・エントリー条件を各トレードに付与する。
func (r *TradeRepositoryImpl) attachAnalysis(trades []*trade.Trade) error {
	if len(trades) == 0 {
		return nil
	}
	tickerSet := map[string]struct{}{}
	for _, t := range trades {
		tickerSet[t.Ticker] = struct{}{}
	}
	tickers := make([]string, 0, len(tickerSet))
	for tk := range tickerSet {
		tickers = append(tickers, tk)
	}

	var logs []model.AnalysisLog
	if err := r.db.Where("ticker IN ?", tickers).Find(&logs).Error; err != nil {
		return err
	}

	type detail struct {
		buy       []string
		noBuy     []string
		entryCond *string
	}
	byKey := map[string]detail{}
	for i := range logs {
		key := logs[i].Ticker + "|" + logs[i].CreatedAt.Format("2006-01-02")
		var parsed struct {
			BuyReasons     []string `json:"buy_reasons"`
			NoBuyReasons   []string `json:"no_buy_reasons"`
			EntryCondition string   `json:"entry_condition"`
		}
		if len(logs[i].Analysis) > 0 {
			_ = json.Unmarshal(logs[i].Analysis, &parsed)
		}
		byKey[key] = detail{
			buy:       parsed.BuyReasons,
			noBuy:     parsed.NoBuyReasons,
			entryCond: ptrStr(parsed.EntryCondition),
		}
	}

	for _, t := range trades {
		key := t.Ticker + "|" + t.CreatedAt.Format("2006-01-02")
		if d, ok := byKey[key]; ok {
			t.BuyReasons = d.buy
			t.NoBuyReasons = d.noBuy
			t.EntryCondition = d.entryCond
		}
	}
	return nil
}

func (r *TradeRepositoryImpl) AggregateByAdmin(mode trade.Mode, weekStart time.Time) (*trade.ModeAggregate, error) {
	var row struct {
		TotalPnl    float64
		WeeklyPnl   float64
		TradeCount  int
		ClosedCount int
		WinCount    int
	}
	err := r.db.Model(&model.Trade{}).
		Select(`
			COALESCE(SUM(CASE WHEN closed_at IS NOT NULL THEN result_pnl ELSE 0 END), 0) AS total_pnl,
			COALESCE(SUM(CASE WHEN closed_at IS NOT NULL AND closed_at >= ? THEN result_pnl ELSE 0 END), 0) AS weekly_pnl,
			COUNT(*) AS trade_count,
			COUNT(CASE WHEN closed_at IS NOT NULL THEN 1 END) AS closed_count,
			COUNT(CASE WHEN closed_at IS NOT NULL AND result_pnl > 0 THEN 1 END) AS win_count
		`, weekStart).
		Where(adminSubQuery).Where("mode = ?", string(mode)).
		Scan(&row).Error
	if err != nil {
		return nil, err
	}
	agg := &trade.ModeAggregate{
		TotalPnl:   row.TotalPnl,
		WeeklyPnl:  row.WeeklyPnl,
		TradeCount: row.TradeCount,
	}
	if row.ClosedCount > 0 {
		agg.WinRate = float64(row.WinCount) / float64(row.ClosedCount) * 100
	}
	return agg, nil
}

func (r *TradeRepositoryImpl) FindOpenVirtualPositions() ([]*trade.OpenPosition, error) {
	// 未決済バーチャルを銘柄ごとに数量合計・加重平均取得単価で集計する。
	var rows []struct {
		Ticker   string
		Name     string
		Quantity int
		AvgPrice float64
	}
	err := r.db.Model(&model.Trade{}).
		Select(`ticker,
			MAX(name) AS name,
			SUM(quantity) AS quantity,
			CASE WHEN SUM(quantity) > 0 THEN SUM(price * quantity) / SUM(quantity) ELSE 0 END AS avg_price`).
		Where(adminSubQuery).
		Where("mode = ? AND action = ? AND closed_at IS NULL", "virtual", "BUY").
		Group("ticker").
		Having("SUM(quantity) > 0").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	out := make([]*trade.OpenPosition, 0, len(rows))
	for _, row := range rows {
		out = append(out, &trade.OpenPosition{
			Ticker:   row.Ticker,
			Name:     row.Name,
			Quantity: row.Quantity,
			AvgPrice: row.AvgPrice,
		})
	}
	return out, nil
}

func (r *TradeRepositoryImpl) FindClosedForTimeseries(mode trade.Mode, from *time.Time) ([]*trade.Trade, error) {
	q := r.db.Model(&model.Trade{}).
		Where(adminSubQuery).
		Where("mode = ? AND closed_at IS NOT NULL", string(mode))
	if from != nil {
		q = q.Where("closed_at >= ?", *from)
	}
	var ms []model.Trade
	if err := q.Order("closed_at ASC").Find(&ms).Error; err != nil {
		return nil, err
	}
	out := make([]*trade.Trade, 0, len(ms))
	for i := range ms {
		out = append(out, toTradeEntity(&ms[i]))
	}
	return out, nil
}

func (r *TradeRepositoryImpl) FindByAdminBetween(start, end time.Time) ([]*trade.Trade, error) {
	var ms []model.Trade
	err := r.db.Model(&model.Trade{}).
		Where(adminSubQuery).
		Where("created_at >= ? AND created_at < ?", start, end.Add(24*time.Hour)).
		Order("created_at ASC").
		Find(&ms).Error
	if err != nil {
		return nil, err
	}
	out := make([]*trade.Trade, 0, len(ms))
	for i := range ms {
		out = append(out, toTradeEntity(&ms[i]))
	}
	return out, nil
}

func (r *TradeRepositoryImpl) Save(t *trade.Trade) error {
	m := model.Trade{
		UserID:   t.UserID,
		Ticker:   t.Ticker,
		Name:     t.Name,
		Mode:     string(t.Mode),
		Action:   string(t.Action),
		Price:    t.Price,
		Quantity: t.Quantity,
		ClosedAt: t.ClosedAt,
	}
	if t.Confidence != nil {
		m.Confidence = *t.Confidence
	}
	if t.Reason != nil {
		m.Reason = *t.Reason
	}
	if t.TargetPrice != nil {
		m.TargetPrice = *t.TargetPrice
	}
	if t.StopLoss != nil {
		m.StopLoss = *t.StopLoss
	}
	if t.ResultPnl != nil {
		m.ResultPnl = *t.ResultPnl
	}
	// seed等で作成日時を明示指定したい場合に対応（autoCreateTimeは値があれば上書きしない）。
	if !t.CreatedAt.IsZero() {
		m.CreatedAt = t.CreatedAt
	}
	if err := r.db.Create(&m).Error; err != nil {
		return err
	}
	t.ID = m.ID
	t.CreatedAt = m.CreatedAt
	return nil
}
