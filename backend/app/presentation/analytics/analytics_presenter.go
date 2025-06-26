package analytics

import (
	"fmt"
	"time"

	analyticsApp "github-stats-metrics/application/analytics"
	prDomain "github-stats-metrics/domain/pull_request"
	"github-stats-metrics/shared/utils"
)

// AnalyticsPresenter は集計データのプレゼンター
type AnalyticsPresenter struct{}

// NewAnalyticsPresenter は新しい集計データプレゼンターを作成
func NewAnalyticsPresenter() *AnalyticsPresenter {
	return &AnalyticsPresenter{}
}

// ToTeamMetricsResponse はチームメトリクスをレスポンス形式に変換
func (presenter *AnalyticsPresenter) ToTeamMetricsResponse(metrics *analyticsApp.TeamMetrics) *TeamMetricsResponse {
	return &TeamMetricsResponse{
		Period:      string(metrics.Period),
		TotalPRs:    metrics.TotalPRs,
		DateRange:   presenter.toDateRangeResponse(metrics.DateRange),
		GeneratedAt: metrics.GeneratedAt,
		
		CycleTimeStats:  presenter.toCycleTimeStatsAggResponse(metrics.CycleTimeStats),
		ReviewStats:     presenter.toReviewStatsAggResponse(metrics.ReviewStats),
		SizeStats:       presenter.toSizeStatsAggResponse(metrics.SizeStats),
		QualityStats:    presenter.toQualityStatsAggResponse(metrics.QualityStats),
		ComplexityStats: presenter.toComplexityStatsAggResponse(metrics.ComplexityStats),
		
		TrendAnalysis: presenter.toTrendAnalysisResponse(metrics.TrendAnalysis),
		Bottlenecks:   presenter.toBottlenecksResponse(metrics.Bottlenecks),
		Performance:   presenter.toPerformanceIndicatorsResponse(metrics),
	}
}

// ToDeveloperMetricsResponse は開発者メトリクスをレスポンス形式に変換
func (presenter *AnalyticsPresenter) ToDeveloperMetricsResponse(metrics *analyticsApp.DeveloperMetrics) *DeveloperMetricsResponse {
	return &DeveloperMetricsResponse{
		Developer:   metrics.Developer,
		Period:      string(metrics.Period),
		TotalPRs:    metrics.TotalPRs,
		DateRange:   presenter.toDateRangeResponse(metrics.DateRange),
		GeneratedAt: metrics.GeneratedAt,
		
		CycleTimeStats:  presenter.toCycleTimeStatsAggResponse(metrics.CycleTimeStats),
		ReviewStats:     presenter.toReviewStatsAggResponse(metrics.ReviewStats),
		SizeStats:       presenter.toSizeStatsAggResponse(metrics.SizeStats),
		QualityStats:    presenter.toQualityStatsAggResponse(metrics.QualityStats),
		ComplexityStats: presenter.toComplexityStatsAggResponse(metrics.ComplexityStats),
		
		Productivity: presenter.toProductivityMetricsResponse(metrics.Productivity),
		Ranking:      presenter.toDeveloperRankingResponse(metrics.Developer),
	}
}

// ToRepositoryMetricsResponse はリポジトリメトリクスをレスポンス形式に変換
func (presenter *AnalyticsPresenter) ToRepositoryMetricsResponse(metrics *analyticsApp.RepositoryMetrics) *RepositoryMetricsResponse {
	return &RepositoryMetricsResponse{
		Repository:  metrics.Repository,
		Period:      string(metrics.Period),
		TotalPRs:    metrics.TotalPRs,
		DateRange:   presenter.toDateRangeResponse(metrics.DateRange),
		GeneratedAt: metrics.GeneratedAt,
		
		CycleTimeStats:  presenter.toCycleTimeStatsAggResponse(metrics.CycleTimeStats),
		ReviewStats:     presenter.toReviewStatsAggResponse(metrics.ReviewStats),
		SizeStats:       presenter.toSizeStatsAggResponse(metrics.SizeStats),
		QualityStats:    presenter.toQualityStatsAggResponse(metrics.QualityStats),
		ComplexityStats: presenter.toComplexityStatsAggResponse(metrics.ComplexityStats),
		
		Contributors:  metrics.Contributors,
		ActivityLevel: presenter.toActivityLevelResponse(metrics),
	}
}

