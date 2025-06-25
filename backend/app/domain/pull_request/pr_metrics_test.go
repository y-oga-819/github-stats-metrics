package pull_request

import (
	"testing"
	"time"
)

func TestPRMetrics_CalculateSizeCategory(t *testing.T) {
	tests := []struct {
		name         string
		linesChanged int
		expected     PRSizeCategory
	}{
		{"XS - 50行以下", 30, PRSizeXSmall},
		{"XS - 境界値50行", 50, PRSizeXSmall},
		{"S - 51行", 51, PRSizeSmall},
		{"S - 100行", 100, PRSizeSmall},
		{"M - 101行", 101, PRSizeMedium},
		{"M - 300行", 300, PRSizeMedium},
		{"L - 301行", 301, PRSizeLarge},
		{"L - 600行", 600, PRSizeLarge},
		{"XL - 601行以上", 601, PRSizeXLarge},
		{"XL - 大きなPR", 1000, PRSizeXLarge},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			metrics := &PRMetrics{
				SizeMetrics: PRSizeMetrics{
					LinesChanged: tt.linesChanged,
				},
			}
			
			result := metrics.CalculateSizeCategory()
			if result != tt.expected {
				t.Errorf("CalculateSizeCategory() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestPRMetrics_IsLargePR(t *testing.T) {
	tests := []struct {
		name         string
		sizeCategory PRSizeCategory
		expected     bool
	}{
		{"XS は大きくない", PRSizeXSmall, false},
		{"S は大きくない", PRSizeSmall, false},
		{"M は大きくない", PRSizeMedium, false},
		{"L は大きい", PRSizeLarge, true},
		{"XL は大きい", PRSizeXLarge, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			metrics := &PRMetrics{
				SizeCategory: tt.sizeCategory,
			}
			
			result := metrics.IsLargePR()
			if result != tt.expected {
				t.Errorf("IsLargePR() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestPRMetrics_CalculateReviewEfficiency(t *testing.T) {
	tests := []struct {
		name               string
		reviewCommentCount int
		reviewRoundCount   int
		expected           float64
	}{
		{"コメントなし", 0, 1, 1.0},
		{"1ラウンド", 5, 1, 1.0},
		{"2ラウンド", 5, 2, 0.5},
		{"3ラウンド", 10, 3, 0.333333},
		{"5ラウンド", 20, 5, 0.2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			metrics := &PRMetrics{
				QualityMetrics: PRQualityMetrics{
					ReviewCommentCount: tt.reviewCommentCount,
					ReviewRoundCount:   tt.reviewRoundCount,
				},
			}
			
			result := metrics.CalculateReviewEfficiency()
			
			// 小数点以下2桁で比較
			if abs(result-tt.expected) > 0.01 {
				t.Errorf("CalculateReviewEfficiency() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestPRMetrics_HasLongReviewTime(t *testing.T) {
	threshold := 24 * time.Hour
	
	tests := []struct {
		name             string
		timeToFirstReview *time.Duration
		expected         bool
	}{
		{"レビュー時間なし", nil, false},
		{"短いレビュー時間", durationPtr(2 * time.Hour), false},
		{"境界値", durationPtr(24 * time.Hour), false},
		{"長いレビュー時間", durationPtr(25 * time.Hour), true},
		{"非常に長いレビュー時間", durationPtr(72 * time.Hour), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			metrics := &PRMetrics{
				TimeMetrics: PRTimeMetrics{
					TimeToFirstReview: tt.timeToFirstReview,
				},
			}
			
			result := metrics.HasLongReviewTime(threshold)
			if result != tt.expected {
				t.Errorf("HasLongReviewTime() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestPRMetrics_IsHighQuality(t *testing.T) {
	tests := []struct {
		name                string
		reviewCommentCount  int
		reviewRoundCount    int
		firstReviewPassRate float64
		expected            bool
	}{
		{"高品質 - 全条件満足", 2, 1, 0.9, true},
		{"高品質 - 境界値", 3, 2, 0.8, true},
		{"低品質 - コメント多い", 5, 1, 0.9, false},
		{"低品質 - ラウンド多い", 2, 3, 0.9, false},
		{"低品質 - 通過率低い", 2, 1, 0.7, false},
		{"低品質 - 全条件悪い", 10, 5, 0.3, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			metrics := &PRMetrics{
				QualityMetrics: PRQualityMetrics{
					ReviewCommentCount:  tt.reviewCommentCount,
					ReviewRoundCount:    tt.reviewRoundCount,
					FirstReviewPassRate: tt.firstReviewPassRate,
				},
			}
			
			result := metrics.IsHighQuality()
			if result != tt.expected {
				t.Errorf("IsHighQuality() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestPRMetrics_GetDominantFileType(t *testing.T) {
	tests := []struct {
		name             string
		fileTypeBreakdown map[string]int
		expected         string
	}{
		{"空のBreakdown", map[string]int{}, "unknown"},
		{"単一ファイルタイプ", map[string]int{".go": 5}, ".go"},
		{"複数ファイルタイプ", map[string]int{".go": 3, ".js": 7, ".md": 1}, ".js"},
		{"同数の場合", map[string]int{".go": 3, ".js": 3}, ".go"}, // mapの順序に依存
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			metrics := &PRMetrics{
				SizeMetrics: PRSizeMetrics{
					FileTypeBreakdown: tt.fileTypeBreakdown,
				},
			}
			
			result := metrics.GetDominantFileType()
			
			// 同数の場合はmapの順序に依存するため、有効な値かチェック
			if tt.name == "同数の場合" {
				if result != ".go" && result != ".js" {
					t.Errorf("GetDominantFileType() = %v, want either .go or .js", result)
				}
			} else if result != tt.expected {
				t.Errorf("GetDominantFileType() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestFileChangeMetrics_Validation(t *testing.T) {
	tests := []struct {
		name     string
		metrics  FileChangeMetrics
		isValid  bool
	}{
		{
			"正常なファイル変更",
			FileChangeMetrics{
				FileName:     "main.go",
				FileType:     ".go",
				LinesAdded:   10,
				LinesDeleted: 5,
				IsNewFile:    false,
				IsDeleted:    false,
				IsRenamed:    false,
			},
			true,
		},
		{
			"新規ファイル",
			FileChangeMetrics{
				FileName:     "new.go",
				FileType:     ".go",
				LinesAdded:   20,
				LinesDeleted: 0,
				IsNewFile:    true,
				IsDeleted:    false,
				IsRenamed:    false,
			},
			true,
		},
		{
			"削除ファイル",
			FileChangeMetrics{
				FileName:     "old.go",
				FileType:     ".go",
				LinesAdded:   0,
				LinesDeleted: 30,
				IsNewFile:    false,
				IsDeleted:    true,
				IsRenamed:    false,
			},
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 基本的な整合性チェック
			if tt.metrics.IsNewFile && tt.metrics.LinesDeleted > 0 {
				t.Error("新規ファイルなのに削除行数がある")
			}
			if tt.metrics.IsDeleted && tt.metrics.LinesAdded > 0 {
				t.Error("削除ファイルなのに追加行数がある")
			}
		})
	}
}

// ヘルパー関数

// durationPtr は time.Duration のポインターを作成
func durationPtr(d time.Duration) *time.Duration {
	return &d
}

// abs は浮動小数点数の絶対値を計算
func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

// createTestPRMetrics はテスト用のPRMetricsを作成
func createTestPRMetrics() *PRMetrics {
	now := time.Now()
	mergedTime := now.Add(24 * time.Hour)
	
	return &PRMetrics{
		PRID:         "pr-123",
		PRNumber:     123,
		Title:        "Test PR",
		Author:       "test-user",
		Repository:   "test-repo",
		CreatedAt:    now,
		MergedAt:     &mergedTime,
		SizeMetrics: PRSizeMetrics{
			LinesAdded:   100,
			LinesDeleted: 50,
			LinesChanged: 150,
			FilesChanged: 5,
			FileTypeBreakdown: map[string]int{
				".go": 3,
				".js": 2,
			},
			DirectoryCount: 2,
			FileChanges: []FileChangeMetrics{
				{
					FileName:     "main.go",
					FileType:     ".go",
					LinesAdded:   50,
					LinesDeleted: 10,
					IsNewFile:    false,
					IsDeleted:    false,
					IsRenamed:    false,
				},
				{
					FileName:     "app.js",
					FileType:     ".js",
					LinesAdded:   30,
					LinesDeleted: 20,
					IsNewFile:    false,
					IsDeleted:    false,
					IsRenamed:    false,
				},
			},
		},
		TimeMetrics: PRTimeMetrics{
			TotalCycleTime:    durationPtr(24 * time.Hour),
			TimeToFirstReview: durationPtr(2 * time.Hour),
			TimeToApproval:    durationPtr(4 * time.Hour),
			TimeToMerge:       durationPtr(1 * time.Hour),
			CreatedHour:       9,
			MergedHour:        intPtr(10),
		},
		QualityMetrics: PRQualityMetrics{
			ReviewCommentCount:    5,
			ReviewRoundCount:      2,
			ReviewerCount:         3,
			ReviewersInvolved:     []string{"reviewer1", "reviewer2", "reviewer3"},
			CommitCount:           8,
			FixupCommitCount:      1,
			ForceUpdateCount:      0,
			FirstReviewPassRate:   0.8,
			AverageCommentPerFile: 1.0,
			ApprovalsReceived:     2,
			ApproversInvolved:     []string{"approver1", "approver2"},
		},
		ComplexityScore: 2.5,
		SizeCategory:    PRSizeMedium,
	}
}

// intPtr は int のポインターを作成
func intPtr(i int) *int {
	return &i
}

// 統合テスト

func TestPRMetrics_IntegrationTest(t *testing.T) {
	metrics := createTestPRMetrics()
	
	// サイズカテゴリ計算
	sizeCategory := metrics.CalculateSizeCategory()
	if sizeCategory != PRSizeMedium {
		t.Errorf("Expected size category M, got %v", sizeCategory)
	}
	
	// 大きなPR判定
	if metrics.IsLargePR() {
		t.Error("Expected not to be large PR")
	}
	
	// レビュー効率計算
	efficiency := metrics.CalculateReviewEfficiency()
	if efficiency != 0.5 { // 2ラウンド = 1/2 = 0.5
		t.Errorf("Expected review efficiency 0.5, got %v", efficiency)
	}
	
	// 高品質判定 (コメント数5、ラウンド数2のため高品質ではない)
	if metrics.IsHighQuality() {
		t.Error("Expected not to be high quality PR with 5 comments and 2 rounds")
	}
	
	// 主要ファイルタイプ
	dominantType := metrics.GetDominantFileType()
	if dominantType != ".go" {
		t.Errorf("Expected dominant file type .go, got %v", dominantType)
	}
	
	// 長いレビュー時間判定 (TimeToFirstReview = 2時間)
	if !metrics.HasLongReviewTime(1 * time.Hour) {
		t.Error("Expected to have long review time with 2 hour threshold")
	}
	if metrics.HasLongReviewTime(3 * time.Hour) {
		t.Error("Expected not to have long review time with 3 hour threshold")
	}
}