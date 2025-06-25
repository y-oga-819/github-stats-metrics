package pull_request

import (
	"testing"
)

func TestPRComplexityAnalyzer_AnalyzeComplexity(t *testing.T) {
	analyzer := NewPRComplexityAnalyzer()
	
	tests := []struct {
		name        string
		metrics     *PRMetrics
		expectedMin float64
		expectedMax float64
	}{
		{
			"シンプルなPR",
			createSimplePRMetrics(),
			0.5,
			2.0,
		},
		{
			"複雑なPR",
			createComplexPRMetrics(),
			3.0,
			7.0,
		},
		{
			"大きなPR",
			createLargePRMetrics(),
			4.0,
			10.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := analyzer.AnalyzeComplexity(tt.metrics)
			
			if result < tt.expectedMin || result > tt.expectedMax {
				t.Errorf("AnalyzeComplexity() = %v, want between %v and %v", 
					result, tt.expectedMin, tt.expectedMax)
			}
			
			// 複雑度は0.1-10.0の範囲内であること
			if result < 0.1 || result > 10.0 {
				t.Errorf("Complexity score %v is out of valid range [0.1, 10.0]", result)
			}
		})
	}
}

func TestPRComplexityAnalyzer_GetComplexityLevel(t *testing.T) {
	analyzer := NewPRComplexityAnalyzer()
	
	tests := []struct {
		name     string
		score    float64
		expected ComplexityLevel
	}{
		{"Very Low", 1.0, ComplexityLevelVeryLow},
		{"Very Low boundary", 1.5, ComplexityLevelVeryLow},
		{"Low", 2.0, ComplexityLevelLow},
		{"Low boundary", 2.5, ComplexityLevelLow},
		{"Medium", 3.0, ComplexityLevelMedium},
		{"Medium boundary", 4.0, ComplexityLevelMedium},
		{"High", 5.0, ComplexityLevelHigh},
		{"High boundary", 6.0, ComplexityLevelHigh},
		{"Very High", 7.0, ComplexityLevelVeryHigh},
		{"Very High max", 10.0, ComplexityLevelVeryHigh},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := analyzer.GetComplexityLevel(tt.score)
			if result != tt.expected {
				t.Errorf("GetComplexityLevel(%v) = %v, want %v", tt.score, result, tt.expected)
			}
		})
	}
}

func TestPRComplexityAnalyzer_AnalyzeFileComplexity(t *testing.T) {
	analyzer := NewPRComplexityAnalyzer()
	
	tests := []struct {
		name        string
		fileChange  FileChangeMetrics
		expectedMin float64
		expectedMax float64
	}{
		{
			"小さなGoファイル",
			FileChangeMetrics{
				FileName:     "main.go",
				FileType:     ".go",
				LinesAdded:   10,
				LinesDeleted: 5,
				IsNewFile:    false,
				IsDeleted:    false,
				IsRenamed:    false,
			},
			0.5,
			2.0,
		},
		{
			"新規ファイル",
			FileChangeMetrics{
				FileName:     "new.go",
				FileType:     ".go",
				LinesAdded:   100,
				LinesDeleted: 0,
				IsNewFile:    true,
				IsDeleted:    false,
				IsRenamed:    false,
			},
			1.5,
			4.0,
		},
		{
			"削除ファイル",
			FileChangeMetrics{
				FileName:     "old.js",
				FileType:     ".js",
				LinesAdded:   0,
				LinesDeleted: 50,
				IsNewFile:    false,
				IsDeleted:    true,
				IsRenamed:    false,
			},
			0.3,
			2.0,
		},
		{
			"リネームファイル",
			FileChangeMetrics{
				FileName:     "renamed.py",
				FileType:     ".py",
				LinesAdded:   20,
				LinesDeleted: 15,
				IsNewFile:    false,
				IsDeleted:    false,
				IsRenamed:    true,
			},
			0.5,
			2.5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := analyzer.AnalyzeFileComplexity(tt.fileChange)
			
			if result < tt.expectedMin || result > tt.expectedMax {
				t.Errorf("AnalyzeFileComplexity() = %v, want between %v and %v", 
					result, tt.expectedMin, tt.expectedMax)
			}
		})
	}
}