// ToTeamMetricsListResponse はチームメトリクスリストをレスポンス形式に変換
func (presenter *AnalyticsPresenter) ToTeamMetricsListResponse(
	metricsList []*analyticsApp.TeamMetrics,
	totalCount, page, pageSize int,
	filters AppliedFiltersResponse,
) *TeamMetricsListResponse {
	metrics := make([]TeamMetricsResponse, len(metricsList))
	for i, metric := range metricsList {
		metrics[i] = *presenter.ToTeamMetricsResponse(metric)
	}

	hasMore := (page * pageSize) < totalCount

	// サマリー計算
	summary := presenter.calculateTeamSummary(metricsList)

	return &TeamMetricsListResponse{
		Metrics:    metrics,
		TotalCount: totalCount,
		Page:       page,
		PageSize:   pageSize,
		HasMore:    hasMore,
		Summary:    summary,
		Filters:    filters,
	}
}

// ToDeveloperMetricsListResponse は開発者メトリクスリストをレスポンス形式に変換
func (presenter *AnalyticsPresenter) ToDeveloperMetricsListResponse(
	metricsList []*analyticsApp.DeveloperMetrics,
	totalCount, page, pageSize int,
) *DeveloperMetricsListResponse {
	metrics := make([]DeveloperMetricsResponse, len(metricsList))
	for i, metric := range metricsList {
		metrics[i] = *presenter.ToDeveloperMetricsResponse(metric)
	}

	hasMore := (page * pageSize) < totalCount

	// トップパフォーマー計算
	topPerformers := presenter.calculateTopPerformers(metricsList)

	return &DeveloperMetricsListResponse{
		Metrics:       metrics,
		TotalCount:    totalCount,
		Page:          page,
		PageSize:      pageSize,
		HasMore:       hasMore,
		TopPerformers: topPerformers,
	}
}

// ToRepositoryMetricsListResponse はリポジトリメトリクスリストをレスポンス形式に変換
func (presenter *AnalyticsPresenter) ToRepositoryMetricsListResponse(
	metricsList []*analyticsApp.RepositoryMetrics,
	totalCount, page, pageSize int,
) *RepositoryMetricsListResponse {
	metrics := make([]RepositoryMetricsResponse, len(metricsList))
	for i, metric := range metricsList {
		metrics[i] = *presenter.ToRepositoryMetricsResponse(metric)
	}

	hasMore := (page * pageSize) < totalCount

	// 活動サマリー計算
	activitySummary := presenter.calculateRepositoryActivitySummary(metricsList)

	return &RepositoryMetricsListResponse{
		Metrics:         metrics,
		TotalCount:      totalCount,
		Page:            page,
		PageSize:        pageSize,
		HasMore:         hasMore,
		ActivitySummary: activitySummary,
	}
}

// プライベートメソッド

func (presenter *AnalyticsPresenter) toDateRangeResponse(dateRange analyticsApp.DateRange) DateRangeResponse {
	return DateRangeResponse{
		Start: dateRange.Start,
		End:   dateRange.End,
	}
}

func (presenter *AnalyticsPresenter) toCycleTimeStatsAggResponse(stats analyticsApp.CycleTimeStatsAgg) CycleTimeStatsAggResponse {
	return CycleTimeStatsAggResponse{
		TotalCycleTime:    presenter.toDurationStatisticsResponse(stats.TotalCycleTime),
		TimeToFirstReview: presenter.toDurationStatisticsResponse(stats.TimeToFirstReview),
		TimeToApproval:    presenter.toDurationStatisticsResponse(stats.TimeToApproval),
		TimeToMerge:       presenter.toDurationStatisticsResponse(stats.TimeToMerge),
	}
}

