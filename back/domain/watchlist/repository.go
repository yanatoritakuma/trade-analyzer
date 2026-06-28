package watchlist

// WatchlistRepository は監視銘柄永続化のインターフェース（全ユーザー共通）。
type WatchlistRepository interface {
	FindAll() ([]*Watchlist, error)
	FindAllWithPrice() ([]*Watchlist, error)
	FindByID(id uint) (*Watchlist, error)
	CountActive() (int64, error)
	ExistsByTicker(ticker string) (bool, error)
	Save(w *Watchlist) error
	Delete(id uint) error
	DeleteByTicker(ticker string) error
}

// CandidateRepository は候補銘柄永続化のインターフェース。
type CandidateRepository interface {
	FindAll() ([]*Candidate, error)
	FindByID(id uint) (*Candidate, error)
	Save(c *Candidate) error
	Update(c *Candidate) error
}
