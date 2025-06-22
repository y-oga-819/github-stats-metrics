package github_api

import (
	"time"
	
	domain "github-stats-metrics/domain/pull_request"
	"github.com/shurcooL/githubv4"
)

// GitHub API用の内部構造体（Infrastructure層でのみ使用）
type githubv4PullRequest struct {
	Id          githubv4.String
	Number      githubv4.Int
	Title       githubv4.String
	BaseRefName githubv4.String
	HeadRefName githubv4.String
	Author      struct {
		Login     githubv4.String
		AvatarURL githubv4.URI `graphql:"avatarUrl(size:72)"`
	}
	Repository struct {
		Name githubv4.String
	}
	URL           githubv4.URI
	Additions     githubv4.Int
	Deletions     githubv4.Int
	CreatedAt     githubv4.DateTime
	FirstReviewed struct {
		Nodes []struct {
			CreatedAt githubv4.DateTime
		}
	} `graphql:"FirstReviewed: reviews(first: 1)"`
	LastApprovedAt struct {
		Nodes []struct {
			CreatedAt githubv4.DateTime
		}
	} `graphql:"LastApprovedAt: reviews(last: 1, states: APPROVED)"`
	MergedAt githubv4.DateTime
}

// convertToDomain はGitHub APIレスポンスをDomainモデルに変換
func convertToDomain(apiPR githubv4PullRequest) domain.PullRequest {
	pr := domain.PullRequest{
		ID:          string(apiPR.Id),
		Number:      int(apiPR.Number),
		Title:       string(apiPR.Title),
		BaseRefName: string(apiPR.BaseRefName),
		HeadRefName: string(apiPR.HeadRefName),
		Author: domain.Author{
			Login:     string(apiPR.Author.Login),
			AvatarURL: apiPR.Author.AvatarURL.String(),
		},
		Repository: domain.RepositoryInfo{
			Name: string(apiPR.Repository.Name),
		},
		URL:       apiPR.URL.String(),
		Additions: int(apiPR.Additions),
		Deletions: int(apiPR.Deletions),
		CreatedAt: apiPR.CreatedAt.Time,
	}
	
	// オプション値の適切な変換
	if !apiPR.MergedAt.Time.IsZero() {
		pr.MergedAt = &apiPR.MergedAt.Time
	}
	
	if len(apiPR.FirstReviewed.Nodes) > 0 && !apiPR.FirstReviewed.Nodes[0].CreatedAt.Time.IsZero() {
		pr.FirstReviewed = &apiPR.FirstReviewed.Nodes[0].CreatedAt.Time
	}
	
	if len(apiPR.LastApprovedAt.Nodes) > 0 && !apiPR.LastApprovedAt.Nodes[0].CreatedAt.Time.IsZero() {
		pr.LastApproved = &apiPR.LastApprovedAt.Nodes[0].CreatedAt.Time
	}
	
	return pr
}

// timePtr はtime.Timeのポインタを返す（nil値の適切な処理）
func timePtr(t time.Time) *time.Time {
	if t.IsZero() {
		return nil
	}
	return &t
}