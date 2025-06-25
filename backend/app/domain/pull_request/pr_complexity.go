package pull_request

import (
	"math"
	"path/filepath"
	"strings"
)

// PRComplexityAnalyzer はPRの複雑度を分析するオブジェクト
type PRComplexityAnalyzer struct {
	// 設定可能な重み付けファクター
	config ComplexityConfig
}

// ComplexityConfig は複雑度計算の設定
type ComplexityConfig struct {
	// ファイルタイプ別の重み（1.0が基準）
	FileTypeWeights map[string]float64
	
	// サイズによる重み調整
	LineSizeMultipliers map[PRSizeCategory]float64
	
	// その他の重み付け要素
	NewFileWeight      float64 // 新規ファイル作成の重み
	DeleteFileWeight   float64 // ファイル削除の重み
	RenameFileWeight   float64 // ファイル名変更の重み
	DirectoryWeight    float64 // ディレクトリ跨ぎの重み
	
	// レビュー履歴による調整
	ReviewRoundPenalty float64 // レビューラウンド数による複雑度増加
	CommentDensityWeight float64 // コメント密度による重み
}

// NewPRComplexityAnalyzer は新しい複雑度分析器を作成
func NewPRComplexityAnalyzer() *PRComplexityAnalyzer {
	return &PRComplexityAnalyzer{
		config: getDefaultComplexityConfig(),
	}
}

// getDefaultComplexityConfig はデフォルトの複雑度設定を返す
func getDefaultComplexityConfig() ComplexityConfig {
	return ComplexityConfig{
		FileTypeWeights: map[string]float64{
			// 高複雑度ファイル
			".go":     1.2,
			".java":   1.2,
			".cpp":    1.3,
			".c":      1.3,
			".rs":     1.2,
			".py":     1.1,
			".ts":     1.1,
			".js":     1.0,
			
			// 中複雑度ファイル
			".tsx":    1.0,
			".jsx":    1.0,
			".php":    1.0,
			".rb":     1.0,
			".swift":  1.1,
			".kt":     1.1,
			
			// 低複雑度ファイル
			".html":   0.7,
			".css":    0.6,
			".scss":   0.7,
			".less":   0.7,
			".json":   0.5,
			".xml":    0.6,
			".yaml":   0.5,
			".yml":    0.5,
			".md":     0.4,
			".txt":    0.3,
			
			// 設定ファイル
			".toml":   0.5,
			".ini":    0.4,
			".conf":   0.4,
			
			// その他
			"":        1.0, // 拡張子なし
		},
		LineSizeMultipliers: map[PRSizeCategory]float64{
			PRSizeXSmall: 0.8,
			PRSizeSmall:  0.9,
			PRSizeMedium: 1.0,
			PRSizeLarge:  1.2,
			PRSizeXLarge: 1.5,
		},
		NewFileWeight:        1.3,
		DeleteFileWeight:     0.8,
		RenameFileWeight:     0.6,
		DirectoryWeight:      0.1,
		ReviewRoundPenalty:   0.15,
		CommentDensityWeight: 0.1,
	}
}

// AnalyzeComplexity はPRの複雑度を総合的に分析
func (analyzer *PRComplexityAnalyzer) AnalyzeComplexity(metrics *PRMetrics) float64 {
	// 基本複雑度の計算
	baseComplexity := analyzer.calculateBaseComplexity(metrics)
	
	// ファイル種別による調整
	fileTypeComplexity := analyzer.calculateFileTypeComplexity(metrics)
	
	// サイズによる調整
	sizeComplexity := analyzer.calculateSizeComplexity(metrics)
	
	// 構造的複雑度（ディレクトリ、新規ファイルなど）
	structuralComplexity := analyzer.calculateStructuralComplexity(metrics)
	
	// レビュー履歴による複雑度調整
	reviewComplexity := analyzer.calculateReviewComplexity(metrics)
	
	// 総合複雑度の計算（重み付け平均）
	totalComplexity := (baseComplexity*0.3 + 
					   fileTypeComplexity*0.25 + 
					   sizeComplexity*0.2 + 
					   structuralComplexity*0.15 + 
					   reviewComplexity*0.1)
	
	// 最終調整（0.1 - 10.0の範囲に正規化）
	return math.Max(0.1, math.Min(10.0, totalComplexity))
}

