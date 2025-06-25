package analytics

import (
	"context"
	"fmt"
	"time"

	prDomain "github-stats-metrics/domain/pull_request"
	"github-stats-metrics/shared/utils"
)

// MetricsAggregator はメトリクスを集計するサービス
type MetricsAggregator struct {
	statsCalc *utils.StatisticsCalculator
}

// NewMetricsAggregator は新しいメトリクス集計器を作成
func NewMetricsAggregator() *MetricsAggregator {
	return &MetricsAggregator{
		statsCalc: utils.NewStatisticsCalculator(),
	}
}

// AggregateTeamMetrics はチーム全体のメトリクスを集計
func (aggregator *MetricsAggregator) AggregateTeamMetrics(ctx context.Context, metrics []*prDomain.PRMetrics, period AggregationPeriod) (*TeamMetrics, error) {
	if len(metrics) == 0 {
		return &TeamMetrics{Period: period}, nil
	}

	teamMetrics := &TeamMetrics{
		Period:      period,
		TotalPRs:    len(metrics),
		DateRange:   aggregator.calculateDateRange(metrics),
		GeneratedAt: time.Now(),
	}

	// サイクルタイム統計
	teamMetrics.CycleTimeStats = aggregator.aggregateCycleTimeStats(metrics)
	
	// レビュー統計
	teamMetrics.ReviewStats = aggregator.aggregateReviewStats(metrics)
	
	// サイズ統計
	teamMetrics.SizeStats = aggregator.aggregateSizeStats(metrics)
	
	// 品質統計
	teamMetrics.QualityStats = aggregator.aggregateQualityStats(metrics)
	
	// 複雑度統計
	teamMetrics.ComplexityStats = aggregator.aggregateComplexityStats(metrics)
	
	// トレンド分析
	teamMetrics.TrendAnalysis = aggregator.calculateTrends(metrics)
	
	// ボトルネック分析
	teamMetrics.Bottlenecks = aggregator.identifyBottlenecks(metrics)

	return teamMetrics, nil
}

// AggregateDeveloperMetrics は開発者別のメトリクスを集計
func (aggregator *MetricsAggregator) AggregateDeveloperMetrics(ctx context.Context, metrics []*prDomain.PRMetrics, period AggregationPeriod) (map[string]*DeveloperMetrics, error) {
	developerData := aggregator.groupByDeveloper(metrics)
	result := make(map[string]*DeveloperMetrics)

	for developer, devMetrics := range developerData {
		result[developer] = &DeveloperMetrics{
			Developer:      developer,
			Period:         period,
			TotalPRs:       len(devMetrics),
			DateRange:      aggregator.calculateDateRange(devMetrics),
			GeneratedAt:    time.Now(),
			CycleTimeStats: aggregator.aggregateCycleTimeStats(devMetrics),
			ReviewStats:    aggregator.aggregateReviewStats(devMetrics),
			SizeStats:      aggregator.aggregateSizeStats(devMetrics),
			QualityStats:   aggregator.aggregateQualityStats(devMetrics),
			ComplexityStats: aggregator.aggregateComplexityStats(devMetrics),
			Productivity:   aggregator.calculateProductivity(devMetrics, period),
		}
	}

	return result, nil
}

// AggregateRepositoryMetrics はリポジトリ別のメトリクスを集計
func (aggregator *MetricsAggregator) AggregateRepositoryMetrics(ctx context.Context, metrics []*prDomain.PRMetrics, period AggregationPeriod) (map[string]*RepositoryMetrics, error) {
	repoData := aggregator.groupByRepository(metrics)
	result := make(map[string]*RepositoryMetrics)

	for repository, repoMetrics := range repoData {
		result[repository] = &RepositoryMetrics{
			Repository:      repository,
			Period:          period,
			TotalPRs:        len(repoMetrics),
			DateRange:       aggregator.calculateDateRange(repoMetrics),
			GeneratedAt:     time.Now(),
			CycleTimeStats:  aggregator.aggregateCycleTimeStats(repoMetrics),
			ReviewStats:     aggregator.aggregateReviewStats(repoMetrics),
			SizeStats:       aggregator.aggregateSizeStats(repoMetrics),
			QualityStats:    aggregator.aggregateQualityStats(repoMetrics),
			ComplexityStats: aggregator.aggregateComplexityStats(repoMetrics),
			Contributors:    aggregator.getUniqueContributors(repoMetrics),
		}
	}

	return result, nil
}

