package pull_request

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// PRAnalysisService はPull Requestの分析を行うサービス
type PRAnalysisService struct {
	complexityAnalyzer *PRComplexityAnalyzer
	cycleTimeCalc      *CycleTimeCalculator
	reviewTimeAnalyzer *ReviewTimeAnalyzer
}

// NewPRAnalysisService は新しいPR分析サービスを作成
func NewPRAnalysisService() *PRAnalysisService {
	return &PRAnalysisService{
		complexityAnalyzer: NewPRComplexityAnalyzer(),
		cycleTimeCalc:      NewCycleTimeCalculator(),
		reviewTimeAnalyzer: NewReviewTimeAnalyzer(),
	}
}

// AnalyzePR はPull Requestの包括的な分析を実行
func (s *PRAnalysisService) AnalyzePR(ctx context.Context, pr PullRequest, reviewEvents []ReviewEvent, fileChanges []FileChangeMetrics) (*PRMetrics, error) {
	// 基本的なPRメトリクスの構築
	metrics := &PRMetrics{
		PRID:       pr.ID,
		PRNumber:   pr.Number,
		Title:      pr.Title,
		Author:     pr.Author.Login,
		Repository: pr.Repository.Name,
		CreatedAt:  pr.CreatedAt,
		MergedAt:   pr.MergedAt,
	}
	
	// サイズメトリクスの計算
	metrics.SizeMetrics = s.calculateSizeMetrics(pr, fileChanges)
	
	// 時間メトリクスの計算
	metrics.TimeMetrics = s.cycleTimeCalc.CalculateTimeMetrics(pr, reviewEvents)
	
	// 品質メトリクスの計算
	metrics.QualityMetrics = s.reviewTimeAnalyzer.CalculateQualityMetrics(pr, reviewEvents)
	
	// 複雑度スコアの計算
	metrics.ComplexityScore = s.complexityAnalyzer.AnalyzeComplexity(metrics)
	
	// サイズカテゴリの決定
	metrics.SizeCategory = metrics.CalculateSizeCategory()
	
	return metrics, nil
}

// AnalyzeBatch は複数のPRを一括で分析
func (s *PRAnalysisService) AnalyzeBatch(ctx context.Context, prs []PullRequest) ([]*PRMetrics, error) {
	results := make([]*PRMetrics, 0, len(prs))
	
	for _, pr := range prs {
		// 簡易分析（詳細データなし）
		metrics := s.analyzeBasicPR(pr)
		results = append(results, metrics)
	}
	
	return results, nil
}

// calculateSizeMetrics はサイズメトリクスを計算
func (s *PRAnalysisService) calculateSizeMetrics(pr PullRequest, fileChanges []FileChangeMetrics) PRSizeMetrics {
	sizeMetrics := PRSizeMetrics{
		LinesAdded:   pr.Additions,
		LinesDeleted: pr.Deletions,
		LinesChanged: pr.Additions + pr.Deletions,
		FilesChanged: len(fileChanges),
	}
	
	// ファイル詳細がある場合
	if len(fileChanges) > 0 {
		sizeMetrics.FileChanges = fileChanges
		
		// ファイルタイプ別集計
		fileTypeBreakdown := make(map[string]int)
		directories := make(map[string]bool)
		
		for _, file := range fileChanges {
			fileTypeBreakdown[file.FileType]++
			// ディレクトリ情報を抽出（簡易版）
			if len(file.FileName) > 0 {
				dirs := extractDirectories(file.FileName)
				for _, dir := range dirs {
					directories[dir] = true
				}
			}
		}
		
		sizeMetrics.FileTypeBreakdown = fileTypeBreakdown
		sizeMetrics.DirectoryCount = len(directories)
	} else {
		// ファイル詳細がない場合の推定
		sizeMetrics.FileChanges = []FileChangeMetrics{}
		sizeMetrics.FileTypeBreakdown = map[string]int{"unknown": 1}
		sizeMetrics.DirectoryCount = 1
	}
	
	return sizeMetrics
}