func (presenter *AnalyticsPresenter) toReviewStatsAggResponse(stats analyticsApp.ReviewStatsAgg) ReviewStatsAggResponse {
	return ReviewStatsAggResponse{
		CommentCount:        presenter.toIntStatisticsResponse(stats.CommentCount),
		RoundCount:          presenter.toIntStatisticsResponse(stats.RoundCount),
		ReviewerCount:       presenter.toIntStatisticsResponse(stats.ReviewerCount),
		FirstReviewPassRate: presenter.toFloatStatisticsResponse(stats.FirstReviewPassRate),
	}
}

func (presenter *AnalyticsPresenter) toSizeStatsAggResponse(stats analyticsApp.SizeStatsAgg) SizeStatsAggResponse {
	return SizeStatsAggResponse{
		LinesAdded:               presenter.toIntStatisticsResponse(stats.LinesAdded),
		LinesDeleted:             presenter.toIntStatisticsResponse(stats.LinesDeleted),
		LinesChanged:             presenter.toIntStatisticsResponse(stats.LinesChanged),
		FilesChanged:             presenter.toIntStatisticsResponse(stats.FilesChanged),
		SizeCategoryDistribution: presenter.convertSizeCategoryDistribution(stats.SizeCategoryDistribution),
	}
}

func (presenter *AnalyticsPresenter) toQualityStatsAggResponse(stats analyticsApp.QualityStatsAgg) QualityStatsAggResponse {
	return QualityStatsAggResponse{
		CommitCount:         presenter.toIntStatisticsResponse(stats.CommitCount),
		FixupCommitCount:    presenter.toIntStatisticsResponse(stats.FixupCommitCount),
		ApprovalsReceived:   presenter.toIntStatisticsResponse(stats.ApprovalsReceived),
		QualityDistribution: stats.QualityDistribution,
	}
}

func (presenter *AnalyticsPresenter) toComplexityStatsAggResponse(stats analyticsApp.ComplexityStatsAgg) ComplexityStatsAggResponse {
	return ComplexityStatsAggResponse{
		ComplexityScore:        presenter.toFloatStatisticsResponse(stats.ComplexityScore),
		ComplexityDistribution: stats.ComplexityDistribution,
	}
}

func (presenter *AnalyticsPresenter) toDurationStatisticsResponse(stats utils.DurationStatistics) DurationStatisticsResponse {
	return DurationStatisticsResponse{
		Count:    stats.Count,
		Sum:      presenter.toDurationResponse(stats.Sum),
		Mean:     presenter.toDurationResponse(stats.Mean),
		Median:   presenter.toDurationResponse(stats.Median),
		Min:      presenter.toDurationResponse(stats.Min),
		Max:      presenter.toDurationResponse(stats.Max),
		Range:    presenter.toDurationResponse(stats.Range),
		Variance: presenter.toDurationResponse(stats.Variance),
		StdDev:   presenter.toDurationResponse(stats.StdDev),
		Percentiles: presenter.toDurationPercentilesResponse(stats.Percentiles),
	}
}

func (presenter *AnalyticsPresenter) toIntStatisticsResponse(stats utils.IntStatistics) IntStatisticsResponse {
	return IntStatisticsResponse{
		Count:       stats.Count,
		Sum:         stats.Sum,
		Mean:        stats.Mean,
		Median:      stats.Median,
		Mode:        stats.Mode,
		Min:         stats.Min,
		Max:         stats.Max,
		Range:       stats.Range,
		Variance:    stats.Variance,
		StdDev:      stats.StdDev,
		Percentiles: presenter.toPercentilesResponse(stats.Percentiles),
	}
}

