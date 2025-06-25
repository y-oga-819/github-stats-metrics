package pull_request

import (
	"fmt"
	"sort"
	"time"
)

// ReviewTimeAnalyzer はレビュー時間を分析するサービス
type ReviewTimeAnalyzer struct {
	config ReviewTimeConfig
}

// ReviewTimeConfig はレビュー時間分析の設定
type ReviewTimeConfig struct {
	// レビューラウンドの判定設定
	MinTimeBetweenRounds time.Duration // レビューラウンド間の最小時間
	
	// レビュー品質の判定閾値
	HighQualityCommentThreshold int     // 高品質PRのコメント数閾値
	LowQualityCommentThreshold  int     // 低品質PRのコメント数閾値
	GoodFirstPassRate           float64 // 良好な初回通過率
	
	// レビュアー分析の設定
	MinReviewsForAnalysis int // 分析に必要な最小レビュー数
}

// NewReviewTimeAnalyzer は新しいレビュー時間分析器を作成
func NewReviewTimeAnalyzer() *ReviewTimeAnalyzer {
	return &ReviewTimeAnalyzer{
		config: getDefaultReviewTimeConfig(),
	}
}

// getDefaultReviewTimeConfig はデフォルトの設定を返す
func getDefaultReviewTimeConfig() ReviewTimeConfig {
	return ReviewTimeConfig{
		MinTimeBetweenRounds:        2 * time.Hour,
		HighQualityCommentThreshold: 3,
		LowQualityCommentThreshold:  10,
		GoodFirstPassRate:          0.8,
		MinReviewsForAnalysis:      3,
	}
}

// CalculateQualityMetrics は品質メトリクスを計算
func (analyzer *ReviewTimeAnalyzer) CalculateQualityMetrics(pr PullRequest, reviewEvents []ReviewEvent) PRQualityMetrics {
	qualityMetrics := PRQualityMetrics{}
	
	// レビューイベントを時系列でソート
	sortedEvents := make([]ReviewEvent, len(reviewEvents))
	copy(sortedEvents, reviewEvents)
	sort.Slice(sortedEvents, func(i, j int) bool {
		return sortedEvents[i].CreatedAt.Before(sortedEvents[j].CreatedAt)
	})
	
	// 基本的なカウント
	qualityMetrics.ReviewCommentCount = analyzer.countReviewComments(sortedEvents)
	qualityMetrics.ReviewRoundCount = analyzer.calculateReviewRounds(sortedEvents)
	qualityMetrics.CommitCount = 1 // 基本値（実際の実装では詳細なコミット情報から取得）
	
	// レビュアー分析
	reviewers, approvers := analyzer.analyzeReviewers(sortedEvents)
	qualityMetrics.ReviewerCount = len(reviewers)
	qualityMetrics.ApprovalsReceived = len(approvers)
	qualityMetrics.ReviewersInvolved = reviewers
	qualityMetrics.ApproversInvolved = approvers
	
	// レビュー効率の計算
	qualityMetrics.FirstReviewPassRate = analyzer.calculateFirstReviewPassRate(sortedEvents)
	qualityMetrics.AverageCommentPerFile = analyzer.calculateAverageCommentPerFile(qualityMetrics.ReviewCommentCount, pr)
	
	// 修正関連の分析（推定）
	qualityMetrics.FixupCommitCount = analyzer.estimateFixupCommits(sortedEvents)
	qualityMetrics.ForceUpdateCount = 0 // 実装時にGitHub APIから取得
	
	return qualityMetrics
}

