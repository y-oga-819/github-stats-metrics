package github_api

import (
	"path/filepath"
	"sort"
	"strings"
	"time"

	prDomain "github-stats-metrics/domain/pull_request"
	"github.com/shurcooL/githubv4"
)

// convertExtendedToDomain は拡張されたGitHub APIレスポンスをDomainモデルに変換
func convertExtendedToDomain(apiPR ExtendedPullRequest) prDomain.PullRequest {
	pr := prDomain.PullRequest{
		ID:          string(apiPR.Id),
		Number:      int(apiPR.Number),
		Title:       string(apiPR.Title),
		BaseRefName: string(apiPR.BaseRefName),
		HeadRefName: string(apiPR.HeadRefName),
		Author: prDomain.Author{
			Login:     string(apiPR.Author.Login),
			AvatarURL: apiPR.Author.AvatarURL.String(),
		},
		Repository: prDomain.RepositoryInfo{
			Name: string(apiPR.Repository.Name),
		},
		URL:       apiPR.URL.String(),
		Additions: int(apiPR.Additions),
		Deletions: int(apiPR.Deletions),
		CreatedAt: apiPR.CreatedAt.Time,
	}
	
	// マージ時刻
	if !apiPR.MergedAt.Time.IsZero() {
		pr.MergedAt = &apiPR.MergedAt.Time
	}
	
	// レビュー情報から最初のレビューと最後の承認を抽出
	if len(apiPR.Reviews.Nodes) > 0 {
		// レビューを時系列でソート
		reviews := apiPR.Reviews.Nodes
		sort.Slice(reviews, func(i, j int) bool {
			return reviews[i].CreatedAt.Time.Before(reviews[j].CreatedAt.Time)
		})
		
		// 最初のレビュー
		pr.FirstReviewed = &reviews[0].CreatedAt.Time
		
		// 最後の承認
		for i := len(reviews) - 1; i >= 0; i-- {
			if reviews[i].State == githubv4.PullRequestReviewStateApproved {
				pr.LastApproved = &reviews[i].CreatedAt.Time
				break
			}
		}
	}
	
	return pr
}

// convertToPRMetrics はGitHub APIレスポンスをPRMetricsに変換
func convertToPRMetrics(apiPR ExtendedPullRequest) *prDomain.PRMetrics {
	// 基本情報
	metrics := &prDomain.PRMetrics{
		PRID:       string(apiPR.Id),
		PRNumber:   int(apiPR.Number),
		Title:      string(apiPR.Title),
		Author:     string(apiPR.Author.Login),
		Repository: string(apiPR.Repository.Name),
		CreatedAt:  apiPR.CreatedAt.Time,
	}
	
	if !apiPR.MergedAt.Time.IsZero() {
		metrics.MergedAt = &apiPR.MergedAt.Time
	}
	
	// サイズメトリクス
	metrics.SizeMetrics = calculateSizeMetrics(apiPR)
	
	// 時間メトリクス
	metrics.TimeMetrics = calculateTimeMetrics(apiPR)
	
	// 品質メトリクス
	metrics.QualityMetrics = calculateQualityMetrics(apiPR)
	
	// サイズカテゴリ
	metrics.SizeCategory = metrics.CalculateSizeCategory()
	
	return metrics
}

// calculateSizeMetrics はサイズ関連メトリクスを計算
func calculateSizeMetrics(apiPR ExtendedPullRequest) prDomain.PRSizeMetrics {
	sizeMetrics := prDomain.PRSizeMetrics{
		LinesAdded:   int(apiPR.Additions),
		LinesDeleted: int(apiPR.Deletions),
		LinesChanged: int(apiPR.Additions) + int(apiPR.Deletions),
		FilesChanged: int(apiPR.ChangedFiles),
	}
	
	// ファイル変更詳細
	fileChanges := make([]prDomain.FileChangeMetrics, 0, len(apiPR.Files.Nodes))
	fileTypeBreakdown := make(map[string]int)
	directories := make(map[string]bool)
	
	for _, file := range apiPR.Files.Nodes {
		fileName := string(file.Path)
		fileType := getFileExtension(fileName)
		
		fileChange := prDomain.FileChangeMetrics{
			FileName:     fileName,
			FileType:     fileType,
			LinesAdded:   int(file.Additions),
			LinesDeleted: int(file.Deletions),
			IsNewFile:    string(file.ChangeType) == "ADDED",
			IsDeleted:    string(file.ChangeType) == "DELETED",
			IsRenamed:    string(file.ChangeType) == "RENAMED",
		}
		
		fileChanges = append(fileChanges, fileChange)
		fileTypeBreakdown[fileType]++
		directories[filepath.Dir(fileName)] = true
	}
	
	sizeMetrics.FileChanges = fileChanges
	sizeMetrics.FileTypeBreakdown = fileTypeBreakdown
	sizeMetrics.DirectoryCount = len(directories)
	
	return sizeMetrics
}

