// バックエンドのレスポンスDTOに対応するフロント側の型定義。
// openapi.yaml と整合（src/types/api.ts は openapi-typescript の自動生成型）。

export type Mode = 'virtual' | 'real' | 'both';
export type TradeMode = 'virtual' | 'real';
export type Action = 'BUY' | 'SELL';
export type SignalAction = 'BUY' | 'SELL' | 'HOLD';

export type ModeSummary = {
  total_pnl: number;
  weekly_pnl: number;
  win_rate: number;
  trade_count: number;
};

export type PortfolioSummary = {
  virtual: ModeSummary;
  real: ModeSummary;
};

export type AnalysisSignal = {
  ticker: string;
  name: string | null;
  action: SignalAction;
  confidence: number | null;
  analyzed_at: string;
};

export type WatchlistItem = {
  id: number;
  ticker: string;
  name: string | null;
  mode: Mode;
  close: number | null;
  change_rate: number | null;
};

export type Position = {
  id: number;
  ticker: string;
  name: string | null;
  quantity: number;
  avg_price: number;
  close: number | null;
  unrealized_pnl: number | null;
  pnl_rate: number | null;
};

export type Trade = {
  id: number;
  ticker: string;
  name: string | null;
  mode: TradeMode;
  action: Action;
  price: number;
  quantity: number;
  confidence: number | null;
  reason: string | null;
  target_price: number | null;
  stop_loss: number | null;
  result_pnl: number | null;
  closed_at: string | null;
  created_at: string;
  buy_reasons: string[] | null;
  no_buy_reasons: string[] | null;
  entry_condition: string | null;
};

export type TradeListResponse = {
  items: Trade[];
  summary: { count: number; win_rate: number; total_pnl: number };
};

export type ReportSummary = {
  week_start: string;
  week_end: string;
  trade_count: number;
  win_rate: number;
  total_pnl: number;
};

export type ReportDetail = ReportSummary & {
  max_drawdown: number | null;
  summary: string;
  lessons: string;
  strategy: string;
  trades: Trade[];
};

export type AdminUser = {
  id: number;
  email: string;
  name: string;
  role: 'admin' | 'user';
  is_active: boolean;
  created_at: string;
};

export type Invitation = {
  id: number;
  code: string;
  expires_at: string;
  used_by: number | null;
  used_at: string | null;
  is_active: boolean;
  status: 'valid' | 'used' | 'expired' | 'disabled';
  created_at: string;
};

export type Theme = {
  id: number;
  name: string;
  description: string | null;
  sort_order: number;
  is_active: boolean;
  created_at: string;
  updated_at: string;
};

export type Screening = {
  min_market_cap: number;
  min_volume: number;
  max_per: number;
};

export type AnalysisStyle =
  | 'short_term_trend'
  | 'short_term_contrarian'
  | 'short_term_both'
  | 'mid_term_trend'
  | 'mid_term_contrarian'
  | 'mid_term_both';

export type AnalysisSetting = {
  id: number;
  theme_ids: number[];
  screening: Screening | null;
  style: AnalysisStyle;
  free_prompt: string | null;
  is_active: boolean;
  updated_at: string;
};

export type WatchlistCandidate = {
  id: number;
  ticker: string;
  name: string | null;
  reason: string | null;
  replace_ticker: string | null;
  confidence: number | null;
  status: 'pending' | 'approved' | 'rejected';
  proposed_at: string;
  decided_at: string | null;
  decided_by: number | null;
};