// calculateBaseComplexity は基本的な複雑度を計算
func (analyzer *PRComplexityAnalyzer) calculateBaseComplexity(metrics *PRMetrics) float64 {
	linesChanged := float64(metrics.SizeMetrics.LinesChanged)
	filesChanged := float64(metrics.SizeMetrics.FilesChanged)
	
	// 行数ベースの複雑度（対数スケール）
	lineComplexity := math.Log10(linesChanged+1) * 0.5
	
	// ファイル数ベースの複雑度
	fileComplexity := math.Sqrt(filesChanged) * 0.3
	
	return lineComplexity + fileComplexity
}

// calculateFileTypeComplexity はファイルタイプ別複雑度を計算
func (analyzer *PRComplexityAnalyzer) calculateFileTypeComplexity(metrics *PRMetrics) float64 {
	totalWeight := 0.0
	totalLines := 0
	
	for _, fileChange := range metrics.SizeMetrics.FileChanges {
		weight := analyzer.getFileTypeWeight(fileChange.FileType)
		lines := fileChange.LinesAdded + fileChange.LinesDeleted
		
		totalWeight += weight * float64(lines)
		totalLines += lines
	}
	
	if totalLines == 0 {
		return 1.0
	}
	
	// 重み付け平均
	return totalWeight / float64(totalLines)
}

// calculateSizeComplexity はサイズによる複雑度を計算
func (analyzer *PRComplexityAnalyzer) calculateSizeComplexity(metrics *PRMetrics) float64 {
	sizeCategory := metrics.CalculateSizeCategory()
	multiplier, exists := analyzer.config.LineSizeMultipliers[sizeCategory]
	if !exists {
		multiplier = 1.0
	}
	
	return multiplier
}

// calculateStructuralComplexity は構造的複雑度を計算
func (analyzer *PRComplexityAnalyzer) calculateStructuralComplexity(metrics *PRMetrics) float64 {
	complexity := 1.0
	
	// 新規ファイル、削除、リネームの複雑度
	for _, fileChange := range metrics.SizeMetrics.FileChanges {
		if fileChange.IsNewFile {
			complexity += analyzer.config.NewFileWeight * 0.1
		}
		if fileChange.IsDeleted {
			complexity += analyzer.config.DeleteFileWeight * 0.1
		}
		if fileChange.IsRenamed {
			complexity += analyzer.config.RenameFileWeight * 0.1
		}
	}
	
	// ディレクトリ数による複雑度
	complexity += float64(metrics.SizeMetrics.DirectoryCount) * analyzer.config.DirectoryWeight
	
	return complexity
}

// calculateReviewComplexity はレビュー履歴による複雑度を計算
func (analyzer *PRComplexityAnalyzer) calculateReviewComplexity(metrics *PRMetrics) float64 {
	complexity := 1.0
	
	// レビューラウンド数による複雑度増加
	if metrics.QualityMetrics.ReviewRoundCount > 1 {
		complexity += float64(metrics.QualityMetrics.ReviewRoundCount-1) * analyzer.config.ReviewRoundPenalty
	}
	
	// コメント密度による複雑度
	if metrics.SizeMetrics.FilesChanged > 0 {
		commentDensity := float64(metrics.QualityMetrics.ReviewCommentCount) / float64(metrics.SizeMetrics.FilesChanged)
		complexity += commentDensity * analyzer.config.CommentDensityWeight
	}
	
	return complexity
}

// getFileTypeWeight はファイルタイプの重みを取得
func (analyzer *PRComplexityAnalyzer) getFileTypeWeight(fileType string) float64 {
	// 拡張子を正規化
	if fileType != "" && !strings.HasPrefix(fileType, ".") {
		fileType = "." + fileType
	}
	
	weight, exists := analyzer.config.FileTypeWeights[fileType]
	if !exists {
		// 未知のファイルタイプは中程度の複雑度とする
		return 1.0
	}
	
	return weight
}