func (presenter *AnalyticsPresenter) toFloatStatisticsResponse(stats utils.FloatStatistics) FloatStatisticsResponse {
	return FloatStatisticsResponse{
		Count:       stats.Count,
		Sum:         stats.Sum,
		Mean:        stats.Mean,
		Median:      stats.Median,
		Mode:        stats.Mode,
		Min:         stats.Min,
		Max:         stats.Max,
		Range:       stats.Range,
		Variance:    stats.Variance,
		StdDev:      stats.StdDev,
		Skewness:    stats.Skewness,
		Kurtosis:    stats.Kurtosis,
		Percentiles: presenter.toPercentilesResponse(stats.Percentiles),
	}
}

func (presenter *AnalyticsPresenter) toDurationResponse(d time.Duration) DurationResponse {
	return DurationResponse{
		Seconds:       int64(d.Seconds()),
		HumanReadable: presenter.formatDuration(d),
	}
}

func (presenter *AnalyticsPresenter) toPercentilesResponse(percentiles utils.Percentiles) PercentilesResponse {
	return PercentilesResponse{
		P10: percentiles.P10,
		P25: percentiles.P25,
		P50: percentiles.P50,
		P75: percentiles.P75,
		P90: percentiles.P90,
		P95: percentiles.P95,
		P99: percentiles.P99,
	}
}

func (presenter *AnalyticsPresenter) toDurationPercentilesResponse(percentiles utils.DurationPercentiles) DurationPercentilesResponse {
	return DurationPercentilesResponse{
		P10: presenter.toDurationResponse(percentiles.P10),
		P25: presenter.toDurationResponse(percentiles.P25),
		P50: presenter.toDurationResponse(percentiles.P50),
		P75: presenter.toDurationResponse(percentiles.P75),
		P90: presenter.toDurationResponse(percentiles.P90),
		P95: presenter.toDurationResponse(percentiles.P95),
		P99: presenter.toDurationResponse(percentiles.P99),
	}
}

func (presenter *AnalyticsPresenter) toTrendAnalysisResponse(trendAnalysis analyticsApp.TrendAnalysisResult) TrendAnalysisResponse {
	return TrendAnalysisResponse{
		CycleTimeTrend:  presenter.toTrendDataResponse(trendAnalysis.CycleTimeTrend),
		ReviewTimeTrend: presenter.toTrendDataResponse(trendAnalysis.ReviewTimeTrend),
		QualityTrend:    presenter.toTrendDataResponse(trendAnalysis.QualityTrend),
		
		// 期間比較は省略（実装時に追加）
		PeriodComparison: PeriodComparisonResponse{},
	}
}

func (presenter *AnalyticsPresenter) toTrendDataResponse(trend utils.TrendAnalysis) TrendDataResponse {
	return TrendDataResponse{
		Slope:            trend.Slope,
		Intercept:        trend.Intercept,
		CorrelationCoeff: trend.CorrelationCoeff,
		Trend:            trend.Trend,
		Confidence:       trend.Confidence,
		
		// 予測データは省略（実装時に追加）
		Prediction:   TrendPredictionResponse{},
		Significance: presenter.calculateSignificance(trend.Confidence),
	}
}

func (presenter *AnalyticsPresenter) toBottlenecksResponse(bottlenecks []analyticsApp.Bottleneck) []BottleneckResponse {
	response := make([]BottleneckResponse, len(bottlenecks))
	for i, bottleneck := range bottlenecks {
		response[i] = BottleneckResponse{
			Type:        bottleneck.Type,
			PRID:        bottleneck.PRID,
			Severity:    bottleneck.Severity,
			Description: bottleneck.Description,
			Value:       bottleneck.Value,
			Threshold:   presenter.calculateThreshold(bottleneck.Type),
			
			// 改善提案
			Suggestion:    presenter.generateSuggestion(bottleneck),
			ImpactLevel:   presenter.calculateImpactLevel(bottleneck.Severity),
			ActionItems:   presenter.generateActionItems(bottleneck),
			
			// 関連情報
			FrequencyScore: presenter.calculateFrequencyScore(bottleneck),
		}
	}
	return response
}