// aggregateCycleTimeStats はサイクルタイム統計を集計
func (aggregator *MetricsAggregator) aggregateCycleTimeStats(metrics []*prDomain.PRMetrics) CycleTimeStatsAgg {
	var totalCycleTimes []time.Duration
	var reviewTimes []time.Duration
	var approvalTimes []time.Duration
	var mergeTimes []time.Duration

	for _, metric := range metrics {
		if metric.TimeMetrics.TotalCycleTime != nil {
			totalCycleTimes = append(totalCycleTimes, *metric.TimeMetrics.TotalCycleTime)
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

	return CycleTimeStatsAgg{
		TotalCycleTime:    aggregator.statsCalc.CalculateDurationStatistics(totalCycleTimes),
		TimeToFirstReview: aggregator.statsCalc.CalculateDurationStatistics(reviewTimes),
		TimeToApproval:    aggregator.statsCalc.CalculateDurationStatistics(approvalTimes),
		TimeToMerge:       aggregator.statsCalc.CalculateDurationStatistics(mergeTimes),
	}
}

// aggregateReviewStats はレビュー統計を集計
func (aggregator *MetricsAggregator) aggregateReviewStats(metrics []*prDomain.PRMetrics) ReviewStatsAgg {
	var commentCounts []int
	var roundCounts []int
	var reviewerCounts []int
	var firstPassRates []float64

	for _, metric := range metrics {
		commentCounts = append(commentCounts, metric.QualityMetrics.ReviewCommentCount)
		roundCounts = append(roundCounts, metric.QualityMetrics.ReviewRoundCount)
		reviewerCounts = append(reviewerCounts, metric.QualityMetrics.ReviewerCount)
		firstPassRates = append(firstPassRates, metric.QualityMetrics.FirstReviewPassRate)
	}

	return ReviewStatsAgg{
		CommentCount:      aggregator.statsCalc.CalculateIntStatistics(commentCounts),
		RoundCount:        aggregator.statsCalc.CalculateIntStatistics(roundCounts),
		ReviewerCount:     aggregator.statsCalc.CalculateIntStatistics(reviewerCounts),
		FirstReviewPassRate: aggregator.statsCalc.CalculateFloatStatistics(firstPassRates),
	}
}

// aggregateSizeStats はサイズ統計を集計
func (aggregator *MetricsAggregator) aggregateSizeStats(metrics []*prDomain.PRMetrics) SizeStatsAgg {
	var linesAdded []int
	var linesDeleted []int
	var linesChanged []int
	var filesChanged []int
	sizeCategoryCount := make(map[prDomain.PRSizeCategory]int)

	for _, metric := range metrics {
		linesAdded = append(linesAdded, metric.SizeMetrics.LinesAdded)
		linesDeleted = append(linesDeleted, metric.SizeMetrics.LinesDeleted)
		linesChanged = append(linesChanged, metric.SizeMetrics.LinesChanged)
		filesChanged = append(filesChanged, metric.SizeMetrics.FilesChanged)
		sizeCategoryCount[metric.SizeCategory]++
	}

	return SizeStatsAgg{
		LinesAdded:         aggregator.statsCalc.CalculateIntStatistics(linesAdded),
		LinesDeleted:       aggregator.statsCalc.CalculateIntStatistics(linesDeleted),
		LinesChanged:       aggregator.statsCalc.CalculateIntStatistics(linesChanged),
		FilesChanged:       aggregator.statsCalc.CalculateIntStatistics(filesChanged),
		SizeCategoryDistribution: sizeCategoryCount,
	}
}

// aggregateQualityStats は品質統計を集計
func (aggregator *MetricsAggregator) aggregateQualityStats(metrics []*prDomain.PRMetrics) QualityStatsAgg {
	var commitCounts []int
	var fixupCommitCounts []int
	var approvalCounts []int

	qualityCategories := make(map[string]int)

	for _, metric := range metrics {
		commitCounts = append(commitCounts, metric.QualityMetrics.CommitCount)
		fixupCommitCounts = append(fixupCommitCounts, metric.QualityMetrics.FixupCommitCount)
		approvalCounts = append(approvalCounts, metric.QualityMetrics.ApprovalsReceived)

		// 品質カテゴリの判定
		category := aggregator.categorizeQuality(metric)
		qualityCategories[category]++
	}

	return QualityStatsAgg{
		CommitCount:       aggregator.statsCalc.CalculateIntStatistics(commitCounts),
		FixupCommitCount:  aggregator.statsCalc.CalculateIntStatistics(fixupCommitCounts),
		ApprovalsReceived: aggregator.statsCalc.CalculateIntStatistics(approvalCounts),
		QualityDistribution: qualityCategories,
	}
}

// aggregateComplexityStats は複雑度統計を集計
func (aggregator *MetricsAggregator) aggregateComplexityStats(metrics []*prDomain.PRMetrics) ComplexityStatsAgg {
	var complexityScores []float64
	complexityLevels := make(map[string]int)

	for _, metric := range metrics {
		complexityScores = append(complexityScores, metric.ComplexityScore)
		
		// 複雑度レベルの分類
		level := aggregator.categorizeComplexity(metric.ComplexityScore)
		complexityLevels[level]++
	}

	return ComplexityStatsAgg{
		ComplexityScore:      aggregator.statsCalc.CalculateFloatStatistics(complexityScores),
		ComplexityDistribution: complexityLevels,
	}
}

// calculateTrends はトレンド分析を実行
func (aggregator *MetricsAggregator) calculateTrends(metrics []*prDomain.PRMetrics) TrendAnalysisResult {
	// 時系列データの準備
	dailyData := aggregator.groupByDay(metrics)
	
	var cycleTimes []float64
	var reviewTimes []float64
	var qualityScores []float64

	for _, dayMetrics := range dailyData {
		if len(dayMetrics) == 0 {
			continue
		}
		
		// 日次平均の計算
		avgCycleTime := aggregator.calculateAverageCycleTime(dayMetrics)
		avgReviewTime := aggregator.calculateAverageReviewTime(dayMetrics)
		avgQualityScore := aggregator.calculateAverageQualityScore(dayMetrics)
		
		cycleTimes = append(cycleTimes, avgCycleTime)
		reviewTimes = append(reviewTimes, avgReviewTime)
		qualityScores = append(qualityScores, avgQualityScore)
	}

	return TrendAnalysisResult{
		CycleTimeTrend:    aggregator.statsCalc.AnalyzeTrend(cycleTimes),
		ReviewTimeTrend:   aggregator.statsCalc.AnalyzeTrend(reviewTimes),
		QualityTrend:      aggregator.statsCalc.AnalyzeTrend(qualityScores),
	}
}

// identifyBottlenecks はボトルネックを特定
func (aggregator *MetricsAggregator) identifyBottlenecks(metrics []*prDomain.PRMetrics) []Bottleneck {
	var bottlenecks []Bottleneck

	// 長いサイクルタイムのPR
	for _, metric := range metrics {
		if metric.TimeMetrics.TotalCycleTime != nil && *metric.TimeMetrics.TotalCycleTime > 7*24*time.Hour {
			bottlenecks = append(bottlenecks, Bottleneck{
				Type:        "long_cycle_time",
				PRID:        metric.PRID,
				Severity:    "high",
				Description: fmt.Sprintf("サイクルタイムが%s", formatDuration(*metric.TimeMetrics.TotalCycleTime)),
				Value:       metric.TimeMetrics.TotalCycleTime.Hours(),
			})
		}
	}

	// 多数のレビューラウンド
	for _, metric := range metrics {
		if metric.QualityMetrics.ReviewRoundCount > 5 {
			bottlenecks = append(bottlenecks, Bottleneck{
				Type:        "multiple_review_rounds",
				PRID:        metric.PRID,
				Severity:    "medium",
				Description: fmt.Sprintf("%d回のレビューラウンド", metric.QualityMetrics.ReviewRoundCount),
				Value:       float64(metric.QualityMetrics.ReviewRoundCount),
			})
		}
	}

	// 大きなPR
	for _, metric := range metrics {
		if metric.IsLargePR() {
			bottlenecks = append(bottlenecks, Bottleneck{
				Type:        "large_pr",
				PRID:        metric.PRID,
				Severity:    "medium",
				Description: fmt.Sprintf("大きなPR（%d行の変更）", metric.SizeMetrics.LinesChanged),
				Value:       float64(metric.SizeMetrics.LinesChanged),
			})
		}
	}

	return bottlenecks
}

// ヘルパーメソッド群

func (aggregator *MetricsAggregator) groupByDeveloper(metrics []*prDomain.PRMetrics) map[string][]*prDomain.PRMetrics {
	result := make(map[string][]*prDomain.PRMetrics)
	for _, metric := range metrics {
		result[metric.Author] = append(result[metric.Author], metric)
	}
	return result
}

func (aggregator *MetricsAggregator) groupByRepository(metrics []*prDomain.PRMetrics) map[string][]*prDomain.PRMetrics {
	result := make(map[string][]*prDomain.PRMetrics)
	for _, metric := range metrics {
		result[metric.Repository] = append(result[metric.Repository], metric)
	}
	return result
}

func (aggregator *MetricsAggregator) groupByDay(metrics []*prDomain.PRMetrics) map[string][]*prDomain.PRMetrics {
	result := make(map[string][]*prDomain.PRMetrics)
	for _, metric := range metrics {
		day := metric.CreatedAt.Format("2006-01-02")
		result[day] = append(result[day], metric)
	}
	return result
}

func (aggregator *MetricsAggregator) calculateDateRange(metrics []*prDomain.PRMetrics) DateRange {
	if len(metrics) == 0 {
		return DateRange{}
	}

	start := metrics[0].CreatedAt
	end := metrics[0].CreatedAt

	for _, metric := range metrics {
		if metric.CreatedAt.Before(start) {
			start = metric.CreatedAt
		}
		if metric.CreatedAt.After(end) {
			end = metric.CreatedAt
		}
		if metric.MergedAt != nil && metric.MergedAt.After(end) {
			end = *metric.MergedAt
		}
	}

	return DateRange{Start: start, End: end}
}

func (aggregator *MetricsAggregator) getUniqueContributors(metrics []*prDomain.PRMetrics) []string {
	contributors := make(map[string]bool)
	for _, metric := range metrics {
		contributors[metric.Author] = true
	}

	result := make([]string, 0, len(contributors))
	for contributor := range contributors {
		result = append(result, contributor)
	}
	return result
}

func (aggregator *MetricsAggregator) calculateProductivity(metrics []*prDomain.PRMetrics, period AggregationPeriod) ProductivityMetrics {
	if len(metrics) == 0 {
		return ProductivityMetrics{}
	}

	dateRange := aggregator.calculateDateRange(metrics)
	var days float64
	
	switch period {
	case AggregationPeriodDaily:
		days = 1
	case AggregationPeriodWeekly:
		days = 7
	case AggregationPeriodMonthly:
		days = 30
	default:
		days = dateRange.End.Sub(dateRange.Start).Hours() / 24
		if days < 1 {
			days = 1
		}
	}

	totalLines := 0
	for _, metric := range metrics {
		totalLines += metric.SizeMetrics.LinesChanged
	}

	return ProductivityMetrics{
		PRsPerDay:      float64(len(metrics)) / days,
		LinesPerDay:    float64(totalLines) / days,
		AvgPRSize:      float64(totalLines) / float64(len(metrics)),
		Throughput:     float64(len(metrics)),
	}
}

func (aggregator *MetricsAggregator) categorizeQuality(metric *prDomain.PRMetrics) string {
	if metric.IsHighQuality() {
		return "high"
	} else if metric.QualityMetrics.ReviewCommentCount > 10 {
		return "low"
	} else {
		return "medium"
	}
}

func (aggregator *MetricsAggregator) categorizeComplexity(score float64) string {
	switch {
	case score <= 1.5:
		return "very_low"
	case score <= 2.5:
		return "low"
	case score <= 4.0:
		return "medium"
	case score <= 6.0:
		return "high"
	default:
		return "very_high"
	}
}

func (aggregator *MetricsAggregator) calculateAverageCycleTime(metrics []*prDomain.PRMetrics) float64 {
	var total time.Duration
	count := 0
	
	for _, metric := range metrics {
		if metric.TimeMetrics.TotalCycleTime != nil {
			total += *metric.TimeMetrics.TotalCycleTime
			count++
		}
	}
	
	if count == 0 {
		return 0
	}
	return (total / time.Duration(count)).Hours()
}

func (aggregator *MetricsAggregator) calculateAverageReviewTime(metrics []*prDomain.PRMetrics) float64 {
	var total time.Duration
	count := 0
	
	for _, metric := range metrics {
		if metric.TimeMetrics.TimeToFirstReview != nil {
			total += *metric.TimeMetrics.TimeToFirstReview
			count++
		}
	}
	
	if count == 0 {
		return 0
	}
	return (total / time.Duration(count)).Hours()
}

func (aggregator *MetricsAggregator) calculateAverageQualityScore(metrics []*prDomain.PRMetrics) float64 {
	total := 0.0
	for _, metric := range metrics {
		// 品質スコア = 初回通過率 - (コメント数/10)
		score := metric.QualityMetrics.FirstReviewPassRate - (float64(metric.QualityMetrics.ReviewCommentCount) / 10.0)
		if score < 0 {
			score = 0
		}
		total += score
	}
	return total / float64(len(metrics))
}

func formatDuration(d time.Duration) string {
	if d < time.Hour {
		return fmt.Sprintf("%.0f分", d.Minutes())
	} else if d < 24*time.Hour {
		return fmt.Sprintf("%.1f時間", d.Hours())
	} else {
		return fmt.Sprintf("%.1f日", d.Hours()/24)
	}
}

// 構造体定義

// AggregationPeriod は集計期間
type AggregationPeriod string

const (
	AggregationPeriodDaily   AggregationPeriod = "daily"
	AggregationPeriodWeekly  AggregationPeriod = "weekly"
	AggregationPeriodMonthly AggregationPeriod = "monthly"
	AggregationPeriodCustom  AggregationPeriod = "custom"
)

// DateRange は日付範囲
type DateRange struct {
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
}

// TeamMetrics はチーム全体のメトリクス
type TeamMetrics struct {
	Period          AggregationPeriod    `json:"period"`
	TotalPRs        int                  `json:"totalPRs"`
	DateRange       DateRange            `json:"dateRange"`
	GeneratedAt     time.Time            `json:"generatedAt"`
	CycleTimeStats  CycleTimeStatsAgg    `json:"cycleTimeStats"`
	ReviewStats     ReviewStatsAgg       `json:"reviewStats"`
	SizeStats       SizeStatsAgg         `json:"sizeStats"`
	QualityStats    QualityStatsAgg      `json:"qualityStats"`
	ComplexityStats ComplexityStatsAgg   `json:"complexityStats"`
	TrendAnalysis   TrendAnalysisResult  `json:"trendAnalysis"`
	Bottlenecks     []Bottleneck         `json:"bottlenecks"`
}

// DeveloperMetrics は開発者別のメトリクス
type DeveloperMetrics struct {
	Developer       string               `json:"developer"`
	Period          AggregationPeriod    `json:"period"`
	TotalPRs        int                  `json:"totalPRs"`
	DateRange       DateRange            `json:"dateRange"`
	GeneratedAt     time.Time            `json:"generatedAt"`
	CycleTimeStats  CycleTimeStatsAgg    `json:"cycleTimeStats"`
	ReviewStats     ReviewStatsAgg       `json:"reviewStats"`
	SizeStats       SizeStatsAgg         `json:"sizeStats"`
	QualityStats    QualityStatsAgg      `json:"qualityStats"`
	ComplexityStats ComplexityStatsAgg   `json:"complexityStats"`
	Productivity    ProductivityMetrics  `json:"productivity"`
}

// RepositoryMetrics はリポジトリ別のメトリクス
type RepositoryMetrics struct {
	Repository      string               `json:"repository"`
	Period          AggregationPeriod    `json:"period"`
	TotalPRs        int                  `json:"totalPRs"`
	DateRange       DateRange            `json:"dateRange"`
	GeneratedAt     time.Time            `json:"generatedAt"`
	CycleTimeStats  CycleTimeStatsAgg    `json:"cycleTimeStats"`
	ReviewStats     ReviewStatsAgg       `json:"reviewStats"`
	SizeStats       SizeStatsAgg         `json:"sizeStats"`
	QualityStats    QualityStatsAgg      `json:"qualityStats"`
	ComplexityStats ComplexityStatsAgg   `json:"complexityStats"`
	Contributors    []string             `json:"contributors"`
}

// 各種統計構造体
type CycleTimeStatsAgg struct {
	TotalCycleTime    utils.DurationStatistics `json:"totalCycleTime"`
	TimeToFirstReview utils.DurationStatistics `json:"timeToFirstReview"`
	TimeToApproval    utils.DurationStatistics `json:"timeToApproval"`
	TimeToMerge       utils.DurationStatistics `json:"timeToMerge"`
}

type ReviewStatsAgg struct {
	CommentCount        utils.IntStatistics   `json:"commentCount"`
	RoundCount          utils.IntStatistics   `json:"roundCount"`
	ReviewerCount       utils.IntStatistics   `json:"reviewerCount"`
	FirstReviewPassRate utils.FloatStatistics `json:"firstReviewPassRate"`
}

type SizeStatsAgg struct {
	LinesAdded               utils.IntStatistics                    `json:"linesAdded"`
	LinesDeleted             utils.IntStatistics                    `json:"linesDeleted"`
	LinesChanged             utils.IntStatistics                    `json:"linesChanged"`
	FilesChanged             utils.IntStatistics                    `json:"filesChanged"`
	SizeCategoryDistribution map[prDomain.PRSizeCategory]int        `json:"sizeCategoryDistribution"`
}

type QualityStatsAgg struct {
	CommitCount         utils.IntStatistics `json:"commitCount"`
	FixupCommitCount    utils.IntStatistics `json:"fixupCommitCount"`
	ApprovalsReceived   utils.IntStatistics `json:"approvalsReceived"`
	QualityDistribution map[string]int      `json:"qualityDistribution"`
}

type ComplexityStatsAgg struct {
	ComplexityScore        utils.FloatStatistics `json:"complexityScore"`
	ComplexityDistribution map[string]int        `json:"complexityDistribution"`
}

type ProductivityMetrics struct {
	PRsPerDay   float64 `json:"prsPerDay"`
	LinesPerDay float64 `json:"linesPerDay"`
	AvgPRSize   float64 `json:"avgPRSize"`
	Throughput  float64 `json:"throughput"`
}

type TrendAnalysisResult struct {
	CycleTimeTrend  utils.TrendAnalysis `json:"cycleTimeTrend"`
	ReviewTimeTrend utils.TrendAnalysis `json:"reviewTimeTrend"`
	QualityTrend    utils.TrendAnalysis `json:"qualityTrend"`
}

type Bottleneck struct {
	Type        string  `json:"type"`
	PRID        string  `json:"prId"`
	Severity    string  `json:"severity"`
	Description string  `json:"description"`
	Value       float64 `json:"value"`
}