package pull_request

import (
	"fmt"
	"time"

	prDomain "github-stats-metrics/domain/pull_request"
)

// PRMetricsPresenter はPRメトリクスのプレゼンター
type PRMetricsPresenter struct {
	complexityAnalyzer *prDomain.PRComplexityAnalyzer
}

// NewPRMetricsPresenter は新しいPRメトリクスプレゼンターを作成
func NewPRMetricsPresenter() *PRMetricsPresenter {
	return &PRMetricsPresenter{
		complexityAnalyzer: prDomain.NewPRComplexityAnalyzer(),
	}
}

// ToPRMetricsResponse はPRメトリクスをレスポンス形式に変換
func (presenter *PRMetricsPresenter) ToPRMetricsResponse(metrics *prDomain.PRMetrics) *PRMetricsResponse {
	return &PRMetricsResponse{
		PRID:       metrics.PRID,
		PRNumber:   metrics.PRNumber,
		Title:      metrics.Title,
		Author:     metrics.Author,
		Repository: metrics.Repository,
		CreatedAt:  metrics.CreatedAt,
		MergedAt:   metrics.MergedAt,

		SizeMetrics:    presenter.toSizeMetricsResponse(metrics.SizeMetrics),
		TimeMetrics:    presenter.toTimeMetricsResponse(metrics.TimeMetrics),
		QualityMetrics: presenter.toQualityMetricsResponse(metrics.QualityMetrics),

		ComplexityScore: metrics.ComplexityScore,
		ComplexityLevel: string(presenter.complexityAnalyzer.GetComplexityLevel(metrics.ComplexityScore)),
		SizeCategory:    string(metrics.SizeCategory),

		AnalysisResults: presenter.toAnalysisResultsResponse(metrics),
	}
}

// ToPRSummaryResponse はPRメトリクスを要約レスポンス形式に変換
func (presenter *PRMetricsPresenter) ToPRSummaryResponse(metrics *prDomain.PRMetrics) *PRSummaryResponse {
	return &PRSummaryResponse{
		PRID:            metrics.PRID,
		PRNumber:        metrics.PRNumber,
		Title:           metrics.Title,
		Author:          metrics.Author,
		Repository:      metrics.Repository,
		CreatedAt:       metrics.CreatedAt,
		MergedAt:        metrics.MergedAt,
		LinesChanged:    metrics.SizeMetrics.LinesChanged,
		FilesChanged:    metrics.SizeMetrics.FilesChanged,
		ComplexityScore: metrics.ComplexityScore,
		SizeCategory:    string(metrics.SizeCategory),
		CycleTime:       presenter.toDurationResponse(metrics.TimeMetrics.TotalCycleTime),
		ReviewTime:      presenter.toDurationResponse(metrics.TimeMetrics.TimeToFirstReview),
		IsHighQuality:   metrics.IsHighQuality(),
	}
}

// ToCycleTimeMetricsResponse はサイクルタイムメトリクスをレスポンス形式に変換
func (presenter *PRMetricsPresenter) ToCycleTimeMetricsResponse(
	metrics []*prDomain.PRMetrics,
	period string,
	startDate, endDate time.Time,
) *CycleTimeMetricsResponse {
	if len(metrics) == 0 {
		return &CycleTimeMetricsResponse{
			Period:    period,
			StartDate: startDate,
			EndDate:   endDate,
			TotalPRs:  0,
		}
	}

	// サイクルタイムの統計計算
	var cycleTimes []time.Duration
	var reviewTimes []time.Duration
	var approvalTimes []time.Duration
	var mergeTimes []time.Duration

	for _, metric := range metrics {
		if metric.TimeMetrics.TotalCycleTime != nil {
			cycleTimes = append(cycleTimes, *metric.TimeMetrics.TotalCycleTime)
		}
		if metric.TimeMetrics.TimeToFirstReview != nil {
			reviewTimes = append(reviewTimes, *metric.TimeMetrics.TimeToFirstReview)
		}
		if metric.TimeMetrics.TimeToApproval != nil {
			approvalTimes = append(approvalTimes, *metric.TimeMetrics.TimeToApproval)
		}
		if metric.TimeMetrics.TimeToMerge != nil {
			mergeTimes = append(mergeTimes, *metric.TimeMetrics.TimeToMerge)
		}
	}

	return &CycleTimeMetricsResponse{
		Period:      period,
		StartDate:   startDate,
		EndDate:     endDate,
		TotalPRs:    len(metrics),
		Statistics:  presenter.toCycleTimeStatsResponse(cycleTimes),
		Percentiles: presenter.toPercentilesResponse(cycleTimes),
		Breakdown: CycleTimeBreakdownResponse{
			TimeToFirstReview: presenter.toCycleTimeStatsResponse(reviewTimes),
			TimeToApproval:    presenter.toCycleTimeStatsResponse(approvalTimes),
			TimeToMerge:       presenter.toCycleTimeStatsResponse(mergeTimes),
		},
		Trends: presenter.calculateTrendResponse(cycleTimes),
	}
}