// analyzeBasicPR は基本的なPR分析を実行（軽量版）
func (s *PRAnalysisService) analyzeBasicPR(pr PullRequest) *PRMetrics {
	metrics := &PRMetrics{
		PRID:       pr.ID,
		PRNumber:   pr.Number,
		Title:      pr.Title,
		Author:     pr.Author.Login,
		Repository: pr.Repository.Name,
		CreatedAt:  pr.CreatedAt,
		MergedAt:   pr.MergedAt,
	}
	
	// 基本的なサイズメトリクス
	metrics.SizeMetrics = PRSizeMetrics{
		LinesAdded:        pr.Additions,
		LinesDeleted:      pr.Deletions,
		LinesChanged:      pr.Additions + pr.Deletions,
		FilesChanged:      1, // 推定値
		FileChanges:       []FileChangeMetrics{},
		FileTypeBreakdown: map[string]int{"unknown": 1},
		DirectoryCount:    1,
	}
	
	// 基本的な時間メトリクス
	metrics.TimeMetrics = PRTimeMetrics{
		CreatedHour: pr.CreatedAt.Hour(),
	}
	
	if pr.MergedAt != nil {
		totalCycle := pr.MergedAt.Sub(pr.CreatedAt)
		metrics.TimeMetrics.TotalCycleTime = &totalCycle
		mergedHour := pr.MergedAt.Hour()
		metrics.TimeMetrics.MergedHour = &mergedHour
	}
	
	if pr.FirstReviewed != nil {
		reviewTime := pr.FirstReviewed.Sub(pr.CreatedAt)
		metrics.TimeMetrics.TimeToFirstReview = &reviewTime
	}
	
	if pr.LastApproved != nil {
		approvalTime := pr.LastApproved.Sub(pr.CreatedAt)
		metrics.TimeMetrics.TimeToApproval = &approvalTime
		
		if pr.MergedAt != nil {
			mergeTime := pr.MergedAt.Sub(*pr.LastApproved)
			metrics.TimeMetrics.TimeToMerge = &mergeTime
		}
	}
	
	// 基本的な品質メトリクス（推定）
	metrics.QualityMetrics = PRQualityMetrics{
		ReviewCommentCount:    0,
		ReviewRoundCount:      1,
		ReviewerCount:         1,
		ReviewersInvolved:     []string{},
		CommitCount:           1,
		FixupCommitCount:      0,
		ForceUpdateCount:      0,
		FirstReviewPassRate:   1.0,
		AverageCommentPerFile: 0.0,
		ApprovalsReceived:     1,
		ApproversInvolved:     []string{},
	}
	
	// 複雑度スコア（基本値）
	metrics.ComplexityScore = s.complexityAnalyzer.AnalyzeComplexity(metrics)
	
	// サイズカテゴリ
	metrics.SizeCategory = metrics.CalculateSizeCategory()
	
	return metrics
}

