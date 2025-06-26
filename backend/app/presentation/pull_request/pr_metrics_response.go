package pull_request

import (
	"time"
)

// PRMetricsResponse はPRメトリクスのAPIレスポンス
type PRMetricsResponse struct {
	// 基本情報
	PRID       string    `json:"prId"`
	PRNumber   int       `json:"prNumber"`
	Title      string    `json:"title"`
	Author     string    `json:"author"`
	Repository string    `json:"repository"`
	CreatedAt  time.Time `json:"createdAt"`
	MergedAt   *time.Time `json:"mergedAt,omitempty"`

	// サイズメトリクス
	SizeMetrics PRSizeMetricsResponse `json:"sizeMetrics"`

	// 時間メトリクス
	TimeMetrics PRTimeMetricsResponse `json:"timeMetrics"`

	// 品質メトリクス
	QualityMetrics PRQualityMetricsResponse `json:"qualityMetrics"`

	// 複雑度
	ComplexityScore  float64 `json:"complexityScore"`
	ComplexityLevel  string  `json:"complexityLevel"`
	SizeCategory     string  `json:"sizeCategory"`

	// 分析結果
	AnalysisResults PRAnalysisResultsResponse `json:"analysisResults"`
}

// PRSizeMetricsResponse はPRサイズメトリクスのレスポンス
type PRSizeMetricsResponse struct {
	LinesAdded        int                       `json:"linesAdded"`
	LinesDeleted      int                       `json:"linesDeleted"`
	LinesChanged      int                       `json:"linesChanged"`
	FilesChanged      int                       `json:"filesChanged"`
	FileTypeBreakdown map[string]int            `json:"fileTypeBreakdown"`
	DirectoryCount    int                       `json:"directoryCount"`
	FileChanges       []FileChangeResponse      `json:"fileChanges"`
}

// FileChangeResponse はファイル変更情報のレスポンス
type FileChangeResponse struct {
	FileName     string `json:"fileName"`
	FileType     string `json:"fileType"`
	LinesAdded   int    `json:"linesAdded"`
	LinesDeleted int    `json:"linesDeleted"`
	IsNewFile    bool   `json:"isNewFile"`
	IsDeleted    bool   `json:"isDeleted"`
	IsRenamed    bool   `json:"isRenamed"`
}

// PRTimeMetricsResponse はPR時間メトリクスのレスポンス
type PRTimeMetricsResponse struct {
	TotalCycleTime     *DurationResponse `json:"totalCycleTime,omitempty"`
	TimeToFirstReview  *DurationResponse `json:"timeToFirstReview,omitempty"`
	TimeToApproval     *DurationResponse `json:"timeToApproval,omitempty"`
	TimeToMerge        *DurationResponse `json:"timeToMerge,omitempty"`
	ReviewWaitTime     *DurationResponse `json:"reviewWaitTime,omitempty"`
	ReviewActiveTime   *DurationResponse `json:"reviewActiveTime,omitempty"`
	FirstCommitToMerge *DurationResponse `json:"firstCommitToMerge,omitempty"`
	CreatedHour        int               `json:"createdHour"`
	MergedHour         *int              `json:"mergedHour,omitempty"`
}

// DurationResponse は時間の人間が読みやすい形式のレスポンス
type DurationResponse struct {
	Seconds     int64  `json:"seconds"`
	HumanReadable string `json:"humanReadable"`
}

// PRQualityMetricsResponse はPR品質メトリクスのレスポンス
type PRQualityMetricsResponse struct {
	ReviewCommentCount    int      `json:"reviewCommentCount"`
	ReviewRoundCount      int      `json:"reviewRoundCount"`
	ReviewerCount         int      `json:"reviewerCount"`
	ReviewersInvolved     []string `json:"reviewersInvolved"`
	CommitCount           int      `json:"commitCount"`
	FixupCommitCount      int      `json:"fixupCommitCount"`
	ForceUpdateCount      int      `json:"forceUpdateCount"`
	FirstReviewPassRate   float64  `json:"firstReviewPassRate"`
	AverageCommentPerFile float64  `json:"averageCommentPerFile"`
	ApprovalsReceived     int      `json:"approvalsReceived"`
	ApproversInvolved     []string `json:"approversInvolved"`
}

