package usecase

import (
	"context"
	"fmt"

	"github.com/yanatoritakuma/trade-analyzer/back/domain"
	"github.com/yanatoritakuma/trade-analyzer/back/domain/watchlist"
)

const maxWatchlist = 3

// WatchlistUsecase はウォッチリストのユースケース。
type WatchlistUsecase struct {
	uow           UnitOfWork
	watchlistRepo watchlist.WatchlistRepository
}

func NewWatchlistUsecase(uow UnitOfWork, watchlistRepo watchlist.WatchlistRepository) *WatchlistUsecase {
	return &WatchlistUsecase{uow: uow, watchlistRepo: watchlistRepo}
}

// FindAllWithPrice はウォッチリストに現在値・前日比を付与して返す（閲覧・全員）。
func (u *WatchlistUsecase) FindAllWithPrice(ctx context.Context) ([]*watchlist.Watchlist, error) {
	return u.watchlistRepo.FindAllWithPrice()
}

// FindAllForInternal は内部API（Lambda）向けに is_active=true の銘柄を返す。
func (u *WatchlistUsecase) FindAllForInternal(ctx context.Context) ([]*watchlist.Watchlist, error) {
	return u.watchlistRepo.FindAll()
}

// Create は4桁コードに .T を付与して登録する（上限3・重複不可）。
func (u *WatchlistUsecase) Create(ctx context.Context, code string, mode watchlist.Mode) (*watchlist.Watchlist, error) {
	if !mode.IsValid() {
		return nil, domain.NewMessageError(domain.ErrInvalidInput, "モードを選択してください")
	}
	ticker := code + ".T"
	var created *watchlist.Watchlist
	err := u.uow.Do(ctx, func(repos *Repositories) error {
		count, err := repos.Watchlist.CountActive()
		if err != nil {
			return err
		}
		if count >= maxWatchlist {
			return domain.NewMessageError(domain.ErrInvalidInput, "最大3銘柄まで登録できます")
		}
		exists, err := repos.Watchlist.ExistsByTicker(ticker)
		if err != nil {
			return err
		}
		if exists {
			return domain.NewMessageError(domain.ErrAlreadyExists, "すでに登録されている銘柄です")
		}
		w := &watchlist.Watchlist{Ticker: ticker, Mode: mode, IsActive: true}
		if err := repos.Watchlist.Save(w); err != nil {
			return fmt.Errorf("Create watchlist: %w", err)
		}
		created = w
		return nil
	})
	if err != nil {
		return nil, err
	}
	return created, nil
}

// Delete はウォッチリスト銘柄を削除する。
func (u *WatchlistUsecase) Delete(ctx context.Context, id uint) error {
	if _, err := u.watchlistRepo.FindByID(id); err != nil {
		return err
	}
	return u.watchlistRepo.Delete(id)
}