// calculateTimeMetrics は時間関連メトリクスを計算
func calculateTimeMetrics(apiPR ExtendedPullRequest) prDomain.PRTimeMetrics {
	timeMetrics := prDomain.PRTimeMetrics{
		CreatedHour: apiPR.CreatedAt.Time.Hour(),
	}
	
	if !apiPR.MergedAt.Time.IsZero() {
		mergedHour := apiPR.MergedAt.Time.Hour()
		timeMetrics.MergedHour = &mergedHour
		
		// 全体サイクルタイム
		totalCycle := apiPR.MergedAt.Time.Sub(apiPR.CreatedAt.Time)
		timeMetrics.TotalCycleTime = &totalCycle
	}
	
	// レビュー関連時間の計算
	if len(apiPR.Reviews.Nodes) > 0 {
		reviews := apiPR.Reviews.Nodes
		sort.Slice(reviews, func(i, j int) bool {
			return reviews[i].CreatedAt.Time.Before(reviews[j].CreatedAt.Time)
		})
		
		// 最初のレビューまでの時間
		firstReviewTime := reviews[0].CreatedAt.Time.Sub(apiPR.CreatedAt.Time)
		timeMetrics.TimeToFirstReview = &firstReviewTime
		
		// 最後の承認までの時間
		for i := len(reviews) - 1; i >= 0; i-- {
			if reviews[i].State == githubv4.PullRequestReviewStateApproved {
				approvalTime := reviews[i].CreatedAt.Time.Sub(apiPR.CreatedAt.Time)
				timeMetrics.TimeToApproval = &approvalTime
				
				// 承認からマージまでの時間
				if !apiPR.MergedAt.Time.IsZero() {
					mergeTime := apiPR.MergedAt.Time.Sub(reviews[i].CreatedAt.Time)
					timeMetrics.TimeToMerge = &mergeTime
				}
				break
			}
		}
	}
	
	return timeMetrics
}

// calculateQualityMetrics は品質関連メトリクスを計算
func calculateQualityMetrics(apiPR ExtendedPullRequest) prDomain.PRQualityMetrics {
	qualityMetrics := prDomain.PRQualityMetrics{
		ReviewCommentCount: int(apiPR.ReviewComments.TotalCount),
		CommitCount:        int(apiPR.Commits.TotalCount),
	}
	
	// レビュアー情報
	reviewers := make(map[string]bool)
	approvers := make(map[string]bool)
	reviewRounds := 0
	
	for _, review := range apiPR.Reviews.Nodes {
		reviewer := string(review.Author.Login)
		reviewers[reviewer] = true
		
		if review.State == githubv4.PullRequestReviewStateApproved {
			approvers[reviewer] = true
		}
		
		// レビューラウンド数のカウント（簡易版）
		if review.State == githubv4.PullRequestReviewStateChangesRequested {
			reviewRounds++
		}
	}
	
	qualityMetrics.ReviewerCount = len(reviewers)
	qualityMetrics.ApprovalsReceived = len(approvers)
	qualityMetrics.ReviewRoundCount = reviewRounds
	
	// レビュアーとアプルーバーのリスト
	for reviewer := range reviewers {
		qualityMetrics.ReviewersInvolved = append(qualityMetrics.ReviewersInvolved, reviewer)
	}
	for approver := range approvers {
		qualityMetrics.ApproversInvolved = append(qualityMetrics.ApproversInvolved, approver)
	}
	
	// ファイルあたりの平均コメント数
	if int(apiPR.ChangedFiles) > 0 {
		qualityMetrics.AverageCommentPerFile = float64(qualityMetrics.ReviewCommentCount) / float64(apiPR.ChangedFiles)
	}
	
	// 初回レビュー通過率（簡易計算）
	if reviewRounds == 0 && len(reviewers) > 0 {
		qualityMetrics.FirstReviewPassRate = 1.0
	} else if reviewRounds > 0 {
		qualityMetrics.FirstReviewPassRate = 1.0 / float64(reviewRounds+1)
	}
	
	// Fixupコミット数の推定
	for _, commit := range apiPR.Commits.Nodes {
		headline := strings.ToLower(string(commit.Commit.MessageHeadline))
		if strings.Contains(headline, "fix") || 
		   strings.Contains(headline, "fixup") || 
		   strings.Contains(headline, "修正") {
			qualityMetrics.FixupCommitCount++
		}
	}
	
	return qualityMetrics
}

// getFileExtension はファイルパスから拡張子を取得
func getFileExtension(fileName string) string {
	ext := filepath.Ext(fileName)
	if ext == "" {
		return "none"
	}
	return ext
}

// convertReviewState はGitHubのレビュー状態をドメインイベントタイプに変換
func convertReviewState(state githubv4.PullRequestReviewState) prDomain.ReviewEventType {
	switch state {
	case githubv4.PullRequestReviewStateApproved:
		return prDomain.ReviewEventTypeApproved
	case githubv4.PullRequestReviewStateChangesRequested:
		return prDomain.ReviewEventTypeChangesRequested
	case githubv4.PullRequestReviewStateCommented:
		return prDomain.ReviewEventTypeCommented
	case githubv4.PullRequestReviewStateDismissed:
		return prDomain.ReviewEventTypeDismissed
	default:
		return prDomain.ReviewEventTypeCommented
	}
}

// ReviewEvent とそのタイプをドメインに追加する必要があります
// ドメインモデルに以下を追加:

// ReviewEvent はレビューイベントの情報
type ReviewEvent struct {
	Type      ReviewEventType `json:"type"`
	CreatedAt time.Time       `json:"createdAt"`
	Actor     string          `json:"actor"`
	Reviewer  string          `json:"reviewer,omitempty"`
}

// ReviewEventType はレビューイベントのタイプ
type ReviewEventType string

const (
	ReviewEventTypeRequested         ReviewEventType = "requested"
	ReviewEventTypeApproved          ReviewEventType = "approved"
	ReviewEventTypeChangesRequested ReviewEventType = "changes_requested"
	ReviewEventTypeCommented         ReviewEventType = "commented"
	ReviewEventTypeDismissed         ReviewEventType = "dismissed"
	ReviewEventTypeReadyForReview    ReviewEventType = "ready_for_review"
	ReviewEventTypeMerged            ReviewEventType = "merged"
)