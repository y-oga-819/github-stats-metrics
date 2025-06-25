package analytics

import (
	"time"

	analyticsApp "github-stats-metrics/application/analytics"
	"github-stats-metrics/shared/utils"
)

// TeamMetricsResponse はチームメトリクスのAPIレスポンス
type TeamMetricsResponse struct {
	Period      string                    `json:"period"`
	TotalPRs    int                       `json:"totalPRs"`
	DateRange   DateRangeResponse         `json:"dateRange"`
	GeneratedAt time.Time                 `json:"generatedAt"`
	
	// 統計データ
	CycleTimeStats  CycleTimeStatsAggResponse `json:"cycleTimeStats"`
	ReviewStats     ReviewStatsAggResponse    `json:"reviewStats"`
	SizeStats       SizeStatsAggResponse      `json:"sizeStats"`
	QualityStats    QualityStatsAggResponse   `json:"qualityStats"`
	ComplexityStats ComplexityStatsAggResponse `json:"complexityStats"`
	
	// 分析結果
	TrendAnalysis TrendAnalysisResponse `json:"trendAnalysis"`
	Bottlenecks   []BottleneckResponse  `json:"bottlenecks"`
	
	// パフォーマンス指標
	Performance PerformanceIndicatorsResponse `json:"performance"`
}

// DeveloperMetricsResponse は開発者メトリクスのAPIレスポンス
type DeveloperMetricsResponse struct {
	Developer   string                    `json:"developer"`
	Period      string                    `json:"period"`
	TotalPRs    int                       `json:"totalPRs"`
	DateRange   DateRangeResponse         `json:"dateRange"`
	GeneratedAt time.Time                 `json:"generatedAt"`
	
	// 統計データ
	CycleTimeStats  CycleTimeStatsAggResponse `json:"cycleTimeStats"`
	ReviewStats     ReviewStatsAggResponse    `json:"reviewStats"`
	SizeStats       SizeStatsAggResponse      `json:"sizeStats"`
	QualityStats    QualityStatsAggResponse   `json:"qualityStats"`
	ComplexityStats ComplexityStatsAggResponse `json:"complexityStats"`
	
	// 生産性指標
	Productivity ProductivityMetricsResponse `json:"productivity"`
	
	// ランキング
	Ranking DeveloperRankingResponse `json:"ranking"`
}

// RepositoryMetricsResponse はリポジトリメトリクスのAPIレスポンス
type RepositoryMetricsResponse struct {
	Repository  string                    `json:"repository"`
	Period      string                    `json:"period"`
	TotalPRs    int                       `json:"totalPRs"`
	DateRange   DateRangeResponse         `json:"dateRange"`
	GeneratedAt time.Time                 `json:"generatedAt"`
	
	// 統計データ
	CycleTimeStats  CycleTimeStatsAggResponse `json:"cycleTimeStats"`
	ReviewStats     ReviewStatsAggResponse    `json:"reviewStats"`
	SizeStats       SizeStatsAggResponse      `json:"sizeStats"`
	QualityStats    QualityStatsAggResponse   `json:"qualityStats"`
	ComplexityStats ComplexityStatsAggResponse `json:"complexityStats"`
	
	// 貢献者情報
	Contributors []string `json:"contributors"`
	
	// 活動レベル
	ActivityLevel ActivityLevelResponse `json:"activityLevel"`
}

// TrendAnalysisResponse はトレンド分析のレスポンス
type TrendAnalysisResponse struct {
	CycleTimeTrend  TrendDataResponse `json:"cycleTimeTrend"`
	ReviewTimeTrend TrendDataResponse `json:"reviewTimeTrend"`
	QualityTrend    TrendDataResponse `json:"qualityTrend"`
	
	// 期間比較
	PeriodComparison PeriodComparisonResponse `json:"periodComparison"`
}

// TrendDataResponse は個別トレンドデータのレスポンス
type TrendDataResponse struct {
	Slope            float64 `json:"slope"`
	Intercept        float64 `json:"intercept"`
	CorrelationCoeff float64 `json:"correlationCoeff"`
	Trend            string  `json:"trend"`
	Confidence       float64 `json:"confidence"`
	
	// 予測データ
	Prediction   TrendPredictionResponse `json:"prediction"`
	Significance string                  `json:"significance"`
}