func TestPRComplexityAnalyzer_SuggestOptimalSplit(t *testing.T) {
	analyzer := NewPRComplexityAnalyzer()
	
	tests := []struct {
		name           string
		metrics        *PRMetrics
		expectSuggestions bool
	}{
		{
			"小さなPR - 分割提案なし",
			createSimplePRMetrics(),
			false,
		},
		{
			"大きなPR - 分割提案あり",
			createLargePRMetrics(),
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			suggestions := analyzer.SuggestOptimalSplit(tt.metrics)
			
			hasSuggestions := len(suggestions) > 0
			if hasSuggestions != tt.expectSuggestions {
				t.Errorf("SuggestOptimalSplit() returned %d suggestions, expected suggestions: %v", 
					len(suggestions), tt.expectSuggestions)
			}
			
			// 分割提案がある場合の内容チェック
			if hasSuggestions {
				for _, suggestion := range suggestions {
					if suggestion.Type == "" {
						t.Error("Split suggestion has empty type")
					}
					if suggestion.Description == "" {
						t.Error("Split suggestion has empty description")
					}
					if len(suggestion.Groups) == 0 {
						t.Error("Split suggestion has no groups")
					}
				}
			}
		})
	}
}

func TestPRComplexityAnalyzer_getFileTypeWeight(t *testing.T) {
	analyzer := NewPRComplexityAnalyzer()
	
	tests := []struct {
		name         string
		fileType     string
		expectedMin  float64
		expectedMax  float64
	}{
		{"Go file", ".go", 1.15, 1.25},
		{"Go file without dot", "go", 1.15, 1.25},
		{"JavaScript file", ".js", 0.95, 1.05},
		{"Markdown file", ".md", 0.35, 0.45},
		{"Unknown file type", ".unknown", 0.95, 1.05},
		{"Empty file type", "", 0.95, 1.05},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := analyzer.getFileTypeWeight(tt.fileType)
			
			if result < tt.expectedMin || result > tt.expectedMax {
				t.Errorf("getFileTypeWeight(%s) = %v, want between %v and %v", 
					tt.fileType, result, tt.expectedMin, tt.expectedMax)
			}
		})
	}
}

func TestPRComplexityAnalyzer_calculateBaseComplexity(t *testing.T) {
	analyzer := NewPRComplexityAnalyzer()
	
	tests := []struct {
		name         string
		linesChanged int
		filesChanged int
		expectedMin  float64
		expectedMax  float64
	}{
		{"Small change", 10, 1, 0.5, 1.5},
		{"Medium change", 100, 5, 1.0, 2.5},
		{"Large change", 1000, 20, 2.0, 4.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			metrics := &PRMetrics{
				SizeMetrics: PRSizeMetrics{
					LinesChanged: tt.linesChanged,
					FilesChanged: tt.filesChanged,
				},
			}
			
			result := analyzer.calculateBaseComplexity(metrics)
			
			if result < tt.expectedMin || result > tt.expectedMax {
				t.Errorf("calculateBaseComplexity() = %v, want between %v and %v", 
					result, tt.expectedMin, tt.expectedMax)
			}
		})
	}
}