// AnalyzeReviewEfficiency はレビュー効率を分析
func (analyzer *ReviewTimeAnalyzer) AnalyzeReviewEfficiency(metrics []*PRMetrics) *ReviewEfficiencyAnalysis {
	if len(metrics) == 0 {
		return &ReviewEfficiencyAnalysis{}
	}
	
	analysis := &ReviewEfficiencyAnalysis{
		TotalPRs: len(metrics),
	}
	
	// レビュアー別の分析
	reviewerStats := make(map[string]*ReviewerStatistics)
	
	var allCommentCounts []int
	var allRoundCounts []int
	var allFirstPassRates []float64
	
	for _, metric := range metrics {
		// 全体統計
		allCommentCounts = append(allCommentCounts, metric.QualityMetrics.ReviewCommentCount)
		allRoundCounts = append(allRoundCounts, metric.QualityMetrics.ReviewRoundCount)
		allFirstPassRates = append(allFirstPassRates, metric.QualityMetrics.FirstReviewPassRate)
		
		// レビュアー別統計
		for _, reviewer := range metric.QualityMetrics.ReviewersInvolved {
			if _, exists := reviewerStats[reviewer]; !exists {
				reviewerStats[reviewer] = &ReviewerStatistics{
					ReviewerName: reviewer,
				}
			}
			reviewerStats[reviewer].TotalReviews++
			reviewerStats[reviewer].TotalComments += metric.QualityMetrics.ReviewCommentCount
		}
		
		for _, approver := range metric.QualityMetrics.ApproversInvolved {
			if stats, exists := reviewerStats[approver]; exists {
				stats.TotalApprovals++
			}
		}
	}
	
	// 全体統計の計算
	analysis.OverallStats = analyzer.calculateOverallStats(allCommentCounts, allRoundCounts, allFirstPassRates)
	
	// レビュアー統計の最終化
	for _, stats := range reviewerStats {
		if stats.TotalReviews > 0 {
			stats.AverageCommentsPerReview = float64(stats.TotalComments) / float64(stats.TotalReviews)
		}
		if stats.TotalReviews >= analyzer.config.MinReviewsForAnalysis {
			analysis.ReviewerStats = append(analysis.ReviewerStats, *stats)
		}
	}
	
	// レビュアー統計をソート（レビュー数順）
	sort.Slice(analysis.ReviewerStats, func(i, j int) bool {
		return analysis.ReviewerStats[i].TotalReviews > analysis.ReviewerStats[j].TotalReviews
	})
	
	return analysis
}

// countReviewComments はレビューコメント数をカウント
func (analyzer *ReviewTimeAnalyzer) countReviewComments(events []ReviewEvent) int {
	count := 0
	for _, event := range events {
		if event.Type == ReviewEventTypeCommented || event.Type == ReviewEventTypeChangesRequested {
			count++
		}
	}
	return count
}

// calculateReviewRounds はレビューラウンド数を計算
func (analyzer *ReviewTimeAnalyzer) calculateReviewRounds(events []ReviewEvent) int {
	rounds := 0
	lastRoundTime := time.Time{}
	
	for _, event := range events {
		if event.Type == ReviewEventTypeChangesRequested {
			// 前回のラウンドから十分時間が経過している場合、新しいラウンドとする
			if event.CreatedAt.Sub(lastRoundTime) > analyzer.config.MinTimeBetweenRounds {
				rounds++
				lastRoundTime = event.CreatedAt
			}
		}
	}
	
	// 最低1ラウンド
	if rounds == 0 && len(events) > 0 {
		rounds = 1
	}
	
	return rounds
}

// analyzeReviewers はレビュアーを分析
func (analyzer *ReviewTimeAnalyzer) analyzeReviewers(events []ReviewEvent) ([]string, []string) {
	reviewers := make(map[string]bool)
	approvers := make(map[string]bool)
	
	for _, event := range events {
		switch event.Type {
		case ReviewEventTypeCommented, ReviewEventTypeChangesRequested:
			if event.Actor != "" {
				reviewers[event.Actor] = true
			}
		case ReviewEventTypeApproved:
			if event.Actor != "" {
				reviewers[event.Actor] = true
				approvers[event.Actor] = true
			}
		}
	}
	
	// マップからスライスに変換
	reviewerList := make([]string, 0, len(reviewers))
	for reviewer := range reviewers {
		reviewerList = append(reviewerList, reviewer)
	}
	
	approverList := make([]string, 0, len(approvers))
	for approver := range approvers {
		approverList = append(approverList, approver)
	}
	
	return reviewerList, approverList
}

