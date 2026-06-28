package usecase

import (
	"context"

	"github.com/yanatoritakuma/trade-analyzer/back/domain"
	"github.com/yanatoritakuma/trade-analyzer/back/domain/watchlist"
)

// CandidateUsecase はウォッチリスト候補承認のユースケース。
type CandidateUsecase struct {
	uow           UnitOfWork
	candidateRepo watchlist.CandidateRepository
}

func NewCandidateUsecase(uow UnitOfWork, candidateRepo watchlist.CandidateRepository) *CandidateUsecase {
	return &CandidateUsecase{uow: uow, candidateRepo: candidateRepo}
}

// List は候補一覧を返す。
func (u *CandidateUsecase) List(ctx context.Context) ([]*watchlist.Candidate, error) {
	return u.candidateRepo.FindAll()
}

// Approve は候補を承認し watchlist を更新する（UoWで一括・上限3維持）。
func (u *CandidateUsecase) Approve(ctx context.Context, id uint, adminID uint, mode watchlist.Mode) (*watchlist.Candidate, error) {
	if mode == "" {
		mode = watchlist.ModeBoth
	}
	if !mode.IsValid() {
		return nil, domain.NewMessageError(domain.ErrInvalidInput, "モードが不正です")
	}
	var result *watchlist.Candidate
	err := u.uow.Do(ctx, func(repos *Repositories) error {
		cand, err := repos.WatchlistCandidate.FindByID(id)
		if err != nil {
			return err
		}
		if cand.Status != watchlist.CandidatePending {
			return domain.NewMessageError(domain.ErrInvalidInput, "この候補はすでに処理されています")
		}

		// replace 指定があれば該当銘柄を削除
		if cand.ReplaceTicker != "" {
			if err := repos.Watchlist.DeleteByTicker(cand.ReplaceTicker); err != nil {
				return err
			}
		}

		// 上限3チェック
		count, err := repos.Watchlist.CountActive()
		if err != nil {
			return err
		}
		if count >= maxWatchlist {
			return domain.NewMessageError(domain.ErrInvalidInput, "最大3銘柄まで登録できます")
		}

		// 重複チェック（既に登録済みならスキップして承認のみ）
		exists, err := repos.Watchlist.ExistsByTicker(cand.Ticker)
		if err != nil {
			return err
		}
		if !exists {
			w := &watchlist.Watchlist{
				Ticker:   cand.Ticker,
				Name:     cand.Name,
				Mode:     mode,
				IsActive: true,
			}
			if err := repos.Watchlist.Save(w); err != nil {
				return err
			}
		}

		cand.Approve(adminID)
		if err := repos.WatchlistCandidate.Update(cand); err != nil {
			return err
		}
		result = cand
		return nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

// Reject は候補を却下する（即時）。
func (u *CandidateUsecase) Reject(ctx context.Context, id uint, adminID uint) (*watchlist.Candidate, error) {
	cand, err := u.candidateRepo.FindByID(id)
	if err != nil {
		return nil, err
	}
	cand.Reject(adminID)
	if err := u.candidateRepo.Update(cand); err != nil {
		return nil, err
	}
	return cand, nil
}
