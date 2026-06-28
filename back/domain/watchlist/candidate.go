package watchlist

import "time"

// CandidateStatus は候補ステータス。
type CandidateStatus string

const (
	CandidatePending  CandidateStatus = "pending"
	CandidateApproved CandidateStatus = "approved"
	CandidateRejected CandidateStatus = "rejected"
)

// Candidate はAIが提案したウォッチリスト候補エンティティ。
type Candidate struct {
	ID            uint
	Ticker        string
	Name          string
	Reason        string
	ReplaceTicker string
	Confidence    float64
	Status        CandidateStatus
	ProposedAt    time.Time
	DecidedAt     *time.Time
	DecidedBy     *uint
}

// Approve は承認状態にする。
func (c *Candidate) Approve(adminID uint) {
	now := time.Now()
	c.Status = CandidateApproved
	c.DecidedAt = &now
	c.DecidedBy = &adminID
}

// Reject は却下状態にする。
func (c *Candidate) Reject(adminID uint) {
	now := time.Now()
	c.Status = CandidateRejected
	c.DecidedAt = &now
	c.DecidedBy = &adminID
}