// calculateFirstReviewPassRate は初回レビュー通過率を計算
func (analyzer *ReviewTimeAnalyzer) calculateFirstReviewPassRate(events []ReviewEvent) float64 {
	if len(events) == 0 {
		return 1.0
	}
	
	// 最初のレビューイベントをチェック
	for _, event := range events {
		if event.Type == ReviewEventTypeApproved {
			return 1.0 // 初回で承認
		} else if event.Type == ReviewEventTypeChangesRequested {
			return 0.0 // 初回で変更要求
		} else if event.Type == ReviewEventTypeCommented {
			// コメントの内容によって判定（簡易版）
			return 0.5 // 中程度
		}
	}
	
	return 1.0
}

// calculateAverageCommentPerFile はファイルあたりの平均コメント数を計算
func (analyzer *ReviewTimeAnalyzer) calculateAverageCommentPerFile(commentCount int, pr PullRequest) float64 {
	// ファイル数の情報がない場合は、変更行数から推定
	estimatedFiles := (pr.Additions + pr.Deletions) / 50 // 50行/ファイルと仮定
	if estimatedFiles < 1 {
		estimatedFiles = 1
	}
	
	return float64(commentCount) / float64(estimatedFiles)
}

// estimateFixupCommits は修正コミット数を推定
func (analyzer *ReviewTimeAnalyzer) estimateFixupCommits(events []ReviewEvent) int {
	fixupCount := 0
	
	for _, event := range events {
		if event.Type == ReviewEventTypeChangesRequested {
			fixupCount++ // 変更要求後の修正コミットを推定
		}
	}
	
	return fixupCount
}

// calculateOverallStats は全体統計を計算
func (analyzer *ReviewTimeAnalyzer) calculateOverallStats(commentCounts, roundCounts []int, firstPassRates []float64) OverallReviewStats {
	stats := OverallReviewStats{}
	
	if len(commentCounts) > 0 {
		// コメント数統計
		sort.Ints(commentCounts)
		stats.CommentStats = analyzer.calculateIntStatistics(commentCounts)
	}
	
	if len(roundCounts) > 0 {
		// ラウンド数統計
		sort.Ints(roundCounts)
		stats.RoundStats = analyzer.calculateIntStatistics(roundCounts)
	}
	
	if len(firstPassRates) > 0 {
		// 初回通過率統計
		sort.Float64s(firstPassRates)
		total := 0.0
		for _, rate := range firstPassRates {
			total += rate
		}
		stats.AverageFirstPassRate = total / float64(len(firstPassRates))
		stats.MedianFirstPassRate = firstPassRates[len(firstPassRates)/2]
	}
	
	return stats
}

// calculateIntStatistics は整数配列の統計を計算
func (analyzer *ReviewTimeAnalyzer) calculateIntStatistics(values []int) IntStatistics {
	if len(values) == 0 {
		return IntStatistics{}
	}
	
	stats := IntStatistics{
		Count: len(values),
		Min:   values[0],
		Max:   values[len(values)-1],
	}
	
	// 平均値
	total := 0
	for _, v := range values {
		total += v
	}
	stats.Average = float64(total) / float64(len(values))
	
	// 中央値
	if len(values)%2 == 0 {
		mid := len(values) / 2
		stats.Median = float64(values[mid-1]+values[mid]) / 2
	} else {
		stats.Median = float64(values[len(values)/2])
	}
	
	// パーセンタイル
	stats.P75 = values[int(float64(len(values))*0.75)]
	stats.P90 = values[int(float64(len(values))*0.90)]
	stats.P95 = values[int(float64(len(values))*0.95)]
	
	return stats
}

