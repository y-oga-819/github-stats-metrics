package pull_request

import (
	"context"
	"log"

	domain "github-stats-metrics/domain/pull_request"
)

// UseCase はPull Request関連のビジネスロジックを統括
type UseCase struct {
	repo domain.Repository
}

// NewUseCase はUseCaseのコンストラクタ（依存性注入）
func NewUseCase(repo domain.Repository) *UseCase {
	return &UseCase{
		repo: repo,
	}
}

// GetPullRequests はPull Requestsを取得し、ビジネスルールを適用
func (uc *UseCase) GetPullRequests(ctx context.Context, req domain.GetPullRequestsRequest) ([]domain.PullRequest, error) {
	// リクエストバリデーション
	if err := req.Validate(); err != nil {
		return nil, NewValidationError("invalid request parameters", err)
	}
	
	log.Printf("Fetching pull requests for developers: %v, period: %s to %s", 
		req.Developers, req.StartDate, req.EndDate)

	// リポジトリから取得（抽象に依存）
	pullRequests, err := uc.repo.GetPullRequests(ctx, req)
	if err != nil {
		return nil, NewRepositoryError("failed to fetch pull requests", err)
	}

	// ビジネスルールの適用（例：epicブランチの除外）
	filtered := uc.applyBusinessRules(pullRequests)

	log.Printf("Retrieved %d pull requests (filtered from %d)", len(filtered), len(pullRequests))
	return filtered, nil
}

// applyBusinessRules はビジネスルールを適用してPRをフィルタリング
func (uc *UseCase) applyBusinessRules(pullRequests []domain.PullRequest) []domain.PullRequest {
	var filtered []domain.PullRequest
	
	for _, pr := range pullRequests {
		if !uc.shouldExcludePR(pr) {
			filtered = append(filtered, pr)
		}
	}
	
	return filtered
}

// shouldExcludePR はPRを除外すべきかを判定するビジネスルール
func (uc *UseCase) shouldExcludePR(pr domain.PullRequest) bool {
	// ルール1: epicブランチは除外
	if isEpicBranch(pr.HeadRefName) {
		return true
	}
	
	// ルール2: マージされていないPRは除外
	if !pr.IsMerged() {
		return true
	}
	
	return false
}

// isEpicBranch はepicブランチかを判定
func isEpicBranch(branchName string) bool {
	return len(branchName) > 5 && branchName[:5] == "epic/"
}

// GetAvailableDevelopers は利用可能な開発者一覧を取得
func (uc *UseCase) GetAvailableDevelopers(ctx context.Context) ([]string, error) {
	repos, err := uc.repo.GetRepositories(ctx)
	if err != nil {
		return nil, NewRepositoryError("failed to get target repositories", err)
	}
	
	developers, err := uc.repo.GetDevelopers(ctx, repos)
	if err != nil {
		return nil, NewRepositoryError("failed to get developers", err)
	}
	
	return developers, nil
}