// GetComplexityLevel は複雑度スコアを人間が理解しやすいレベルに変換
func (analyzer *PRComplexityAnalyzer) GetComplexityLevel(score float64) ComplexityLevel {
	switch {
	case score <= 1.5:
		return ComplexityLevelVeryLow
	case score <= 2.5:
		return ComplexityLevelLow
	case score <= 4.0:
		return ComplexityLevelMedium
	case score <= 6.0:
		return ComplexityLevelHigh
	default:
		return ComplexityLevelVeryHigh
	}
}

// ComplexityLevel は複雑度レベル
type ComplexityLevel string

const (
	ComplexityLevelVeryLow  ComplexityLevel = "very_low"
	ComplexityLevelLow      ComplexityLevel = "low"
	ComplexityLevelMedium   ComplexityLevel = "medium"
	ComplexityLevelHigh     ComplexityLevel = "high"
	ComplexityLevelVeryHigh ComplexityLevel = "very_high"
)

// AnalyzeFileComplexity は個別ファイルの複雑度を分析
func (analyzer *PRComplexityAnalyzer) AnalyzeFileComplexity(fileChange FileChangeMetrics) float64 {
	baseComplexity := math.Log10(float64(fileChange.LinesAdded+fileChange.LinesDeleted) + 1)
	
	// ファイルタイプ重み
	typeWeight := analyzer.getFileTypeWeight(fileChange.FileType)
	
	// 操作タイプによる調整
	operationWeight := 1.0
	if fileChange.IsNewFile {
		operationWeight *= analyzer.config.NewFileWeight
	}
	if fileChange.IsDeleted {
		operationWeight *= analyzer.config.DeleteFileWeight
	}
	if fileChange.IsRenamed {
		operationWeight *= analyzer.config.RenameFileWeight
	}
	
	return baseComplexity * typeWeight * operationWeight
}

// SuggestOptimalSplit は大きなPRの分割提案を生成
func (analyzer *PRComplexityAnalyzer) SuggestOptimalSplit(metrics *PRMetrics) []SplitSuggestion {
	if !metrics.IsLargePR() {
		return nil
	}
	
	suggestions := []SplitSuggestion{}
	
	// ディレクトリ別分割の提案
	dirGroups := analyzer.groupFilesByDirectory(metrics.SizeMetrics.FileChanges)
	if len(dirGroups) > 1 {
		suggestions = append(suggestions, SplitSuggestion{
			Type:        SplitTypeByDirectory,
			Description: "ディレクトリ別に分割することで、影響範囲を限定できます",
			Groups:      dirGroups,
		})
	}
	
	// ファイルタイプ別分割の提案
	typeGroups := analyzer.groupFilesByType(metrics.SizeMetrics.FileChanges)
	if len(typeGroups) > 1 {
		suggestions = append(suggestions, SplitSuggestion{
			Type:        SplitTypeByFileType,
			Description: "ファイルタイプ別に分割することで、レビューの専門性を向上できます",
			Groups:      typeGroups,
		})
	}
	
	return suggestions
}

// groupFilesByDirectory はファイルをディレクトリ別にグループ化
func (analyzer *PRComplexityAnalyzer) groupFilesByDirectory(files []FileChangeMetrics) map[string][]string {
	groups := make(map[string][]string)
	
	for _, file := range files {
		dir := filepath.Dir(file.FileName)
		if dir == "." {
			dir = "root"
		}
		groups[dir] = append(groups[dir], file.FileName)
	}
	
	return groups
}

// groupFilesByType はファイルをタイプ別にグループ化
func (analyzer *PRComplexityAnalyzer) groupFilesByType(files []FileChangeMetrics) map[string][]string {
	groups := make(map[string][]string)
	
	for _, file := range files {
		fileType := file.FileType
		if fileType == "" {
			fileType = "other"
		}
		groups[fileType] = append(groups[fileType], file.FileName)
	}
	
	return groups
}

// SplitSuggestion はPR分割の提案
type SplitSuggestion struct {
	Type        SplitType             `json:"type"`
	Description string                `json:"description"`
	Groups      map[string][]string   `json:"groups"`
}

// SplitType は分割タイプ
type SplitType string

const (
	SplitTypeByDirectory SplitType = "by_directory"
	SplitTypeByFileType  SplitType = "by_file_type"
	SplitTypeByFeature   SplitType = "by_feature"
)