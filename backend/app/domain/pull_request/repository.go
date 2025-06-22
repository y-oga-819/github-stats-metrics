package pull_request

import (
	"context"
)

// Repository はPull Requestデータアクセスの抽象化
type Repository interface {
	// GetPullRequests は条件に基づいてPull Requestsを取得
	GetPullRequests(ctx context.Context, req GetPullRequestsRequest) ([]PullRequest, error)
	
	// GetPullRequestByID は特定IDのPull Requestを取得
	GetPullRequestByID(ctx context.Context, id string) (*PullRequest, error)
	
	// GetRepositories は対象リポジトリ一覧を取得
	GetRepositories(ctx context.Context) ([]string, error)
	
	// GetDevelopers は開発者一覧を取得
	GetDevelopers(ctx context.Context, repositories []string) ([]string, error)
}