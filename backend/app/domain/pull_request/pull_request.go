package pull_request

import (
	"time"
)

type PullRequest struct {
	ID          string
	Number      int
	Title       string
	BaseRefName string
	HeadRefName string
	Author      Author
	Repository  Repository
	URL         string
	Additions   int
	Deletions   int
	CreatedAt   time.Time
	FirstReviewed *time.Time
	LastApproved  *time.Time
	MergedAt     *time.Time
}

type Author struct {
	Login     string
	AvatarURL string
}

type Repository struct {
	Name string
}

// ドメインロジック
func (pr PullRequest) IsReviewed() bool {
	return pr.FirstReviewed != nil
}

func (pr PullRequest) IsApproved() bool {
	return pr.LastApproved != nil
}

func (pr PullRequest) IsMerged() bool {
	return pr.MergedAt != nil
}

func (pr PullRequest) ReviewTime() *time.Duration {
	if pr.FirstReviewed == nil {
		return nil
	}
	duration := pr.FirstReviewed.Sub(pr.CreatedAt)
	return &duration
}

func (pr PullRequest) ApprovalTime() *time.Duration {
	if pr.FirstReviewed == nil || pr.LastApproved == nil {
		return nil
	}
	duration := pr.LastApproved.Sub(*pr.FirstReviewed)
	return &duration
}

func (pr PullRequest) MergeTime() *time.Duration {
	if pr.LastApproved == nil || pr.MergedAt == nil {
		return nil
	}
	duration := pr.MergedAt.Sub(*pr.LastApproved)
	return &duration
}