func TestPRComplexityAnalyzer_calculateSizeComplexity(t *testing.T) {
	analyzer := NewPRComplexityAnalyzer()
	
	tests := []struct {
		name         string
		linesChanged int
		expected     float64
	}{
		{"XS size", 30, 0.8},
		{"S size", 80, 0.9},
		{"M size", 200, 1.0},
		{"L size", 500, 1.2},
		{"XL size", 800, 1.5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			metrics := &PRMetrics{
				SizeMetrics: PRSizeMetrics{
					LinesChanged: tt.linesChanged,
				},
			}
			
			result := analyzer.calculateSizeComplexity(metrics)
			
			if result != tt.expected {
				t.Errorf("calculateSizeComplexity() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestComplexityConfig_DefaultValues(t *testing.T) {
	config := getDefaultComplexityConfig()
	
	// 主要なファイルタイプの重みをチェック
	if weight, exists := config.FileTypeWeights[".go"]; !exists || weight != 1.2 {
		t.Errorf("Expected .go weight to be 1.2, got %v", weight)
	}
	
	if weight, exists := config.FileTypeWeights[".js"]; !exists || weight != 1.0 {
		t.Errorf("Expected .js weight to be 1.0, got %v", weight)
	}
	
	if weight, exists := config.FileTypeWeights[".md"]; !exists || weight != 0.4 {
		t.Errorf("Expected .md weight to be 0.4, got %v", weight)
	}
	
	// サイズ乗数をチェック
	if multiplier, exists := config.LineSizeMultipliers[PRSizeMedium]; !exists || multiplier != 1.0 {
		t.Errorf("Expected medium size multiplier to be 1.0, got %v", multiplier)
	}
	
	// その他の設定値をチェック
	if config.NewFileWeight != 1.3 {
		t.Errorf("Expected NewFileWeight to be 1.3, got %v", config.NewFileWeight)
	}
}

// テスト用のヘルパー関数

func createSimplePRMetrics() *PRMetrics {
	return &PRMetrics{
		SizeMetrics: PRSizeMetrics{
			LinesAdded:   30,
			LinesDeleted: 20,
			LinesChanged: 50,
			FilesChanged: 2,
			FileTypeBreakdown: map[string]int{
				".js": 2,
			},
			DirectoryCount: 1,
			FileChanges: []FileChangeMetrics{
				{
					FileName:     "app.js",
					FileType:     ".js",
					LinesAdded:   20,
					LinesDeleted: 10,
					IsNewFile:    false,
					IsDeleted:    false,
					IsRenamed:    false,
				},
				{
					FileName:     "style.css",
					FileType:     ".css",
					LinesAdded:   10,
					LinesDeleted: 10,
					IsNewFile:    false,
					IsDeleted:    false,
					IsRenamed:    false,
				},
			},
		},
		QualityMetrics: PRQualityMetrics{
			ReviewRoundCount:   1,
			ReviewCommentCount: 2,
		},
		SizeCategory: PRSizeXSmall,
	}
}

func createComplexPRMetrics() *PRMetrics {
	return &PRMetrics{
		SizeMetrics: PRSizeMetrics{
			LinesAdded:   200,
			LinesDeleted: 100,
			LinesChanged: 300,
			FilesChanged: 8,
			FileTypeBreakdown: map[string]int{
				".go":  3,
				".cpp": 2,
				".js":  2,
				".md":  1,
			},
			DirectoryCount: 3,
			FileChanges: []FileChangeMetrics{
				{
					FileName:     "main.go",
					FileType:     ".go",
					LinesAdded:   80,
					LinesDeleted: 30,
					IsNewFile:    false,
					IsDeleted:    false,
					IsRenamed:    false,
				},
				{
					FileName:     "new_service.go",
					FileType:     ".go",
					LinesAdded:   60,
					LinesDeleted: 0,
					IsNewFile:    true,
					IsDeleted:    false,
					IsRenamed:    false,
				},
				{
					FileName:     "algorithm.cpp",
					FileType:     ".cpp",
					LinesAdded:   40,
					LinesDeleted: 20,
					IsNewFile:    false,
					IsDeleted:    false,
					IsRenamed:    false,
				},
			},
		},
		QualityMetrics: PRQualityMetrics{
			ReviewRoundCount:   3,
			ReviewCommentCount: 15,
		},
		SizeCategory: PRSizeMedium,
	}
}

func createLargePRMetrics() *PRMetrics {
	return &PRMetrics{
		SizeMetrics: PRSizeMetrics{
			LinesAdded:   400,
			LinesDeleted: 200,
			LinesChanged: 600,
			FilesChanged: 15,
			FileTypeBreakdown: map[string]int{
				".go":   5,
				".java": 3,
				".cpp":  3,
				".js":   2,
				".py":   2,
			},
			DirectoryCount: 5,
			FileChanges: []FileChangeMetrics{
				{
					FileName:     "backend/service/user.go",
					FileType:     ".go",
					LinesAdded:   100,
					LinesDeleted: 50,
					IsNewFile:    false,
					IsDeleted:    false,
					IsRenamed:    false,
				},
				{
					FileName:     "frontend/components/UserList.js",
					FileType:     ".js",
					LinesAdded:   80,
					LinesDeleted: 30,
					IsNewFile:    false,
					IsDeleted:    false,
					IsRenamed:    false,
				},
				{
					FileName:     "core/algorithm/sort.cpp",
					FileType:     ".cpp",
					LinesAdded:   120,
					LinesDeleted: 60,
					IsNewFile:    true,
					IsDeleted:    false,
					IsRenamed:    false,
				},
				{
					FileName:     "scripts/migration.py",
					FileType:     ".py",
					LinesAdded:   100,
					LinesDeleted: 60,
					IsNewFile:    false,
					IsDeleted:    false,
					IsRenamed:    true,
				},
			},
		},
		QualityMetrics: PRQualityMetrics{
			ReviewRoundCount:   4,
			ReviewCommentCount: 25,
		},
		SizeCategory: PRSizeLarge,
	}
}