// PRAnalysisResultsResponse はPR分析結果のレスポンス
type PRAnalysisResultsResponse struct {
	IsHighQuality       bool                     `json:"isHighQuality"`
	IsLargePR           bool                     `json:"isLargePR"`
	HasLongReviewTime   bool                     `json:"hasLongReviewTime"`
	ReviewEfficiency    float64                  `json:"reviewEfficiency"`
	DominantFileType    string                   `json:"dominantFileType"`
	Recommendations     []string                 `json:"recommendations"`
	SplitSuggestions    []SplitSuggestionResponse `json:"splitSuggestions,omitempty"`
}

// SplitSuggestionResponse はPR分割提案のレスポンス
type SplitSuggestionResponse struct {
	Type        string              `json:"type"`
	Description string              `json:"description"`
	Groups      map[string][]string `json:"groups"`
}

// CycleTimeMetricsResponse はサイクルタイムメトリクスのレスポンス
type CycleTimeMetricsResponse struct {
	Period      string                   `json:"period"`
	StartDate   time.Time                `json:"startDate"`
	EndDate     time.Time                `json:"endDate"`
	TotalPRs    int                      `json:"totalPRs"`
	Statistics  CycleTimeStatsResponse   `json:"statistics"`
	Percentiles PercentilesResponse      `json:"percentiles"`
	Breakdown   CycleTimeBreakdownResponse `json:"breakdown"`
	Trends      TrendResponse            `json:"trends"`
}

// CycleTimeStatsResponse はサイクルタイム統計のレスポンス
type CycleTimeStatsResponse struct {
	Mean     *DurationResponse `json:"mean"`
	Median   *DurationResponse `json:"median"`
	Min      *DurationResponse `json:"min"`
	Max      *DurationResponse `json:"max"`
	StdDev   *DurationResponse `json:"stdDev"`
}

// PercentilesResponse はパーセンタイルのレスポンス
type PercentilesResponse struct {
	P25 *DurationResponse `json:"p25"`
	P50 *DurationResponse `json:"p50"`
	P75 *DurationResponse `json:"p75"`
	P90 *DurationResponse `json:"p90"`
	P95 *DurationResponse `json:"p95"`
	P99 *DurationResponse `json:"p99"`
}

// CycleTimeBreakdownResponse はサイクルタイム内訳のレスポンス
type CycleTimeBreakdownResponse struct {
	TimeToFirstReview CycleTimeStatsResponse `json:"timeToFirstReview"`
	TimeToApproval    CycleTimeStatsResponse `json:"timeToApproval"`
	TimeToMerge       CycleTimeStatsResponse `json:"timeToMerge"`
}

// ReviewTimeMetricsResponse はレビュー時間メトリクスのレスポンス
type ReviewTimeMetricsResponse struct {
	Period           string                     `json:"period"`
	StartDate        time.Time                  `json:"startDate"`
	EndDate          time.Time                  `json:"endDate"`
	TotalPRs         int                        `json:"totalPRs"`
	ReviewStatistics ReviewTimeStatsResponse    `json:"reviewStatistics"`
	EfficiencyMetrics ReviewEfficiencyResponse  `json:"efficiencyMetrics"`
	Bottlenecks      []BottleneckResponse       `json:"bottlenecks"`
	Trends           TrendResponse              `json:"trends"`
}

// ReviewTimeStatsResponse はレビュー時間統計のレスポンス
type ReviewTimeStatsResponse struct {
	AvgTimeToFirstReview *DurationResponse `json:"avgTimeToFirstReview"`
	AvgReviewWaitTime    *DurationResponse `json:"avgReviewWaitTime"`
	AvgReviewActiveTime  *DurationResponse `json:"avgReviewActiveTime"`
	AvgReviewRounds      float64           `json:"avgReviewRounds"`
	FirstPassRate        float64           `json:"firstPassRate"`
}

// ReviewEfficiencyResponse はレビュー効率のレスポンス
type ReviewEfficiencyResponse struct {
	OverallEfficiency     float64 `json:"overallEfficiency"`
	CommentEfficiency     float64 `json:"commentEfficiency"`
	RoundEfficiency       float64 `json:"roundEfficiency"`
	TimeEfficiency        float64 `json:"timeEfficiency"`
}