// GenerateInsights はPRの分析結果から洞察を生成
func (s *PRAnalysisService) GenerateInsights(metrics *PRMetrics) *PRInsights {
	insights := &PRInsights{
		PRID: metrics.PRID,
	}
	
	// サイズ関連の洞察
	if metrics.IsLargePR() {
		insights.SizeIssues = append(insights.SizeIssues, SizeIssue{
			Type:        SizeIssueTypeTooLarge,
			Severity:    getSizeSeverity(metrics.SizeCategory),
			Description: fmt.Sprintf("PRが大きすぎます（%d行の変更）。レビューが困難になる可能性があります。", metrics.SizeMetrics.LinesChanged),
			Suggestion:  "PRを機能単位で分割することを検討してください。",
		})
		
		// 分割提案
		if suggestions := s.complexityAnalyzer.SuggestOptimalSplit(metrics); len(suggestions) > 0 {
			insights.SplitSuggestions = suggestions
		}
	}
	
	// 時間関連の洞察
	if metrics.TimeMetrics.TimeToFirstReview != nil {
		reviewTime := *metrics.TimeMetrics.TimeToFirstReview
		if reviewTime > 24*time.Hour {
			insights.TimeIssues = append(insights.TimeIssues, TimeIssue{
				Type:        TimeIssueTypeSlowReview,
				Severity:    IssueseverityMedium,
				Description: fmt.Sprintf("初回レビューまでに%s かかりました。", formatDuration(reviewTime)),
				Suggestion:  "レビュアーの割り当てやレビュー依頼のタイミングを見直してください。",
			})
		}
	}
	
	// 品質関連の洞察
	if metrics.QualityMetrics.ReviewRoundCount > 3 {
		insights.QualityIssues = append(insights.QualityIssues, QualityIssue{
			Type:        QualityIssueTypeMultipleRounds,
			Severity:    IssueseverityMedium,
			Description: fmt.Sprintf("レビューが%d回行われました。", metrics.QualityMetrics.ReviewRoundCount),
			Suggestion:  "事前のセルフレビューやコード品質の向上を検討してください。",
		})
	}
	
	// 複雑度関連の洞察
	complexityLevel := s.complexityAnalyzer.GetComplexityLevel(metrics.ComplexityScore)
	if complexityLevel == ComplexityLevelHigh || complexityLevel == ComplexityLevelVeryHigh {
		insights.ComplexityIssues = append(insights.ComplexityIssues, ComplexityIssue{
			Type:        ComplexityIssueTypeHighComplexity,
			Severity:    getComplexitySeverity(complexityLevel),
			Description: fmt.Sprintf("PRの複雑度が高い（スコア: %.2f）です。", metrics.ComplexityScore),
			Suggestion:  "機能を小さく分割し、シンプルな変更にすることを検討してください。",
		})
	}
	
	// 推奨アクション
	insights.RecommendedActions = s.generateRecommendedActions(metrics)
	
	return insights
}

// generateRecommendedActions は推奨アクションを生成
func (s *PRAnalysisService) generateRecommendedActions(metrics *PRMetrics) []RecommendedAction {
	var actions []RecommendedAction
	
	// サイズベースの推奨
	if metrics.IsLargePR() {
		actions = append(actions, RecommendedAction{
			Type:        ActionTypeSplitPR,
			Priority:    ActionPriorityHigh,
			Description: "PRを小さく分割してレビューしやすくする",
			Reason:      "大きなPRはレビューが困難でバグの見逃しリスクが高まります",
		})
	}
	
	// 時間ベースの推奨
	if metrics.HasLongReviewTime(4 * time.Hour) {
		actions = append(actions, RecommendedAction{
			Type:        ActionTypeImproveReviewProcess,
			Priority:    ActionPriorityMedium,
			Description: "レビュープロセスの改善（レビュアー追加、優先度設定）",
			Reason:      "レビュー時間が長く、開発スピードに影響しています",
		})
	}
	
	// 品質ベースの推奨
	if !metrics.IsHighQuality() {
		actions = append(actions, RecommendedAction{
			Type:        ActionTypeImproveCodeQuality,
			Priority:    ActionPriorityMedium,
			Description: "コード品質の向上（事前チェック、テスト追加）",
			Reason:      "レビューコメントが多く、修正回数が多い傾向があります",
		})
	}
	
	// 複雑度ベースの推奨
	complexityLevel := s.complexityAnalyzer.GetComplexityLevel(metrics.ComplexityScore)
	if complexityLevel >= ComplexityLevelHigh {
		actions = append(actions, RecommendedAction{
			Type:        ActionTypeSimplifyChanges,
			Priority:    ActionPriorityHigh,
			Description: "変更内容の簡素化と段階的実装",
			Reason:      "複雑な変更はレビューが困難で、バグ混入リスクが高いです",
		})
	}
	
	return actions
}

// extractDirectories はファイルパスからディレクトリを抽出
func extractDirectories(filePath string) []string {
	// 簡易実装：ファイルパスを分解してディレクトリを抽出
	if filePath == "" {
		return []string{}
	}
	
	// "src/components/Button.tsx" -> ["src", "src/components"]
	dirs := []string{}
	parts := strings.Split(filePath, "/")
	
	for i := 0; i < len(parts)-1; i++ { // 最後はファイル名なので除外
		if i == 0 {
			dirs = append(dirs, parts[i])
		} else {
			dirs = append(dirs, strings.Join(parts[0:i+1], "/"))
		}
	}
	
	return dirs
}

