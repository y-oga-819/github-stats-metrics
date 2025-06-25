package pull_request

import (
	"time"
)

// PRMetrics はPull Request個別の詳細メトリクス情報
type PRMetrics struct {
	// 基本情報
	PRID         string    `json:"prId"`
	PRNumber     int       `json:"prNumber"`
	Title        string    `json:"title"`
	Author       string    `json:"author"`
	Repository   string    `json:"repository"`
	CreatedAt    time.Time `json:"createdAt"`
	MergedAt     *time.Time `json:"mergedAt,omitempty"`

	// サイズメトリクス
	SizeMetrics PRSizeMetrics `json:"sizeMetrics"`

	// 時間メトリクス
	TimeMetrics PRTimeMetrics `json:"timeMetrics"`

	// 品質メトリクス
	QualityMetrics PRQualityMetrics `json:"qualityMetrics"`

	// 複雑度スコア
	ComplexityScore float64 `json:"complexityScore"`
	
	// PRサイズ分類
	SizeCategory PRSizeCategory `json:"sizeCategory"`
}

// PRSizeMetrics はPRのサイズ関連メトリクス
type PRSizeMetrics struct {
	// 行数関連
	LinesAdded   int `json:"linesAdded"`
	LinesDeleted int `json:"linesDeleted"`
	LinesChanged int `json:"linesChanged"`
	
	// ファイル関連
	FilesChanged     int                    `json:"filesChanged"`
	FileTypeBreakdown map[string]int        `json:"fileTypeBreakdown"` // 拡張子別ファイル数
	DirectoryCount   int                    `json:"directoryCount"`
	
	// 変更の詳細
	FileChanges []FileChangeMetrics `json:"fileChanges"`
}

// FileChangeMetrics は個別ファイルの変更メトリクス
type FileChangeMetrics struct {
	FileName     string `json:"fileName"`
	FileType     string `json:"fileType"`
	LinesAdded   int    `json:"linesAdded"`
	LinesDeleted int    `json:"linesDeleted"`
	IsNewFile    bool   `json:"isNewFile"`
	IsDeleted    bool   `json:"isDeleted"`
	IsRenamed    bool   `json:"isRenamed"`
}

// PRTimeMetrics はPRの時間関連メトリクス
type PRTimeMetrics struct {
	// サイクルタイム（全体）
	TotalCycleTime *time.Duration `json:"totalCycleTime,omitempty"`
	
	// 各段階の時間
	TimeToFirstReview *time.Duration `json:"timeToFirstReview,omitempty"`
	TimeToApproval    *time.Duration `json:"timeToApproval,omitempty"`
	TimeToMerge       *time.Duration `json:"timeToMerge,omitempty"`
	
	// レビュー関連時間
	ReviewWaitTime    *time.Duration `json:"reviewWaitTime,omitempty"`
	ReviewActiveTime  *time.Duration `json:"reviewActiveTime,omitempty"`
	
	// その他の時間
	FirstCommitToMerge *time.Duration `json:"firstCommitToMerge,omitempty"`
	
	// 時間帯分析
	CreatedHour int `json:"createdHour"` // 作成時刻（0-23）
	MergedHour  *int `json:"mergedHour,omitempty"` // マージ時刻（0-23）
}

// PRQualityMetrics はPRの品質関連メトリクス
type PRQualityMetrics struct {
	// レビュー関連
	ReviewCommentCount    int `json:"reviewCommentCount"`
	ReviewRoundCount      int `json:"reviewRoundCount"`
	ReviewerCount         int `json:"reviewerCount"`
	ReviewersInvolved     []string `json:"reviewersInvolved"`
	
	// 修正関連
	CommitCount           int `json:"commitCount"`
	FixupCommitCount      int `json:"fixupCommitCount"`
	ForceUpdateCount      int `json:"forceUpdateCount"`
	
	// レビュー効率
	FirstReviewPassRate   float64 `json:"firstReviewPassRate"` // 初回レビュー通過率
	AverageCommentPerFile float64 `json:"averageCommentPerFile"`
	
	// 承認関連
	ApprovalsReceived     int      `json:"approvalsReceived"`
	ApproversInvolved     []string `json:"approversInvolved"`
}

// PRSizeCategory はPRのサイズ分類
type PRSizeCategory string

const (
	PRSizeXSmall PRSizeCategory = "XS" // ~50行
	PRSizeSmall  PRSizeCategory = "S"  // 51-100行
	PRSizeMedium PRSizeCategory = "M"  // 101-300行
	PRSizeLarge  PRSizeCategory = "L"  // 301-600行
	PRSizeXLarge PRSizeCategory = "XL" // 601行以上
)

// ドメインロジック: PRメトリクス計算

// CalculateSizeCategory はPRサイズカテゴリを計算
func (m *PRMetrics) CalculateSizeCategory() PRSizeCategory {
	totalLines := m.SizeMetrics.LinesChanged
	
	switch {
	case totalLines <= 50:
		return PRSizeXSmall
	case totalLines <= 100:
		return PRSizeSmall
	case totalLines <= 300:
		return PRSizeMedium
	case totalLines <= 600:
		return PRSizeLarge
	default:
		return PRSizeXLarge
	}
}

// IsLargePR は大きすぎるPRかを判定
func (m *PRMetrics) IsLargePR() bool {
	return m.SizeCategory == PRSizeLarge || m.SizeCategory == PRSizeXLarge
}

// CalculateReviewEfficiency はレビュー効率を計算
func (m *PRMetrics) CalculateReviewEfficiency() float64 {
	if m.QualityMetrics.ReviewCommentCount == 0 {
		return 1.0 // コメントなし = 100%効率
	}
	
	// レビューラウンド数が少ないほど効率が良い
	efficiency := 1.0 / float64(m.QualityMetrics.ReviewRoundCount)
	if efficiency > 1.0 {
		efficiency = 1.0
	}
	
	return efficiency
}

// HasLongReviewTime はレビュー時間が長いかを判定
func (m *PRMetrics) HasLongReviewTime(threshold time.Duration) bool {
	if m.TimeMetrics.TimeToFirstReview == nil {
		return false
	}
	return *m.TimeMetrics.TimeToFirstReview > threshold
}

// IsHighQuality は高品質PRかを判定
func (m *PRMetrics) IsHighQuality() bool {
	// レビューコメント数が少なく、修正回数も少ない
	return m.QualityMetrics.ReviewCommentCount <= 3 && 
		   m.QualityMetrics.ReviewRoundCount <= 2 &&
		   m.QualityMetrics.FirstReviewPassRate >= 0.8
}

// GetDominantFileType は最も変更が多いファイルタイプを取得
func (m *PRMetrics) GetDominantFileType() string {
	if len(m.SizeMetrics.FileTypeBreakdown) == 0 {
		return "unknown"
	}
	
	maxCount := 0
	dominantType := ""
	
	for fileType, count := range m.SizeMetrics.FileTypeBreakdown {
		if count > maxCount {
			maxCount = count
			dominantType = fileType
		}
	}
	
	return dominantType
}

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