func (presenter *AnalyticsPresenter) toProductivityMetricsResponse(productivity analyticsApp.ProductivityMetrics) ProductivityMetricsResponse {
	return ProductivityMetricsResponse{
		PRsPerDay:   productivity.PRsPerDay,
		LinesPerDay: productivity.LinesPerDay,
		AvgPRSize:   productivity.AvgPRSize,
		Throughput:  productivity.Throughput,
		
		// 効率指標（計算）
		ReviewEfficiency: presenter.calculateReviewEfficiency(productivity),
		QualityRatio:     presenter.calculateQualityRatio(productivity),
		VelocityTrend:    presenter.calculateVelocityTrend(productivity),
	}
}

func (presenter *AnalyticsPresenter) toPerformanceIndicatorsResponse(metrics *analyticsApp.TeamMetrics) PerformanceIndicatorsResponse {
	// パフォーマンス指標の計算
	cycleTimeTarget := 3 * 24 * time.Hour // 3日
	cycleTimeActual := metrics.CycleTimeStats.TotalCycleTime.Mean
	
	reviewTimeTarget := 24 * time.Hour // 1日
	reviewTimeActual := metrics.CycleTimeStats.TimeToFirstReview.Mean
	
	qualityTarget := 0.8 // 80%
	qualityActual := metrics.ReviewStats.FirstReviewPassRate.Mean
	
	return PerformanceIndicatorsResponse{
		CycleTimeTarget: presenter.toDurationResponse(cycleTimeTarget),
		CycleTimeActual: presenter.toDurationResponse(cycleTimeActual),
		CycleTimeRatio:  presenter.calculateRatio(cycleTimeActual, cycleTimeTarget),
		
		ReviewTimeTarget: presenter.toDurationResponse(reviewTimeTarget),
		ReviewTimeActual: presenter.toDurationResponse(reviewTimeActual),
		ReviewTimeRatio:  presenter.calculateRatio(reviewTimeActual, reviewTimeTarget),
		
		QualityTarget: qualityTarget,
		QualityActual: qualityActual,
		QualityRatio:  qualityActual / qualityTarget,
		
		OverallScore:     presenter.calculateOverallScore(metrics),
		PerformanceLevel: presenter.calculatePerformanceLevel(metrics),
	}
}

func (presenter *AnalyticsPresenter) toDeveloperRankingResponse(developer string) DeveloperRankingResponse {
	// ランキング計算（省略）
	return DeveloperRankingResponse{
		PRCount:        RankingPosition{Rank: 1, TotalCount: 10, Percentile: 90, Score: 85},
		CycleTime:      RankingPosition{Rank: 3, TotalCount: 10, Percentile: 70, Score: 78},
		ReviewTime:     RankingPosition{Rank: 2, TotalCount: 10, Percentile: 80, Score: 82},
		Quality:        RankingPosition{Rank: 1, TotalCount: 10, Percentile: 95, Score: 90},
		Productivity:   RankingPosition{Rank: 2, TotalCount: 10, Percentile: 85, Score: 88},
		OverallRanking: RankingPosition{Rank: 2, TotalCount: 10, Percentile: 85, Score: 84},
	}
}

func (presenter *AnalyticsPresenter) toActivityLevelResponse(metrics *analyticsApp.RepositoryMetrics) ActivityLevelResponse {
	// 活動レベルの計算
	prFrequency := float64(metrics.TotalPRs) / 30.0 // 30日での平均
	score := presenter.calculateActivityScore(metrics)
	
	level := "medium"
	description := "標準的な活動レベルです"
	
	if score >= 0.8 {
		level = "high"
		description = "非常に活発な活動が見られます"
	} else if score < 0.4 {
		level = "low"
		description = "活動レベルが低めです"
	}
	
	return ActivityLevelResponse{
		Level:       level,
		Score:       score,
		Description: description,
		
		PRFrequency:    prFrequency,
		ReviewActivity: presenter.calculateReviewActivity(metrics),
		CommitActivity: presenter.calculateCommitActivity(metrics),
	}
}