// DateRangeResponse は日付範囲のレスポンス
type DateRangeResponse struct {
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
}

// 統計レスポンス構造体群
type CycleTimeStatsAggResponse struct {
	TotalCycleTime    DurationStatisticsResponse `json:"totalCycleTime"`
	TimeToFirstReview DurationStatisticsResponse `json:"timeToFirstReview"`
	TimeToApproval    DurationStatisticsResponse `json:"timeToApproval"`
	TimeToMerge       DurationStatisticsResponse `json:"timeToMerge"`
}

type ReviewStatsAggResponse struct {
	CommentCount        IntStatisticsResponse   `json:"commentCount"`
	RoundCount          IntStatisticsResponse   `json:"roundCount"`
	ReviewerCount       IntStatisticsResponse   `json:"reviewerCount"`
	FirstReviewPassRate FloatStatisticsResponse `json:"firstReviewPassRate"`
}

type SizeStatsAggResponse struct {
	LinesAdded               IntStatisticsResponse     `json:"linesAdded"`
	LinesDeleted             IntStatisticsResponse     `json:"linesDeleted"`
	LinesChanged             IntStatisticsResponse     `json:"linesChanged"`
	FilesChanged             IntStatisticsResponse     `json:"filesChanged"`
	SizeCategoryDistribution map[string]int            `json:"sizeCategoryDistribution"`
}

type QualityStatsAggResponse struct {
	CommitCount         IntStatisticsResponse `json:"commitCount"`
	FixupCommitCount    IntStatisticsResponse `json:"fixupCommitCount"`
	ApprovalsReceived   IntStatisticsResponse `json:"approvalsReceived"`
	QualityDistribution map[string]int        `json:"qualityDistribution"`
}

type ComplexityStatsAggResponse struct {
	ComplexityScore        FloatStatisticsResponse `json:"complexityScore"`
	ComplexityDistribution map[string]int          `json:"complexityDistribution"`
}

// 基本統計レスポンス構造体群
type DurationStatisticsResponse struct {
	Count       int                       `json:"count"`
	Sum         DurationResponse          `json:"sum"`
	Mean        DurationResponse          `json:"mean"`
	Median      DurationResponse          `json:"median"`
	Min         DurationResponse          `json:"min"`
	Max         DurationResponse          `json:"max"`
	Range       DurationResponse          `json:"range"`
	Variance    DurationResponse          `json:"variance"`
	StdDev      DurationResponse          `json:"stdDev"`
	Percentiles DurationPercentilesResponse `json:"percentiles"`
}

type IntStatisticsResponse struct {
	Count       int                 `json:"count"`
	Sum         int                 `json:"sum"`
	Mean        float64             `json:"mean"`
	Median      float64             `json:"median"`
	Mode        int                 `json:"mode"`
	Min         int                 `json:"min"`
	Max         int                 `json:"max"`
	Range       int                 `json:"range"`
	Variance    float64             `json:"variance"`
	StdDev      float64             `json:"stdDev"`
	Percentiles PercentilesResponse `json:"percentiles"`
}

type FloatStatisticsResponse struct {
	Count       int                 `json:"count"`
	Sum         float64             `json:"sum"`
	Mean        float64             `json:"mean"`
	Median      float64             `json:"median"`
	Mode        float64             `json:"mode"`
	Min         float64             `json:"min"`
	Max         float64             `json:"max"`
	Range       float64             `json:"range"`
	Variance    float64             `json:"variance"`
	StdDev      float64             `json:"stdDev"`
	Skewness    float64             `json:"skewness"`
	Kurtosis    float64             `json:"kurtosis"`
	Percentiles PercentilesResponse `json:"percentiles"`
}

type DurationResponse struct {
	Seconds       int64  `json:"seconds"`
	HumanReadable string `json:"humanReadable"`
}

type PercentilesResponse struct {
	P10 float64 `json:"p10"`
	P25 float64 `json:"p25"`
	P50 float64 `json:"p50"`
	P75 float64 `json:"p75"`
	P90 float64 `json:"p90"`
	P95 float64 `json:"p95"`
	P99 float64 `json:"p99"`
}