// AnalyzeReviewBottlenecks はレビューボトルネックを分析
func (analyzer *ReviewTimeAnalyzer) AnalyzeReviewBottlenecks(metrics []*PRMetrics) []ReviewBottleneck {
	var bottlenecks []ReviewBottleneck
	
	// 長時間レビュー待ちのPRを特定
	for _, metric := range metrics {
		if metric.TimeMetrics.TimeToFirstReview != nil {
			reviewTime := *metric.TimeMetrics.TimeToFirstReview
			if reviewTime > 24*time.Hour {
				bottlenecks = append(bottlenecks, ReviewBottleneck{
					Type:        BottleneckTypeSlowReview,
					PRID:        metric.PRID,
					Severity:    analyzer.getBottleneckSeverity(reviewTime),
					Description: fmt.Sprintf("レビュー開始まで%s", formatDuration(reviewTime)),
					Impact:      "開発速度の低下",
					Suggestion:  "レビュアーの早期アサインとレビュー優先度の設定",
				})
			}
		}
		
		// 複数ラウンドのレビューが発生しているPRを特定
		if metric.QualityMetrics.ReviewRoundCount > 3 {
			bottlenecks = append(bottlenecks, ReviewBottleneck{
				Type:        BottleneckTypeMultipleRounds,
				PRID:        metric.PRID,
				Severity:    IssueseverityMedium,
				Description: fmt.Sprintf("%d回のレビューラウンドが発生", metric.QualityMetrics.ReviewRoundCount),
				Impact:      "レビュアーとエンジニアの時間消費",
				Suggestion:  "事前のセルフレビューとコード品質の向上",
			})
		}
		
		// コメント数が多いPRを特定
		if metric.QualityMetrics.ReviewCommentCount > analyzer.config.LowQualityCommentThreshold {
			bottlenecks = append(bottlenecks, ReviewBottleneck{
				Type:        BottleneckTypeManyComments,
				PRID:        metric.PRID,
				Severity:    IssueseverityMedium,
				Description: fmt.Sprintf("%d件のレビューコメント", metric.QualityMetrics.ReviewCommentCount),
				Impact:      "レビューとフィードバック対応の長期化",
				Suggestion:  "PRサイズの縮小とコード品質の事前チェック",
			})
		}
	}
	
	return bottlenecks
}

// getBottleneckSeverity はボトルネックの重要度を判定
func (analyzer *ReviewTimeAnalyzer) getBottleneckSeverity(duration time.Duration) IssueSeverity {
	if duration > 72*time.Hour {
		return IssueseverityHigh
	} else if duration > 24*time.Hour {
		return IssueseverityMedium
	} else {
		return IssueseverityLow
	}
}

// 構造体定義

// ReviewEfficiencyAnalysis はレビュー効率分析の結果
type ReviewEfficiencyAnalysis struct {
	TotalPRs      int                   `json:"totalPRs"`
	OverallStats  OverallReviewStats    `json:"overallStats"`
	ReviewerStats []ReviewerStatistics  `json:"reviewerStats"`
}

// OverallReviewStats は全体のレビュー統計
type OverallReviewStats struct {
	CommentStats          IntStatistics `json:"commentStats"`
	RoundStats            IntStatistics `json:"roundStats"`
	AverageFirstPassRate  float64       `json:"averageFirstPassRate"`
	MedianFirstPassRate   float64       `json:"medianFirstPassRate"`
}

// IntStatistics は整数の統計情報
type IntStatistics struct {
	Count   int     `json:"count"`
	Average float64 `json:"average"`
	Median  float64 `json:"median"`
	Min     int     `json:"min"`
	Max     int     `json:"max"`
	P75     int     `json:"p75"`
	P90     int     `json:"p90"`
	P95     int     `json:"p95"`
}

// ReviewerStatistics はレビュアーの統計情報
type ReviewerStatistics struct {
	ReviewerName            string  `json:"reviewerName"`
	TotalReviews            int     `json:"totalReviews"`
	TotalApprovals          int     `json:"totalApprovals"`
	TotalComments           int     `json:"totalComments"`
	AverageCommentsPerReview float64 `json:"averageCommentsPerReview"`
}

// ReviewBottleneck はレビューボトルネック情報
type ReviewBottleneck struct {
	Type        BottleneckType `json:"type"`
	PRID        string         `json:"prId"`
	Severity    IssueSeverity  `json:"severity"`
	Description string         `json:"description"`
	Impact      string         `json:"impact"`
	Suggestion  string         `json:"suggestion"`
}

// BottleneckType はボトルネックのタイプ
type BottleneckType string

const (
	BottleneckTypeSlowReview     BottleneckType = "slow_review"
	BottleneckTypeMultipleRounds BottleneckType = "multiple_rounds"
	BottleneckTypeManyComments   BottleneckType = "many_comments"
	BottleneckTypeLackOfReviewer BottleneckType = "lack_of_reviewer"
)

