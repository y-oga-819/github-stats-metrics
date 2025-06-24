package pull_request

import (
	"context"
	"log"

	domain "github-stats-metrics/domain/pull_request"
	"github-stats-metrics/shared/errors"
)

// UseCase はPull Request関連のビジネスロジックを統括
type UseCase struct {
	repo    domain.Repository
	service *domain.PullRequestService
}

// NewUseCase はUseCaseのコンストラクタ（依存性注入）
func NewUseCase(repo domain.Repository) *UseCase {
	return &UseCase{
		repo:    repo,
		service: domain.NewPullRequestService(),
	}
}

// GetPullRequests はPull Requestsを取得し、ビジネスルールを適用
func (uc *UseCase) GetPullRequests(ctx context.Context, req domain.GetPullRequestsRequest) ([]domain.PullRequest, error) {
	// リクエストバリデーション
	if err := req.Validate(); err != nil {
		return nil, errors.NewValidationError(errors.ErrCodeInvalidRequest, "invalid request parameters", err.Error())
	}
	
	log.Printf("Fetching pull requests for developers: %v, period: %s to %s", 
		req.Developers, req.StartDate, req.EndDate)

	// リポジトリから取得（抽象に依存）
	pullRequests, err := uc.repo.GetPullRequests(ctx, req)
	if err != nil {
		return nil, errors.NewRepositoryError(errors.ErrCodeExternalAPIError, "failed to fetch pull requests", err)
	}

	// ドメインサービスでビジネスルールを適用
	filtered := uc.service.FilterByBusinessRules(pullRequests)

	log.Printf("Retrieved %d pull requests (filtered from %d)", len(filtered), len(pullRequests))
	return filtered, nil
}

// GetPullRequestMetrics はPull Requestsのメトリクスを取得
func (uc *UseCase) GetPullRequestMetrics(ctx context.Context, req domain.GetPullRequestsRequest) (domain.PullRequestMetrics, error) {
	// Pull Requestsを取得
	pullRequests, err := uc.GetPullRequests(ctx, req)
	if err != nil {
		return domain.PullRequestMetrics{}, err
	}
	
	// ドメインサービスでメトリクスを計算
	metrics := uc.service.CalculateMetrics(pullRequests)
	
	log.Printf("Calculated metrics for %d pull requests", metrics.TotalPullRequests)
	return metrics, nil
}

// GetAvailableDevelopers は利用可能な開発者一覧を取得
func (uc *UseCase) GetAvailableDevelopers(ctx context.Context) ([]string, error) {
	repos, err := uc.repo.GetRepositories(ctx)
	if err != nil {
		return nil, errors.NewRepositoryError(errors.ErrCodeExternalAPIError, "failed to get target repositories", err)
	}
	
	developers, err := uc.repo.GetDevelopers(ctx, repos)
	if err != nil {
		return nil, errors.NewRepositoryError(errors.ErrCodeExternalAPIError, "failed to get developers", err)
	}
	
	return developers, nil
}

// ValidateRequest はリクエストの高度なバリデーションを実行
func (uc *UseCase) ValidateRequest(ctx context.Context, req domain.GetPullRequestsRequest) error {
	// 基本バリデーション
	if err := req.Validate(); err != nil {
		return errors.NewValidationError(errors.ErrCodeInvalidRequest, "basic validation failed", err.Error())
	}
	
	// 開発者リストの検証（ドメインサービスを使用）
	if len(req.Developers) > 0 {
		availableDevelopers, err := uc.GetAvailableDevelopers(ctx)
		if err != nil {
			return errors.NewRepositoryError(errors.ErrCodeExternalAPIError, "failed to get available developers for validation", err)
		}
		
		if err := uc.service.ValidateDeveloperList(ctx, req.Developers, availableDevelopers); err != nil {
			return errors.WrapDomainError(err)
		}
	}
	
	return nil
}