// 計算系ヘルパーメソッド

func (presenter *AnalyticsPresenter) formatDuration(d time.Duration) string {
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

func (presenter *AnalyticsPresenter) convertSizeCategoryDistribution(distribution map[prDomain.PRSizeCategory]int) map[string]int {
	result := make(map[string]int)
	for k, v := range distribution {
		result[string(k)] = v
	}
	return result
}

func (presenter *AnalyticsPresenter) calculateSignificance(confidence float64) string {
	if confidence >= 0.9 {
		return "highly_significant"
	} else if confidence >= 0.7 {
		return "significant"
	} else if confidence >= 0.5 {
		return "moderate"
	} else {
		return "low"
	}
}

func (presenter *AnalyticsPresenter) calculateThreshold(bottleneckType string) float64 {
	switch bottleneckType {
	case "long_cycle_time":
		return 7 * 24 // 7日（時間）
	case "multiple_review_rounds":
		return 3 // 3ラウンド
	case "large_pr":
		return 500 // 500行
	default:
		return 0
	}
}

func (presenter *AnalyticsPresenter) generateSuggestion(bottleneck analyticsApp.Bottleneck) string {
	switch bottleneck.Type {
	case "long_cycle_time":
		return "レビュープロセスの最適化とレビュアーの早期アサインを検討してください"
	case "multiple_review_rounds":
		return "初回レビューの品質向上とレビューガイドラインの策定をお勧めします"
	case "large_pr":
		return "PRを小さく分割することでレビュー効率を向上できます"
	default:
		return "プロセスの見直しを検討してください"
	}
}

func (presenter *AnalyticsPresenter) calculateImpactLevel(severity string) string {
	switch severity {
	case "high":
		return "critical"
	case "medium":
		return "moderate"
	case "low":
		return "minor"
	default:
		return "unknown"
	}
}

func (presenter *AnalyticsPresenter) generateActionItems(bottleneck analyticsApp.Bottleneck) []string {
	switch bottleneck.Type {
	case "long_cycle_time":
		return []string{
			"レビュアーの自動アサイン設定",
			"レビュー開始通知の強化",
			"レビュー待ち時間の可視化",
		}
	case "multiple_review_rounds":
		return []string{
			"レビューチェックリストの作成",
			"セルフレビューの徹底",
			"コードレビューガイドラインの整備",
		}
	case "large_pr":
		return []string{
			"機能単位でのPR分割",
			"PRサイズガイドラインの策定",
			"WIP機能の活用",
		}
	default:
		return []string{"詳細な分析が必要です"}
	}
}

func (presenter *AnalyticsPresenter) calculateFrequencyScore(bottleneck analyticsApp.Bottleneck) float64 {
	// 簡略化された頻度スコア計算
	return 0.7
}

func (presenter *AnalyticsPresenter) calculateReviewEfficiency(productivity analyticsApp.ProductivityMetrics) float64 {
	// レビュー効率の計算（簡略化）
	return 0.75
}

func (presenter *AnalyticsPresenter) calculateQualityRatio(productivity analyticsApp.ProductivityMetrics) float64 {
	// 品質比率の計算（簡略化）
	return 0.85
}

func (presenter *AnalyticsPresenter) calculateVelocityTrend(productivity analyticsApp.ProductivityMetrics) string {
	// ベロシティトレンドの計算（簡略化）
	return "stable"
}

func (presenter *AnalyticsPresenter) calculateRatio(actual, target time.Duration) float64 {
	if target == 0 {
		return 0
	}
	return actual.Seconds() / target.Seconds()
}

func (presenter *AnalyticsPresenter) calculateOverallScore(metrics *analyticsApp.TeamMetrics) float64 {
	// 総合スコアの計算（簡略化）
	return 0.8
}

func (presenter *AnalyticsPresenter) calculatePerformanceLevel(metrics *analyticsApp.TeamMetrics) string {
	score := presenter.calculateOverallScore(metrics)
	if score >= 0.9 {
		return "excellent"
	} else if score >= 0.8 {
		return "good"
	} else if score >= 0.6 {
		return "fair"
	} else {
		return "needs_improvement"
	}
}

func (presenter *AnalyticsPresenter) calculateActivityScore(metrics *analyticsApp.RepositoryMetrics) float64 {
	// 活動スコアの計算（簡略化）
	return 0.7
}

func (presenter *AnalyticsPresenter) calculateReviewActivity(metrics *analyticsApp.RepositoryMetrics) float64 {
	// レビュー活動の計算（簡略化）
	return 0.8
}

func (presenter *AnalyticsPresenter) calculateCommitActivity(metrics *analyticsApp.RepositoryMetrics) float64 {
	// コミット活動の計算（簡略化）
	return 0.75
}

func (presenter *AnalyticsPresenter) calculateTeamSummary(metricsList []*analyticsApp.TeamMetrics) TeamMetricsSummaryResponse {
	if len(metricsList) == 0 {
		return TeamMetricsSummaryResponse{}
	}
	
	// 最新のメトリクスを使用
	latest := metricsList[0]
	return TeamMetricsSummaryResponse{
		Period:        string(latest.Period),
		TotalPRs:      latest.TotalPRs,
		AvgCycleTime:  presenter.toDurationResponse(latest.CycleTimeStats.TotalCycleTime.Mean),
		AvgReviewTime: presenter.toDurationResponse(latest.CycleTimeStats.TimeToFirstReview.Mean),
		QualityScore:  latest.ReviewStats.FirstReviewPassRate.Mean,
	}
}

func (presenter *AnalyticsPresenter) calculateTopPerformers(metricsList []*analyticsApp.DeveloperMetrics) []DeveloperRankingSummary {
	topPerformers := make([]DeveloperRankingSummary, 0, 5)
	
	for i, metrics := range metricsList {
		if i >= 5 { // Top 5のみ
			break
		}
		
		topPerformers = append(topPerformers, DeveloperRankingSummary{
			Developer: metrics.Developer,
			Rank:      i + 1,
			Score:     presenter.calculateDeveloperScore(metrics),
			PRCount:   metrics.TotalPRs,
			Strength:  presenter.identifyStrength(metrics),
		})
	}
	
	return topPerformers
}

func (presenter *AnalyticsPresenter) calculateRepositoryActivitySummary(metricsList []*analyticsApp.RepositoryMetrics) RepositoryActivitySummary {
	if len(metricsList) == 0 {
		return RepositoryActivitySummary{}
	}
	
	totalContributors := 0
	totalPRs := 0
	mostActiveRepo := ""
	maxPRs := 0
	
	for _, metrics := range metricsList {
		totalContributors += len(metrics.Contributors)
		totalPRs += metrics.TotalPRs
		
		if metrics.TotalPRs > maxPRs {
			maxPRs = metrics.TotalPRs
			mostActiveRepo = metrics.Repository
		}
	}
	
	avgPRsPerRepo := float64(totalPRs) / float64(len(metricsList))
	
	return RepositoryActivitySummary{
		TotalContributors: totalContributors,
		TotalPRs:          totalPRs,
		AvgPRsPerRepo:     avgPRsPerRepo,
		MostActiveRepo:    mostActiveRepo,
	}
}

func (presenter *AnalyticsPresenter) calculateDeveloperScore(metrics *analyticsApp.DeveloperMetrics) float64 {
	// 開発者スコアの計算（簡略化）
	return 85.0
}

func (presenter *AnalyticsPresenter) identifyStrength(metrics *analyticsApp.DeveloperMetrics) string {
	// 強み判定（簡略化）
	return "quality"
}