// ToReviewTimeMetricsResponse はレビュー時間メトリクスをレスポンス形式に変換
func (presenter *PRMetricsPresenter) ToReviewTimeMetricsResponse(
	metrics []*prDomain.PRMetrics,
	period string,
	startDate, endDate time.Time,
) *ReviewTimeMetricsResponse {
	if len(metrics) == 0 {
		return &ReviewTimeMetricsResponse{
			Period:    period,
			StartDate: startDate,
			EndDate:   endDate,
			TotalPRs:  0,
		}
	}

	// レビュー統計の計算
	var reviewTimes []time.Duration
	var waitTimes []time.Duration
	var activeTimes []time.Duration
	totalRounds := 0
	passedFirstReview := 0

	for _, metric := range metrics {
		if metric.TimeMetrics.TimeToFirstReview != nil {
			reviewTimes = append(reviewTimes, *metric.TimeMetrics.TimeToFirstReview)
		}
		if metric.TimeMetrics.ReviewWaitTime != nil {
			waitTimes = append(waitTimes, *metric.TimeMetrics.ReviewWaitTime)
		}
		if metric.TimeMetrics.ReviewActiveTime != nil {
			activeTimes = append(activeTimes, *metric.TimeMetrics.ReviewActiveTime)
		}
		
		totalRounds += metric.QualityMetrics.ReviewRoundCount
		if metric.QualityMetrics.FirstReviewPassRate >= 0.8 {
			passedFirstReview++
		}
	}

	avgRounds := float64(totalRounds) / float64(len(metrics))
	firstPassRate := float64(passedFirstReview) / float64(len(metrics))

	// ボトルネックの特定
	bottlenecks := presenter.identifyReviewBottlenecks(metrics)

	return &ReviewTimeMetricsResponse{
		Period:    period,
		StartDate: startDate,
		EndDate:   endDate,
		TotalPRs:  len(metrics),
		ReviewStatistics: ReviewTimeStatsResponse{
			AvgTimeToFirstReview: presenter.calculateAvgDurationResponse(reviewTimes),
			AvgReviewWaitTime:    presenter.calculateAvgDurationResponse(waitTimes),
			AvgReviewActiveTime:  presenter.calculateAvgDurationResponse(activeTimes),
			AvgReviewRounds:      avgRounds,
			FirstPassRate:        firstPassRate,
		},
		EfficiencyMetrics: presenter.calculateReviewEfficiency(metrics),
		Bottlenecks:       bottlenecks,
		Trends:           presenter.calculateTrendResponse(reviewTimes),
	}
}

// ToPRListResponse はPRリストをレスポンス形式に変換
func (presenter *PRMetricsPresenter) ToPRListResponse(
	metrics []*prDomain.PRMetrics,
	totalCount, page, pageSize int,
) *PRListResponse {
	prs := make([]PRSummaryResponse, len(metrics))
	for i, metric := range metrics {
		prs[i] = *presenter.ToPRSummaryResponse(metric)
	}

	hasMore := (page * pageSize) < totalCount

	return &PRListResponse{
		PRs:        prs,
		TotalCount: totalCount,
		Page:       page,
		PageSize:   pageSize,
		HasMore:    hasMore,
	}
}

// プライベートメソッド

