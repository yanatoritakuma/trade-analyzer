package controller

import (
	"time"

	"github.com/yanatoritakuma/trade-analyzer/back/domain/analysis"
	"github.com/yanatoritakuma/trade-analyzer/back/domain/invitation"
	"github.com/yanatoritakuma/trade-analyzer/back/domain/learning"
	"github.com/yanatoritakuma/trade-analyzer/back/domain/position"
	"github.com/yanatoritakuma/trade-analyzer/back/domain/trade"
	"github.com/yanatoritakuma/trade-analyzer/back/domain/user"
	"github.com/yanatoritakuma/trade-analyzer/back/domain/watchlist"
	"github.com/yanatoritakuma/trade-analyzer/back/usecase"
)

// ---- User ----

type userDTO struct {
	ID    uint   `json:"id"`
	Email string `json:"email"`
	Name  string `json:"name"`
	Role  string `json:"role"`
}

func toUserDTO(u *user.User) userDTO {
	return userDTO{ID: u.ID, Email: u.Email, Name: u.Name, Role: string(u.Role)}
}

type adminUserDTO struct {
	ID        uint      `json:"id"`
	Email     string    `json:"email"`
	Name      string    `json:"name"`
	Role      string    `json:"role"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
}

func toAdminUserDTO(u *user.User) adminUserDTO {
	return adminUserDTO{
		ID: u.ID, Email: u.Email, Name: u.Name, Role: string(u.Role),
		IsActive: u.IsActive, CreatedAt: u.CreatedAt,
	}
}

// ---- Portfolio ----

func toPortfolioSummaryDTO(s *usecase.PortfolioSummary) gin_h {
	return gin_h{"virtual": s.Virtual, "real": s.Real}
}

// ---- Analysis signal ----

type analysisSignalDTO struct {
	Ticker     string    `json:"ticker"`
	Name       *string   `json:"name"`
	Action     string    `json:"action"`
	Confidence *float64  `json:"confidence"`
	AnalyzedAt time.Time `json:"analyzed_at"`
}

func toAnalysisSignalDTO(a *analysis.AnalysisLog) analysisSignalDTO {
	return analysisSignalDTO{
		Ticker: a.Ticker, Name: a.Name, Action: string(a.Action),
		Confidence: a.Confidence, AnalyzedAt: a.AnalyzedAt,
	}
}

// ---- Watchlist ----

type watchlistItemDTO struct {
	ID         uint     `json:"id"`
	Ticker     string   `json:"ticker"`
	Name       *string  `json:"name"`
	Mode       string   `json:"mode"`
	Close      *float64 `json:"close"`
	ChangeRate *float64 `json:"change_rate"`
}

func toWatchlistItemDTO(w *watchlist.Watchlist) watchlistItemDTO {
	var name *string
	if w.Name != "" {
		n := w.Name
		name = &n
	}
	return watchlistItemDTO{
		ID: w.ID, Ticker: w.Ticker, Name: name, Mode: string(w.Mode),
		Close: w.Close, ChangeRate: w.ChangeRate,
	}
}

// ---- Position ----

type positionDTO struct {
	ID            uint     `json:"id"`
	Ticker        string   `json:"ticker"`
	Name          *string  `json:"name"`
	Quantity      int      `json:"quantity"`
	AvgPrice      float64  `json:"avg_price"`
	Close         *float64 `json:"close"`
	UnrealizedPnl *float64 `json:"unrealized_pnl"`
	PnlRate       *float64 `json:"pnl_rate"`
}

func toPositionDTO(p *position.Position) positionDTO {
	var name *string
	if p.Name != "" {
		n := p.Name
		name = &n
	}
	return positionDTO{
		ID: p.ID, Ticker: p.Ticker, Name: name, Quantity: p.Quantity, AvgPrice: p.AvgPrice,
		Close: p.Close, UnrealizedPnl: p.UnrealizedPnl, PnlRate: p.PnlRate,
	}
}

// ---- Trade ----

type tradeDTO struct {
	ID             uint       `json:"id"`
	Ticker         string     `json:"ticker"`
	Name           *string    `json:"name"`
	Mode           string     `json:"mode"`
	Action         string     `json:"action"`
	Price          float64    `json:"price"`
	Quantity       int        `json:"quantity"`
	Confidence     *float64   `json:"confidence"`
	Reason         *string    `json:"reason"`
	TargetPrice    *float64   `json:"target_price"`
	StopLoss       *float64   `json:"stop_loss"`
	ResultPnl      *float64   `json:"result_pnl"`
	ClosedAt       *time.Time `json:"closed_at"`
	CreatedAt      time.Time  `json:"created_at"`
	BuyReasons     []string   `json:"buy_reasons"`
	NoBuyReasons   []string   `json:"no_buy_reasons"`
	EntryCondition *string    `json:"entry_condition"`
}

func toTradeDTO(t *trade.Trade) tradeDTO {
	var name *string
	if t.Name != "" {
		n := t.Name
		name = &n
	}
	return tradeDTO{
		ID: t.ID, Ticker: t.Ticker, Name: name, Mode: string(t.Mode), Action: string(t.Action),
		Price: t.Price, Quantity: t.Quantity, Confidence: t.Confidence, Reason: t.Reason,
		TargetPrice: t.TargetPrice, StopLoss: t.StopLoss, ResultPnl: t.ResultPnl,
		ClosedAt: t.ClosedAt, CreatedAt: t.CreatedAt,
		BuyReasons: t.BuyReasons, NoBuyReasons: t.NoBuyReasons, EntryCondition: t.EntryCondition,
	}
}

func toTradeDTOs(ts []*trade.Trade) []tradeDTO {
	out := make([]tradeDTO, 0, len(ts))
	for _, t := range ts {
		out = append(out, toTradeDTO(t))
	}
	return out
}

// ---- Report ----

type reportSummaryDTO struct {
	WeekStart  string  `json:"week_start"`
	WeekEnd    string  `json:"week_end"`
	TradeCount int     `json:"trade_count"`
	WinRate    float64 `json:"win_rate"`
	TotalPnl   float64 `json:"total_pnl"`
}

func toReportSummaryDTO(l *learning.LearningLog) reportSummaryDTO {
	return reportSummaryDTO{
		WeekStart: l.WeekStart.Format("2006-01-02"), WeekEnd: l.WeekEnd.Format("2006-01-02"),
		TradeCount: l.TradeCount, WinRate: l.WinRate, TotalPnl: l.TotalPnl,
	}
}

// ---- Invitation ----

type invitationDTO struct {
	ID        uint       `json:"id"`
	Code      string     `json:"code"`
	ExpiresAt time.Time  `json:"expires_at"`
	UsedBy    *uint      `json:"used_by"`
	UsedAt    *time.Time `json:"used_at"`
	IsActive  bool       `json:"is_active"`
	Status    string     `json:"status"`
	CreatedAt time.Time  `json:"created_at"`
}

func toInvitationDTO(i *invitation.InvitationCode) invitationDTO {
	return invitationDTO{
		ID: i.ID, Code: i.Code, ExpiresAt: i.ExpiresAt, UsedBy: i.UsedBy, UsedAt: i.UsedAt,
		IsActive: i.IsActive, Status: i.Status(), CreatedAt: i.CreatedAt,
	}
}

// ---- Theme ----

type themeDTO struct {
	ID          uint      `json:"id"`
	Name        string    `json:"name"`
	Description *string   `json:"description"`
	SortOrder   int       `json:"sort_order"`
	IsActive    bool      `json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func toThemeDTO(t *analysis.Theme) themeDTO {
	var desc *string
	if t.Description != "" {
		d := t.Description
		desc = &d
	}
	return themeDTO{
		ID: t.ID, Name: t.Name, Description: desc, SortOrder: t.SortOrder,
		IsActive: t.IsActive, CreatedAt: t.CreatedAt, UpdatedAt: t.UpdatedAt,
	}
}

// ---- Analysis setting ----

type analysisSettingDTO struct {
	ID         uint                `json:"id"`
	ThemeIDs   []int64             `json:"theme_ids"`
	Screening  *analysis.Screening `json:"screening"`
	Style      string              `json:"style"`
	FreePrompt *string             `json:"free_prompt"`
	IsActive   bool                `json:"is_active"`
	UpdatedAt  time.Time           `json:"updated_at"`
}

func toAnalysisSettingDTO(s *analysis.Setting) analysisSettingDTO {
	var fp *string
	if s.FreePrompt != "" {
		v := s.FreePrompt
		fp = &v
	}
	ids := s.ThemeIDs
	if ids == nil {
		ids = []int64{}
	}
	return analysisSettingDTO{
		ID: s.ID, ThemeIDs: ids, Screening: s.Screening, Style: string(s.Style),
		FreePrompt: fp, IsActive: s.IsActive, UpdatedAt: s.UpdatedAt,
	}
}

// ---- Candidate ----

type candidateDTO struct {
	ID            uint       `json:"id"`
	Ticker        string     `json:"ticker"`
	Name          *string    `json:"name"`
	Reason        *string    `json:"reason"`
	ReplaceTicker *string    `json:"replace_ticker"`
	Confidence    *float64   `json:"confidence"`
	Status        string     `json:"status"`
	ProposedAt    time.Time  `json:"proposed_at"`
	DecidedAt     *time.Time `json:"decided_at"`
	DecidedBy     *uint      `json:"decided_by"`
}

func toCandidateDTO(c *watchlist.Candidate) candidateDTO {
	strPtr := func(s string) *string {
		if s == "" {
			return nil
		}
		return &s
	}
	var conf *float64
	if c.Confidence != 0 {
		v := c.Confidence
		conf = &v
	}
	return candidateDTO{
		ID: c.ID, Ticker: c.Ticker, Name: strPtr(c.Name), Reason: strPtr(c.Reason),
		ReplaceTicker: strPtr(c.ReplaceTicker), Confidence: conf, Status: string(c.Status),
		ProposedAt: c.ProposedAt, DecidedAt: c.DecidedAt, DecidedBy: c.DecidedBy,
	}
}

// gin_h は gin.H の別名（dto.go内で簡易マップを返すため）。
type gin_h = map[string]interface{}
