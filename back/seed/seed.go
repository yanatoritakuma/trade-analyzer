// seed はローカル動作確認用のサンプルデータを投入する。
// 実行: docker compose -f docker-compose.dev.yml exec app go run seed/seed.go
//
// 投入内容: 管理者/一般ユーザー、招待コード、ウォッチリスト、株価スナップショット、
// トレード（バーチャル/実運用・決済済/未決済）、実運用保有株、分析シグナル、
// 週次レポート、分析テーマ、分析設定、ウォッチリスト候補。
package main

import (
	"log"
	"time"

	"github.com/joho/godotenv"
	"github.com/lib/pq"
	"gorm.io/datatypes"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/yanatoritakuma/trade-analyzer/back/db"
	"github.com/yanatoritakuma/trade-analyzer/back/model"
	"github.com/yanatoritakuma/trade-analyzer/back/utils"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println(".env を読み込めませんでした（環境変数を直接使用します）")
	}
	database := db.NewDB()
	defer db.CloseDB(database)

	if err := seed(database); err != nil {
		log.Fatalln("seed failed:", err)
	}
	log.Println("Seed completed")
}

func seed(database *gorm.DB) error {
	jst := utils.JST
	now := time.Now().In(jst)
	weekStart := utils.CurrentWeekStartJST()

	pwHash, err := utils.HashPassword("password123")
	if err != nil {
		return err
	}

	// ---- ユーザー ----
	admin := model.User{Email: "admin@example.com", Name: "管理者", PasswordHash: pwHash, Role: "admin", IsActive: true}
	if err := upsertUser(database, &admin); err != nil {
		return err
	}
	user := model.User{Email: "user@example.com", Name: "一般ユーザー", PasswordHash: pwHash, Role: "user", IsActive: true}
	if err := upsertUser(database, &user); err != nil {
		return err
	}

	// ---- 招待コード ----
	invitations := []model.InvitationCode{
		{Code: "TRADE-DEMO-0001", CreatedBy: &admin.ID, ExpiresAt: now.AddDate(0, 0, 14), IsActive: true},
		{Code: "TRADE-DEMO-0002", CreatedBy: &admin.ID, ExpiresAt: now.AddDate(0, 0, 7), IsActive: true},
	}
	for i := range invitations {
		if err := upsertByField(database, "code", invitations[i].Code, &invitations[i]); err != nil {
			return err
		}
	}

	// ---- ウォッチリスト ----
	watchlist := []model.Watchlist{
		{Ticker: "7203.T", Name: "トヨタ自動車", Mode: "both", IsActive: true},
		{Ticker: "6758.T", Name: "ソニーグループ", Mode: "both", IsActive: true},
		{Ticker: "9984.T", Name: "ソフトバンクグループ", Mode: "virtual", IsActive: true},
	}
	for i := range watchlist {
		if err := upsertByField(database, "ticker", watchlist[i].Ticker, &watchlist[i]); err != nil {
			return err
		}
	}

	// ---- 株価スナップショット ----
	prices := []model.StockPrice{
		{Ticker: "7203.T", Date: now, Open: 2800, High: 2870, Low: 2790, Close: 2850, PrevClose: 2800, ChangeAmount: 50, ChangeRate: 1.79, Volume: 12000000},
		{Ticker: "6758.T", Date: now, Open: 13500, High: 13700, Low: 13400, Close: 13600, PrevClose: 13800, ChangeAmount: -200, ChangeRate: -1.45, Volume: 8000000},
		{Ticker: "9984.T", Date: now, Open: 9000, High: 9200, Low: 8950, Close: 9100, PrevClose: 9050, ChangeAmount: 50, ChangeRate: 0.55, Volume: 15000000},
	}
	for i := range prices {
		if err := database.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "ticker"}},
			DoUpdates: clause.AssignmentColumns([]string{"date", "open", "high", "low", "close", "prev_close", "change_amount", "change_rate", "volume", "updated_at"}),
		}).Create(&prices[i]).Error; err != nil {
			return err
		}
	}

	// ---- 実運用保有株（admin） ----
	positions := []model.RealPosition{
		{UserID: admin.ID, Ticker: "7203.T", Name: "トヨタ自動車", Quantity: 100, AvgPrice: 2700},
		{UserID: admin.ID, Ticker: "6758.T", Name: "ソニーグループ", Quantity: 10, AvgPrice: 13000},
	}
	for i := range positions {
		if err := database.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "user_id"}, {Name: "ticker"}},
			DoUpdates: clause.AssignmentColumns([]string{"name", "quantity", "avg_price", "updated_at"}),
		}).Create(&positions[i]).Error; err != nil {
			return err
		}
	}

	// ---- トレード（admin・共通ポートフォリオ）。再実行時は一旦削除して投入 ----
	if err := database.Unscoped().Where("user_id = ?", admin.ID).Delete(&model.Trade{}).Error; err != nil {
		return err
	}
	closed1 := weekStart.AddDate(0, 0, 1)
	closed2 := now.AddDate(0, 0, -10)
	trades := []model.Trade{
		// バーチャル・決済済（今週）
		{UserID: admin.ID, Ticker: "7203.T", Name: "トヨタ自動車", Mode: "virtual", Action: "SELL", Price: 2850, Quantity: 100, Confidence: 0.82, Reason: "目標株価に到達したため利益確定。", TargetPrice: 2850, StopLoss: 2650, ResultPnl: 15000, ClosedAt: &closed1, Model: gorm.Model{CreatedAt: weekStart}},
		// バーチャル・決済済（過去・損失）
		{UserID: admin.ID, Ticker: "6758.T", Name: "ソニーグループ", Mode: "virtual", Action: "SELL", Price: 13200, Quantity: 10, Confidence: 0.65, Reason: "損切りラインに到達。", TargetPrice: 14000, StopLoss: 13200, ResultPnl: -8000, ClosedAt: &closed2, Model: gorm.Model{CreatedAt: now.AddDate(0, 0, -14)}},
		// バーチャル・未決済（保有中）
		{UserID: admin.ID, Ticker: "9984.T", Name: "ソフトバンクグループ", Mode: "virtual", Action: "BUY", Price: 8900, Quantity: 20, Confidence: 0.71, Reason: "上昇トレンド継続を期待。", TargetPrice: 9800, StopLoss: 8500, Model: gorm.Model{CreatedAt: now.AddDate(0, 0, -3)}},
		// 実運用・決済済（今週）
		{UserID: admin.ID, Ticker: "7203.T", Name: "トヨタ自動車", Mode: "real", Action: "SELL", Price: 2820, Quantity: 50, Confidence: 0.78, Reason: "短期上昇を確定。", TargetPrice: 2820, StopLoss: 2680, ResultPnl: 6000, ClosedAt: &closed1, Model: gorm.Model{CreatedAt: weekStart}},
		// 実運用・未決済
		{UserID: admin.ID, Ticker: "6758.T", Name: "ソニーグループ", Mode: "real", Action: "BUY", Price: 13000, Quantity: 10, Confidence: 0.69, Reason: "押し目買い。", TargetPrice: 14500, StopLoss: 12500, Model: gorm.Model{CreatedAt: now.AddDate(0, 0, -5)}},
	}
	for i := range trades {
		if err := database.Create(&trades[i]).Error; err != nil {
			return err
		}
	}

	// ---- 分析シグナル（analysis_logs）。trades.created_atと日付一致でJOINされる ----
	if err := database.Unscoped().Where("1 = 1").Delete(&model.AnalysisLog{}).Error; err != nil {
		return err
	}
	buyAnalysis := datatypes.JSON([]byte(`{"buy_reasons":["移動平均が上向き","出来高増加"],"no_buy_reasons":["地合いが弱い"],"entry_condition":"押し目で2800円を割らなければ買い"}`))
	conf78 := 0.78
	conf82 := 0.82
	analysisLogs := []model.AnalysisLog{
		{Ticker: "9984.T", Action: "BUY", Confidence: conf78, Analysis: buyAnalysis, Model: gorm.Model{CreatedAt: now.AddDate(0, 0, -3)}},
		{Ticker: "7203.T", Action: "SELL", Confidence: conf82, Analysis: buyAnalysis, Model: gorm.Model{CreatedAt: weekStart}},
		{Ticker: "6758.T", Action: "HOLD", Confidence: 0.5, Analysis: datatypes.JSON([]byte(`{}`)), Model: gorm.Model{CreatedAt: now.AddDate(0, 0, -1)}},
	}
	for i := range analysisLogs {
		if err := database.Create(&analysisLogs[i]).Error; err != nil {
			return err
		}
	}

	// ---- 週次レポート（learning_logs） ----
	lastWeekStart := weekStart.AddDate(0, 0, -7)
	lastWeekEnd := weekStart.AddDate(0, 0, -1)
	reports := []model.LearningLog{
		{WeekStart: lastWeekStart, WeekEnd: lastWeekEnd, TradeCount: 4, WinRate: 75, TotalPnl: 23000,
			Summary: "上昇トレンド銘柄での順張りが奏功した1週間。", Lessons: "利確の目標株価設定が適切だった。", Strategy: "来週も順張りを継続しつつ、地合い悪化時は早めの損切りを徹底する。"},
	}
	for i := range reports {
		var count int64
		database.Model(&model.LearningLog{}).Where("week_start = ?", reports[i].WeekStart.Format("2006-01-02")).Count(&count)
		if count == 0 {
			if err := database.Create(&reports[i]).Error; err != nil {
				return err
			}
		}
	}

	// ---- 分析テーマ ----
	themes := []model.AnalysisTheme{
		{Name: "半導体", Description: "半導体関連銘柄", SortOrder: 1, IsActive: true, CreatedBy: &admin.ID},
		{Name: "AI・データセンター", Description: "AI需要の恩恵銘柄", SortOrder: 2, IsActive: true, CreatedBy: &admin.ID},
		{Name: "高配当", Description: "配当利回りの高い銘柄", SortOrder: 3, IsActive: true, CreatedBy: &admin.ID},
	}
	for i := range themes {
		if err := upsertByField(database, "name", themes[i].Name, &themes[i]); err != nil {
			return err
		}
	}

	// ---- 分析設定（単一active） ----
	var settingCount int64
	database.Model(&model.AnalysisSetting{}).Where("is_active = ?", true).Count(&settingCount)
	if settingCount == 0 {
		setting := model.AnalysisSetting{
			ThemeIDs:   pq.Int64Array{int64(themes[0].ID), int64(themes[1].ID)},
			Screening:  datatypes.JSON([]byte(`{"min_market_cap":100000000000,"min_volume":1000000,"max_per":25}`)),
			Style:      "short_term_trend",
			FreePrompt: "決算発表が近い銘柄は慎重に判断してください。",
			IsActive:   true,
			CreatedBy:  &admin.ID,
		}
		if err := database.Create(&setting).Error; err != nil {
			return err
		}
	}

	// ---- ウォッチリスト候補 ----
	if err := database.Unscoped().Where("1 = 1").Delete(&model.WatchlistCandidate{}).Error; err != nil {
		return err
	}
	candidates := []model.WatchlistCandidate{
		{Ticker: "8035.T", Name: "東京エレクトロン", Reason: "半導体製造装置の需要拡大が見込まれるため。", ReplaceTicker: "9984.T", Confidence: 0.84, Status: "pending", ProposedAt: now.AddDate(0, 0, -1)},
		{Ticker: "6098.T", Name: "リクルートHD", Reason: "業績好調・出来高増加。", Confidence: 0.76, Status: "pending", ProposedAt: now.AddDate(0, 0, -1)},
	}
	for i := range candidates {
		if err := database.Create(&candidates[i]).Error; err != nil {
			return err
		}
	}

	log.Printf("管理者: admin@example.com / password123")
	log.Printf("一般ユーザー: user@example.com / password123")
	log.Printf("招待コード: TRADE-DEMO-0001 / TRADE-DEMO-0002")
	return nil
}

// upsertUser はメールアドレスで既存ユーザーを引き当て、無ければ作成する。
func upsertUser(database *gorm.DB, u *model.User) error {
	var existing model.User
	err := database.Where("email = ?", u.Email).First(&existing).Error
	if err == nil {
		u.ID = existing.ID
		return nil
	}
	if err != gorm.ErrRecordNotFound {
		return err
	}
	return database.Create(u).Error
}

// upsertByField は指定フィールドで既存を引き当て、無ければ作成してIDを埋める。
func upsertByField(database *gorm.DB, field, value string, dest interface{}) error {
	if err := database.Where(field+" = ?", value).First(dest).Error; err == nil {
		return nil
	} else if err != gorm.ErrRecordNotFound {
		return err
	}
	return database.Create(dest).Error
}