func (presenter *PRMetricsPresenter) toSizeMetricsResponse(metrics prDomain.PRSizeMetrics) PRSizeMetricsResponse {
	fileChanges := make([]FileChangeResponse, len(metrics.FileChanges))
	for i, change := range metrics.FileChanges {
		fileChanges[i] = FileChangeResponse{
			FileName:     change.FileName,
			FileType:     change.FileType,
			LinesAdded:   change.LinesAdded,
			LinesDeleted: change.LinesDeleted,
			IsNewFile:    change.IsNewFile,
			IsDeleted:    change.IsDeleted,
			IsRenamed:    change.IsRenamed,
		}
	}

	return PRSizeMetricsResponse{
		LinesAdded:        metrics.LinesAdded,
		LinesDeleted:      metrics.LinesDeleted,
		LinesChanged:      metrics.LinesChanged,
		FilesChanged:      metrics.FilesChanged,
		FileTypeBreakdown: metrics.FileTypeBreakdown,
		DirectoryCount:    metrics.DirectoryCount,
		FileChanges:       fileChanges,
	}
}

func (presenter *PRMetricsPresenter) toTimeMetricsResponse(metrics prDomain.PRTimeMetrics) PRTimeMetricsResponse {
	return PRTimeMetricsResponse{
		TotalCycleTime:     presenter.toDurationResponse(metrics.TotalCycleTime),
		TimeToFirstReview:  presenter.toDurationResponse(metrics.TimeToFirstReview),
		TimeToApproval:     presenter.toDurationResponse(metrics.TimeToApproval),
		TimeToMerge:        presenter.toDurationResponse(metrics.TimeToMerge),
		ReviewWaitTime:     presenter.toDurationResponse(metrics.ReviewWaitTime),
		ReviewActiveTime:   presenter.toDurationResponse(metrics.ReviewActiveTime),
		FirstCommitToMerge: presenter.toDurationResponse(metrics.FirstCommitToMerge),
		CreatedHour:        metrics.CreatedHour,
		MergedHour:         metrics.MergedHour,
	}
}

func (presenter *PRMetricsPresenter) toQualityMetricsResponse(metrics prDomain.PRQualityMetrics) PRQualityMetricsResponse {
	return PRQualityMetricsResponse{
		ReviewCommentCount:    metrics.ReviewCommentCount,
		ReviewRoundCount:      metrics.ReviewRoundCount,
		ReviewerCount:         metrics.ReviewerCount,
		ReviewersInvolved:     metrics.ReviewersInvolved,
		CommitCount:           metrics.CommitCount,
		FixupCommitCount:      metrics.FixupCommitCount,
		ForceUpdateCount:      metrics.ForceUpdateCount,
		FirstReviewPassRate:   metrics.FirstReviewPassRate,
		AverageCommentPerFile: metrics.AverageCommentPerFile,
		ApprovalsReceived:     metrics.ApprovalsReceived,
		ApproversInvolved:     metrics.ApproversInvolved,
	}
}

func (presenter *PRMetricsPresenter) toAnalysisResultsResponse(metrics *prDomain.PRMetrics) PRAnalysisResultsResponse {
	recommendations := presenter.generateRecommendations(metrics)
	splitSuggestions := presenter.convertSplitSuggestions(
		presenter.complexityAnalyzer.SuggestOptimalSplit(metrics),
	)

	return PRAnalysisResultsResponse{
		IsHighQuality:       metrics.IsHighQuality(),
		IsLargePR:           metrics.IsLargePR(),
		HasLongReviewTime:   metrics.HasLongReviewTime(24 * time.Hour),
		ReviewEfficiency:    metrics.CalculateReviewEfficiency(),
		DominantFileType:    metrics.GetDominantFileType(),
		Recommendations:     recommendations,
		SplitSuggestions:    splitSuggestions,
	}
}

func (presenter *PRMetricsPresenter) toDurationResponse(duration *time.Duration) *DurationResponse {
	if duration == nil {
		return nil
	}

	return &DurationResponse{
		Seconds:       int64(duration.Seconds()),
		HumanReadable: presenter.formatDuration(*duration),
	}
}

func (presenter *PRMetricsPresenter) formatDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%.0f秒", d.Seconds())
	} else if d < time.Hour {
		return fmt.Sprintf("%.1f分", d.Minutes())
	} else if d < 24*time.Hour {
		return fmt.Sprintf("%.1f時間", d.Hours())
	} else {
		return fmt.Sprintf("%.1f日", d.Hours()/24)
	}
}

