package pull_request

import (
	"context"
	"strings"
)

// PullRequestService はPull Requestのドメインサービス
// 複数のエンティティにまたがるビジネスルールやドメインロジックを処理
type PullRequestService struct{}

// NewPullRequestService はPullRequestServiceのコンストラクタ
func NewPullRequestService() *PullRequestService {
	return &PullRequestService{}
}

// FilterByBusinessRules はビジネスルールに基づいてPull Requestsをフィルタリング
func (s *PullRequestService) FilterByBusinessRules(pullRequests []PullRequest) []PullRequest {
	var filtered []PullRequest
	
	for _, pr := range pullRequests {
		if s.isValidForAnalysis(pr) {
			filtered = append(filtered, pr)
		}
	}
	
	return filtered
}

// isValidForAnalysis はPRが分析対象として有効かを判定
func (s *PullRequestService) isValidForAnalysis(pr PullRequest) bool {
	// ルール1: マージされていないPRは除外
	if !pr.IsMerged() {
		return false
	}
	
	// ルール2: epicブランチは除外（組織のワークフロー特有ルール）
	if s.isEpicBranch(pr.HeadRefName) {
		return false
	}
	
	// ルール3: レビュープロセスが完了していないPRは除外
	if !s.hasCompletedReviewProcess(pr) {
		return false
	}
	
	return true
}

// isEpicBranch はepicブランチかを判定
func (s *PullRequestService) isEpicBranch(branchName string) bool {
	return strings.HasPrefix(branchName, "epic/")
}

// hasCompletedReviewProcess はレビュープロセスが完了しているかを判定
func (s *PullRequestService) hasCompletedReviewProcess(pr PullRequest) bool {
	// 最低1回のレビューが必要（FirstReviewedが存在すること）
	return pr.IsReviewed()
}

// CalculateMetrics はPull Requestsのメトリクスを計算
func (s *PullRequestService) CalculateMetrics(pullRequests []PullRequest) PullRequestMetrics {
	if len(pullRequests) == 0 {
		return PullRequestMetrics{}
	}
	
	var totalReviewTime, totalApprovalTime, totalMergeTime int64
	var validReviewCount, validApprovalCount, validMergeCount int
	
	for _, pr := range pullRequests {
		// レビューまでの時間
		if reviewTime := s.calculateTimeToFirstReview(pr); reviewTime > 0 {
			totalReviewTime += reviewTime
			validReviewCount++
		}
		
		// 承認までの時間
		if approvalTime := s.calculateTimeToApproval(pr); approvalTime > 0 {
			totalApprovalTime += approvalTime
			validApprovalCount++
		}
		
		// マージまでの時間
		if mergeTime := s.calculateTimeToMerge(pr); mergeTime > 0 {
			totalMergeTime += mergeTime
			validMergeCount++
		}
	}
	
	return PullRequestMetrics{
		AverageTimeToFirstReview: s.safeAverage(totalReviewTime, validReviewCount),
		AverageTimeToApproval:    s.safeAverage(totalApprovalTime, validApprovalCount),
		AverageTimeToMerge:       s.safeAverage(totalMergeTime, validMergeCount),
		TotalPullRequests:        len(pullRequests),
	}
}

// calculateTimeToFirstReview は最初のレビューまでの時間を計算（秒）
func (s *PullRequestService) calculateTimeToFirstReview(pr PullRequest) int64 {
	if pr.FirstReviewed == nil {
		return 0
	}
	
	return pr.FirstReviewed.Unix() - pr.CreatedAt.Unix()
}

// calculateTimeToApproval は承認までの時間を計算（秒）
func (s *PullRequestService) calculateTimeToApproval(pr PullRequest) int64 {
	if pr.LastApproved == nil {
		return 0
	}
	
	return pr.LastApproved.Unix() - pr.CreatedAt.Unix()
}

// calculateTimeToMerge はマージまでの時間を計算（秒）
func (s *PullRequestService) calculateTimeToMerge(pr PullRequest) int64 {
	if pr.MergedAt == nil {
		return 0
	}
	
	return pr.MergedAt.Unix() - pr.CreatedAt.Unix()
}

// safeAverage は安全な平均値計算
func (s *PullRequestService) safeAverage(total int64, count int) float64 {
	if count == 0 {
		return 0
	}
	return float64(total) / float64(count)
}

// ValidateDeveloperList は開発者リストの妥当性を検証
func (s *PullRequestService) ValidateDeveloperList(ctx context.Context, developers []string, availableDevelopers []string) error {
	if len(developers) == 0 {
		return &DomainError{
			Type:    "VALIDATION_ERROR",
			Message: "developer list cannot be empty",
		}
	}
	
	// 使用可能な開発者のマップ作成
	availableMap := make(map[string]bool)
	for _, dev := range availableDevelopers {
		availableMap[dev] = true
	}
	
	// 不正な開発者名をチェック
	var invalidDevelopers []string
	for _, dev := range developers {
		if !availableMap[dev] {
			invalidDevelopers = append(invalidDevelopers, dev)
		}
	}
	
	if len(invalidDevelopers) > 0 {
		return &DomainError{
			Type:    "VALIDATION_ERROR",
			Message: "invalid developers specified",
			Details: &ValidationDetails{InvalidDevelopers: invalidDevelopers},
		}
	}
	
	return nil
}