type DurationPercentilesResponse struct {
	P10 DurationResponse `json:"p10"`
	P25 DurationResponse `json:"p25"`
	P50 DurationResponse `json:"p50"`
	P75 DurationResponse `json:"p75"`
	P90 DurationResponse `json:"p90"`
	P95 DurationResponse `json:"p95"`
	P99 DurationResponse `json:"p99"`
}

// 生産性・パフォーマンス関連レスポンス
type ProductivityMetricsResponse struct {
	PRsPerDay   float64 `json:"prsPerDay"`
	LinesPerDay float64 `json:"linesPerDay"`
	AvgPRSize   float64 `json:"avgPRSize"`
	Throughput  float64 `json:"throughput"`
	
	// 効率指標
	ReviewEfficiency float64 `json:"reviewEfficiency"`
	QualityRatio     float64 `json:"qualityRatio"`
	VelocityTrend    string  `json:"velocityTrend"`
}

type PerformanceIndicatorsResponse struct {
	CycleTimeTarget    DurationResponse `json:"cycleTimeTarget"`
	CycleTimeActual    DurationResponse `json:"cycleTimeActual"`
	CycleTimeRatio     float64          `json:"cycleTimeRatio"`
	
	ReviewTimeTarget   DurationResponse `json:"reviewTimeTarget"`
	ReviewTimeActual   DurationResponse `json:"reviewTimeActual"`
	ReviewTimeRatio    float64          `json:"reviewTimeRatio"`
	
	QualityTarget      float64 `json:"qualityTarget"`
	QualityActual      float64 `json:"qualityActual"`
	QualityRatio       float64 `json:"qualityRatio"`
	
	OverallScore       float64 `json:"overallScore"`
	PerformanceLevel   string  `json:"performanceLevel"`
}

// ランキング・比較関連レスポンス
type DeveloperRankingResponse struct {
	PRCount          RankingPosition `json:"prCount"`
	CycleTime        RankingPosition `json:"cycleTime"`
	ReviewTime       RankingPosition `json:"reviewTime"`
	Quality          RankingPosition `json:"quality"`
	Productivity     RankingPosition `json:"productivity"`
	OverallRanking   RankingPosition `json:"overallRanking"`
}

type RankingPosition struct {
	Rank       int     `json:"rank"`
	TotalCount int     `json:"totalCount"`
	Percentile float64 `json:"percentile"`
	Score      float64 `json:"score"`
}

type ActivityLevelResponse struct {
	Level       string  `json:"level"` // "high", "medium", "low"
	Score       float64 `json:"score"`
	Description string  `json:"description"`
	
	// 活動指標
	PRFrequency     float64 `json:"prFrequency"`
	ReviewActivity  float64 `json:"reviewActivity"`
	CommitActivity  float64 `json:"commitActivity"`
}

// 期間比較レスポンス
type PeriodComparisonResponse struct {
	PreviousPeriod TeamMetricsSummaryResponse `json:"previousPeriod"`
	CurrentPeriod  TeamMetricsSummaryResponse `json:"currentPeriod"`
	Changes        MetricsChangesResponse     `json:"changes"`
}

type TeamMetricsSummaryResponse struct {
	Period        string           `json:"period"`
	TotalPRs      int              `json:"totalPRs"`
	AvgCycleTime  DurationResponse `json:"avgCycleTime"`
	AvgReviewTime DurationResponse `json:"avgReviewTime"`
	QualityScore  float64          `json:"qualityScore"`
}

type MetricsChangesResponse struct {
	PRCountChange      ChangeIndicatorResponse `json:"prCountChange"`
	CycleTimeChange    ChangeIndicatorResponse `json:"cycleTimeChange"`
	ReviewTimeChange   ChangeIndicatorResponse `json:"reviewTimeChange"`
	QualityChange      ChangeIndicatorResponse `json:"qualityChange"`
	OverallTrend       string                  `json:"overallTrend"`
}

type ChangeIndicatorResponse struct {
	Value       float64 `json:"value"`
	Percentage  float64 `json:"percentage"`
	Direction   string  `json:"direction"` // "up", "down", "stable"
	Significance string  `json:"significance"` // "significant", "minor", "negligible"
}