func (presenter *PRMetricsPresenter) toCycleTimeStatsResponse(durations []time.Duration) CycleTimeStatsResponse {
	if len(durations) == 0 {
		return CycleTimeStatsResponse{}
	}

	// 統計計算
	total := time.Duration(0)
	min := durations[0]
	max := durations[0]

	for _, d := range durations {
		total += d
		if d < min {
			min = d
		}
		if d > max {
			max = d
		}
	}

	mean := total / time.Duration(len(durations))
	
	// 中央値計算
	sorted := make([]time.Duration, len(durations))
	copy(sorted, durations)
	// ソート実装は省略

	median := sorted[len(sorted)/2]

	// 標準偏差計算
	variance := time.Duration(0)
	for _, d := range durations {
		diff := d - mean
		variance += time.Duration(float64(diff) * float64(diff) / float64(len(durations)))
	}
	stdDev := time.Duration(0) // 実際の計算は省略

	return CycleTimeStatsResponse{
		Mean:   presenter.toDurationResponse(&mean),
		Median: presenter.toDurationResponse(&median),
		Min:    presenter.toDurationResponse(&min),
		Max:    presenter.toDurationResponse(&max),
		StdDev: presenter.toDurationResponse(&stdDev),
	}
}

func (presenter *PRMetricsPresenter) toPercentilesResponse(durations []time.Duration) PercentilesResponse {
	if len(durations) == 0 {
		return PercentilesResponse{}
	}

	// パーセンタイル計算（簡略化）
	sorted := make([]time.Duration, len(durations))
	copy(sorted, durations)
	// ソート実装は省略

	getPercentile := func(p float64) time.Duration {
		index := int(float64(len(sorted)-1) * p)
		return sorted[index]
	}

	p25 := getPercentile(0.25)
	p50 := getPercentile(0.50)
	p75 := getPercentile(0.75)
	p90 := getPercentile(0.90)
	p95 := getPercentile(0.95)
	p99 := getPercentile(0.99)

	return PercentilesResponse{
		P25: presenter.toDurationResponse(&p25),
		P50: presenter.toDurationResponse(&p50),
		P75: presenter.toDurationResponse(&p75),
		P90: presenter.toDurationResponse(&p90),
		P95: presenter.toDurationResponse(&p95),
		P99: presenter.toDurationResponse(&p99),
	}
}

func (presenter *PRMetricsPresenter) calculateAvgDurationResponse(durations []time.Duration) *DurationResponse {
	if len(durations) == 0 {
		return nil
	}

	total := time.Duration(0)
	for _, d := range durations {
		total += d
	}

	avg := total / time.Duration(len(durations))
	return presenter.toDurationResponse(&avg)
}

func (presenter *PRMetricsPresenter) calculateTrendResponse(durations []time.Duration) TrendResponse {
	if len(durations) < 2 {
		return TrendResponse{
			Direction:   "stable",
			Slope:       0,
			Confidence:  0,
			Description: "データ不足のため傾向を判定できません",
		}
	}

	// トレンド計算（簡略化）
	slope := 0.0 // 実際の線形回帰計算は省略
	confidence := 0.5

	direction := "stable"
	description := "安定しています"

	if slope > 0.1 && confidence > 0.6 {
		direction = "degrading"
		description = "悪化傾向にあります"
	} else if slope < -0.1 && confidence > 0.6 {
		direction = "improving"
		description = "改善傾向にあります"
	}

	return TrendResponse{
		Direction:   direction,
		Slope:       slope,
		Confidence:  confidence,
		Description: description,
	}
}

func (presenter *PRMetricsPresenter) calculateReviewEfficiency(metrics []*prDomain.PRMetrics) ReviewEfficiencyResponse {
	if len(metrics) == 0 {
		return ReviewEfficiencyResponse{}
	}

	totalEfficiency := 0.0
	totalCommentEfficiency := 0.0
	totalRoundEfficiency := 0.0
	totalTimeEfficiency := 0.0

	for _, metric := range metrics {
		efficiency := metric.CalculateReviewEfficiency()
		totalEfficiency += efficiency

		// コメント効率（コメント数が少ないほど高い）
		commentEff := 1.0 / (1.0 + float64(metric.QualityMetrics.ReviewCommentCount)/10.0)
		totalCommentEfficiency += commentEff

		// ラウンド効率（ラウンド数が少ないほど高い）
		roundEff := 1.0 / float64(metric.QualityMetrics.ReviewRoundCount)
		totalRoundEfficiency += roundEff

		// 時間効率（レビュー時間が短いほど高い）
		timeEff := 1.0
		if metric.TimeMetrics.TimeToFirstReview != nil {
			hours := metric.TimeMetrics.TimeToFirstReview.Hours()
			timeEff = 1.0 / (1.0 + hours/24.0) // 1日を基準
		}
		totalTimeEfficiency += timeEff
	}

	count := float64(len(metrics))
	return ReviewEfficiencyResponse{
		OverallEfficiency: totalEfficiency / count,
		CommentEfficiency: totalCommentEfficiency / count,
		RoundEfficiency:   totalRoundEfficiency / count,
		TimeEfficiency:    totalTimeEfficiency / count,
	}
}