// BottleneckResponse はボトルネックのレスポンス
type BottleneckResponse struct {
	Type        string  `json:"type"`
	PRID        string  `json:"prId"`
	Severity    string  `json:"severity"`
	Description string  `json:"description"`
	Value       float64 `json:"value"`
	Suggestion  string  `json:"suggestion"`
}

// TrendResponse はトレンドのレスポンス
type TrendResponse struct {
	Direction   string  `json:"direction"` // "improving", "degrading", "stable"
	Slope       float64 `json:"slope"`
	Confidence  float64 `json:"confidence"`
	Description string  `json:"description"`
}

// PRListResponse はPRリストのレスポンス
type PRListResponse struct {
	PRs        []PRSummaryResponse `json:"prs"`
	TotalCount int                 `json:"totalCount"`
	Page       int                 `json:"page"`
	PageSize   int                 `json:"pageSize"`
	HasMore    bool                `json:"hasMore"`
}

// PRSummaryResponse はPR要約のレスポンス
type PRSummaryResponse struct {
	PRID            string    `json:"prId"`
	PRNumber        int       `json:"prNumber"`
	Title           string    `json:"title"`
	Author          string    `json:"author"`
	Repository      string    `json:"repository"`
	CreatedAt       time.Time `json:"createdAt"`
	MergedAt        *time.Time `json:"mergedAt,omitempty"`
	LinesChanged    int       `json:"linesChanged"`
	FilesChanged    int       `json:"filesChanged"`
	ComplexityScore float64   `json:"complexityScore"`
	SizeCategory    string    `json:"sizeCategory"`
	CycleTime       *DurationResponse `json:"cycleTime,omitempty"`
	ReviewTime      *DurationResponse `json:"reviewTime,omitempty"`
	IsHighQuality   bool      `json:"isHighQuality"`
}

// ErrorResponse はエラーレスポンス
type ErrorResponse struct {
	Error   string `json:"error"`
	Code    string `json:"code"`
	Message string `json:"message"`
	Details interface{} `json:"details,omitempty"`
}

// ValidationErrorResponse はバリデーションエラーのレスポンス
type ValidationErrorResponse struct {
	Error   string                   `json:"error"`
	Code    string                   `json:"code"`
	Message string                   `json:"message"`
	Fields  []FieldValidationError   `json:"fields"`
}

// FieldValidationError はフィールドバリデーションエラー
type FieldValidationError struct {
	Field   string `json:"field"`
	Value   interface{} `json:"value"`
	Message string `json:"message"`
}

// SuccessResponse は成功レスポンス
type SuccessResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// PaginationResponse はページネーションレスポンス
type PaginationResponse struct {
	Page       int  `json:"page"`
	PageSize   int  `json:"pageSize"`
	TotalCount int  `json:"totalCount"`
	TotalPages int  `json:"totalPages"`
	HasMore    bool `json:"hasMore"`
}

// MetaResponse はメタ情報のレスポンス
type MetaResponse struct {
	GeneratedAt    time.Time `json:"generatedAt"`
	ProcessingTime string    `json:"processingTime"`
	Version        string    `json:"version"`
	RequestID      string    `json:"requestId"`
}

// HealthCheckResponse はヘルスチェックのレスポンス
type HealthCheckResponse struct {
	Status      string            `json:"status"`
	Version     string            `json:"version"`
	Timestamp   time.Time         `json:"timestamp"`
	Services    map[string]string `json:"services"`
	Database    DatabaseStatus    `json:"database"`
	GitHubAPI   GitHubAPIStatus   `json:"githubApi"`
}

// DatabaseStatus はデータベースステータス
type DatabaseStatus struct {
	Connected   bool   `json:"connected"`
	LastChecked time.Time `json:"lastChecked"`
	ResponseTime string `json:"responseTime"`
}

// GitHubAPIStatus はGitHub APIステータス
type GitHubAPIStatus struct {
	Available    bool   `json:"available"`
	LastChecked  time.Time `json:"lastChecked"`
	ResponseTime string `json:"responseTime"`
	RateLimit    RateLimitStatus `json:"rateLimit"`
}

// RateLimitStatus はレート制限ステータス
type RateLimitStatus struct {
	Limit     int       `json:"limit"`
	Remaining int       `json:"remaining"`
	Reset     time.Time `json:"reset"`
}