// トレンド予測レスポンス
type TrendPredictionResponse struct {
	NextPeriodValue   float64 `json:"nextPeriodValue"`
	ConfidenceLevel   float64 `json:"confidenceLevel"`
	PredictionRange   PredictionRangeResponse `json:"predictionRange"`
	RecommendedAction string  `json:"recommendedAction"`
}

type PredictionRangeResponse struct {
	Lower float64 `json:"lower"`
	Upper float64 `json:"upper"`
}

// ボトルネックレスポンス（拡張版）
type BottleneckResponse struct {
	Type        string  `json:"type"`
	PRID        string  `json:"prId,omitempty"`
	Severity    string  `json:"severity"`
	Description string  `json:"description"`
	Value       float64 `json:"value"`
	Threshold   float64 `json:"threshold"`
	
	// 改善提案
	Suggestion    string   `json:"suggestion"`
	ImpactLevel   string   `json:"impactLevel"`
	ActionItems   []string `json:"actionItems"`
	
	// 関連情報
	AffectedDevelopers []string `json:"affectedDevelopers,omitempty"`
	FrequencyScore     float64  `json:"frequencyScore"`
}

// リスト・ページネーション関連レスポンス
type TeamMetricsListResponse struct {
	Metrics    []TeamMetricsResponse `json:"metrics"`
	TotalCount int                   `json:"totalCount"`
	Page       int                   `json:"page"`
	PageSize   int                   `json:"pageSize"`
	HasMore    bool                  `json:"hasMore"`
	
	// メタ情報
	Summary TeamMetricsSummaryResponse `json:"summary"`
	Filters AppliedFiltersResponse     `json:"filters"`
}

type DeveloperMetricsListResponse struct {
	Metrics    []DeveloperMetricsResponse `json:"metrics"`
	TotalCount int                        `json:"totalCount"`
	Page       int                        `json:"page"`
	PageSize   int                        `json:"pageSize"`
	HasMore    bool                       `json:"hasMore"`
	
	// ランキング情報
	TopPerformers []DeveloperRankingSummary `json:"topPerformers"`
}

type RepositoryMetricsListResponse struct {
	Metrics    []RepositoryMetricsResponse `json:"metrics"`
	TotalCount int                         `json:"totalCount"`
	Page       int                         `json:"page"`
	PageSize   int                         `json:"pageSize"`
	HasMore    bool                        `json:"hasMore"`
	
	// 活動サマリー
	ActivitySummary RepositoryActivitySummary `json:"activitySummary"`
}

type DeveloperRankingSummary struct {
	Developer   string  `json:"developer"`
	Rank        int     `json:"rank"`
	Score       float64 `json:"score"`
	PRCount     int     `json:"prCount"`
	Strength    string  `json:"strength"`
}

type RepositoryActivitySummary struct {
	TotalContributors int     `json:"totalContributors"`
	TotalPRs          int     `json:"totalPRs"`
	AvgPRsPerRepo     float64 `json:"avgPRsPerRepo"`
	MostActiveRepo    string  `json:"mostActiveRepo"`
}

type AppliedFiltersResponse struct {
	Period       string   `json:"period"`
	DateRange    DateRangeResponse `json:"dateRange"`
	Developers   []string `json:"developers,omitempty"`
	Repositories []string `json:"repositories,omitempty"`
	Metrics      []string `json:"metrics,omitempty"`
}

// エラー・成功レスポンス
type AnalyticsErrorResponse struct {
	Error     string                 `json:"error"`
	Code      string                 `json:"code"`
	Message   string                 `json:"message"`
	Details   interface{}            `json:"details,omitempty"`
	RequestID string                 `json:"requestId"`
	Timestamp time.Time              `json:"timestamp"`
}

type AnalyticsSuccessResponse struct {
	Success   bool        `json:"success"`
	Message   string      `json:"message"`
	Data      interface{} `json:"data"`
	RequestID string      `json:"requestId"`
	Timestamp time.Time   `json:"timestamp"`
}

// メタ情報レスポンス
type AnalyticsMetaResponse struct {
	GeneratedAt     time.Time `json:"generatedAt"`
	ProcessingTime  string    `json:"processingTime"`
	DataFreshness   string    `json:"dataFreshness"`
	RecordCount     int       `json:"recordCount"`
	QualityScore    float64   `json:"qualityScore"`
	ReliabilityNote string    `json:"reliabilityNote,omitempty"`
}