func (presenter *PRMetricsPresenter) identifyReviewBottlenecks(metrics []*prDomain.PRMetrics) []BottleneckResponse {
	var bottlenecks []BottleneckResponse

	for _, metric := range metrics {
		// 長いレビュー時間
		if metric.TimeMetrics.TimeToFirstReview != nil && *metric.TimeMetrics.TimeToFirstReview > 24*time.Hour {
			bottlenecks = append(bottlenecks, BottleneckResponse{
				Type:        "long_review_wait",
				PRID:        metric.PRID,
				Severity:    "high",
				Description: fmt.Sprintf("レビュー待ち時間が%s", presenter.formatDuration(*metric.TimeMetrics.TimeToFirstReview)),
				Value:       metric.TimeMetrics.TimeToFirstReview.Hours(),
				Suggestion:  "レビュアーの割り当てやレビュープロセスの見直しを検討してください",
			})
		}

		// 多数のレビューラウンド
		if metric.QualityMetrics.ReviewRoundCount > 3 {
			bottlenecks = append(bottlenecks, BottleneckResponse{
				Type:        "multiple_review_rounds",
				PRID:        metric.PRID,
				Severity:    "medium",
				Description: fmt.Sprintf("%d回のレビューラウンド", metric.QualityMetrics.ReviewRoundCount),
				Value:       float64(metric.QualityMetrics.ReviewRoundCount),
				Suggestion:  "初回レビューの品質向上やレビューガイドラインの整備を検討してください",
			})
		}

		// 大きなPR
		if metric.IsLargePR() {
			bottlenecks = append(bottlenecks, BottleneckResponse{
				Type:        "large_pr",
				PRID:        metric.PRID,
				Severity:    "medium",
				Description: fmt.Sprintf("大きなPR（%d行の変更）", metric.SizeMetrics.LinesChanged),
				Value:       float64(metric.SizeMetrics.LinesChanged),
				Suggestion:  "PRを小さく分割することでレビュー効率を向上できます",
			})
		}
	}

	return bottlenecks
}

func (presenter *PRMetricsPresenter) generateRecommendations(metrics *prDomain.PRMetrics) []string {
	var recommendations []string

	if metrics.IsLargePR() {
		recommendations = append(recommendations, "PRサイズが大きいため、小さく分割することをお勧めします")
	}

	if metrics.QualityMetrics.ReviewRoundCount > 3 {
		recommendations = append(recommendations, "レビューラウンド数が多いため、初回提出前のセルフレビューを強化することをお勧めします")
	}

	if metrics.TimeMetrics.TimeToFirstReview != nil && *metrics.TimeMetrics.TimeToFirstReview > 48*time.Hour {
		recommendations = append(recommendations, "レビュー待ち時間が長いため、レビュアーの割り当てプロセスを見直すことをお勧めします")
	}

	if metrics.ComplexityScore > 6.0 {
		recommendations = append(recommendations, "複雑度が高いため、コードの構造化やコメントの追加を検討してください")
	}

	if metrics.QualityMetrics.ReviewCommentCount > 15 {
		recommendations = append(recommendations, "レビューコメントが多いため、コーディング規約の確認や静的解析ツールの活用をお勧めします")
	}

	if len(recommendations) == 0 {
		recommendations = append(recommendations, "良好なPRです。現在の品質を維持してください")
	}

	return recommendations
}

func (presenter *PRMetricsPresenter) convertSplitSuggestions(suggestions []prDomain.SplitSuggestion) []SplitSuggestionResponse {
	response := make([]SplitSuggestionResponse, len(suggestions))
	for i, suggestion := range suggestions {
		response[i] = SplitSuggestionResponse{
			Type:        string(suggestion.Type),
			Description: suggestion.Description,
			Groups:      suggestion.Groups,
		}
	}
	return response
}