// formatDuration は時間を人間が読みやすい形式でフォーマット
func formatDuration(d time.Duration) string {
	if d < time.Hour {
		return fmt.Sprintf("%.0f分", d.Minutes())
	} else if d < 24*time.Hour {
		return fmt.Sprintf("%.1f時間", d.Hours())
	} else {
		return fmt.Sprintf("%.1f日", d.Hours()/24)
	}
}

// ヘルパー関数群

func getSizeSeverity(category PRSizeCategory) IssueSeverity {
	switch category {
	case PRSizeLarge:
		return IssueseverityMedium
	case PRSizeXLarge:
		return IssueseverityHigh
	default:
		return IssueseverityLow
	}
}

func getComplexitySeverity(level ComplexityLevel) IssueSeverity {
	switch level {
	case ComplexityLevelHigh:
		return IssueseverityMedium
	case ComplexityLevelVeryHigh:
		return IssueseverityHigh
	default:
		return IssueseverityLow
	}
}

// PRInsights はPR分析の洞察結果
type PRInsights struct {
	PRID string `json:"prId"`
	
	// 問題の分類
	SizeIssues       []SizeIssue       `json:"sizeIssues"`
	TimeIssues       []TimeIssue       `json:"timeIssues"`
	QualityIssues    []QualityIssue    `json:"qualityIssues"`
	ComplexityIssues []ComplexityIssue `json:"complexityIssues"`
	
	// 提案
	SplitSuggestions    []SplitSuggestion    `json:"splitSuggestions"`
	RecommendedActions  []RecommendedAction  `json:"recommendedActions"`
}

// 各種Issue構造体
type SizeIssue struct {
	Type        SizeIssueType `json:"type"`
	Severity    IssueSeverity `json:"severity"`
	Description string        `json:"description"`
	Suggestion  string        `json:"suggestion"`
}

type TimeIssue struct {
	Type        TimeIssueType `json:"type"`
	Severity    IssueSeverity `json:"severity"`
	Description string        `json:"description"`
	Suggestion  string        `json:"suggestion"`
}

type QualityIssue struct {
	Type        QualityIssueType `json:"type"`
	Severity    IssueSeverity    `json:"severity"`
	Description string           `json:"description"`
	Suggestion  string           `json:"suggestion"`
}

type ComplexityIssue struct {
	Type        ComplexityIssueType `json:"type"`
	Severity    IssueSeverity       `json:"severity"`
	Description string              `json:"description"`
	Suggestion  string              `json:"suggestion"`
}

type RecommendedAction struct {
	Type        ActionType     `json:"type"`
	Priority    ActionPriority `json:"priority"`
	Description string         `json:"description"`
	Reason      string         `json:"reason"`
}

// 各種Enum定義
type SizeIssueType string
const (
	SizeIssueTypeTooLarge SizeIssueType = "too_large"
	SizeIssueTypeTooManyFiles SizeIssueType = "too_many_files"
)

type TimeIssueType string
const (
	TimeIssueTypeSlowReview TimeIssueType = "slow_review"
	TimeIssueTypeLongCycle TimeIssueType = "long_cycle"
)

type QualityIssueType string
const (
	QualityIssueTypeMultipleRounds QualityIssueType = "multiple_rounds"
	QualityIssueTypeManyComments QualityIssueType = "many_comments"
)

type ComplexityIssueType string
const (
	ComplexityIssueTypeHighComplexity ComplexityIssueType = "high_complexity"
	ComplexityIssueTypeMixedFileTypes ComplexityIssueType = "mixed_file_types"
)

type ActionType string
const (
	ActionTypeSplitPR ActionType = "split_pr"
	ActionTypeImproveReviewProcess ActionType = "improve_review_process"
	ActionTypeImproveCodeQuality ActionType = "improve_code_quality"
	ActionTypeSimplifyChanges ActionType = "simplify_changes"
)

type ActionPriority string
const (
	ActionPriorityHigh ActionPriority = "high"
	ActionPriorityMedium ActionPriority = "medium"
	ActionPriorityLow ActionPriority = "low"
)

type IssueSeverity string
const (
	IssueseverityHigh IssueSeverity = "high"
	IssueseverityMedium IssueSeverity = "medium"
	IssueseverityLow IssueSeverity = "low"
)