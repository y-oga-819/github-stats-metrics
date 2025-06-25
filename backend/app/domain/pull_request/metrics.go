package pull_request

// PullRequestMetrics はPull Requestsのメトリクス情報
type PullRequestMetrics struct {
	// 平均レビュー時間（秒）
	AverageTimeToFirstReview float64 `json:"averageTimeToFirstReview"`
	
	// 平均承認時間（秒）
	AverageTimeToApproval float64 `json:"averageTimeToApproval"`
	
	// 平均マージ時間（秒）
	AverageTimeToMerge float64 `json:"averageTimeToMerge"`
	
	// 総Pull Request数
	TotalPullRequests int `json:"totalPullRequests"`
}

// ValidationDetails はバリデーションエラーの詳細情報
type ValidationDetails struct {
	InvalidDevelopers []string `json:"invalidDevelopers,omitempty"`
	MissingFields     []string `json:"missingFields,omitempty"`
	InvalidDateRange  bool     `json:"invalidDateRange,omitempty"`
}