package pull_request

import (
	"testing"
	"time"
)

func TestReviewTimeAnalyzer_AnalyzeReviewEfficiency(t *testing.T) {
	analyzer := NewReviewTimeAnalyzer()
	
	tests := []struct {
		name         string
		metrics      []*PRMetrics
		expectedMin  int
		expectedMax  int
	}{
		{
			"効率的なレビュー",
			[]*PRMetrics{
				{
					QualityMetrics: PRQualityMetrics{
						ReviewRoundCount:   1,
						ReviewCommentCount: 2,
						ReviewerCount:      2,
						ReviewersInvolved:  []string{"reviewer1", "reviewer2"},
						ApproversInvolved:  []string{"reviewer1"},
					},
				},
			},
			1,
			1,
		},
		{
			"複数のPRメトリクス",
			[]*PRMetrics{
				{
					QualityMetrics: PRQualityMetrics{
						ReviewRoundCount:   1,
						ReviewCommentCount: 2,
						ReviewerCount:      2,
						ReviewersInvolved:  []string{"reviewer1", "reviewer2"},
						ApproversInvolved:  []string{"reviewer1"},
					},
				},
				{
					QualityMetrics: PRQualityMetrics{
						ReviewRoundCount:   3,
						ReviewCommentCount: 15,
						ReviewerCount:      1,
						ReviewersInvolved:  []string{"reviewer3"},
						ApproversInvolved:  []string{"reviewer3"},
					},
				},
			},
			2,
			2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := analyzer.AnalyzeReviewEfficiency(tt.metrics)
			
			if result.TotalPRs < tt.expectedMin || result.TotalPRs > tt.expectedMax {
				t.Errorf("TotalPRs = %v, want between %v and %v", 
					result.TotalPRs, tt.expectedMin, tt.expectedMax)
			}
			
			// 基本的なフィールドの存在確認
			if result.OverallStats.CommentStats.Count == 0 && len(tt.metrics) > 0 {
				t.Error("OverallStats should have comment statistics")
			}
		})
	}
}

func TestReviewTimeAnalyzer_AnalyzeReviewBottlenecks(t *testing.T) {
	analyzer := NewReviewTimeAnalyzer()
	
	baseTime := time.Date(2024, 1, 15, 9, 0, 0, 0, time.UTC)
	
	tests := []struct {
		name             string
		metrics          []*PRMetrics
		expectedMinCount int
		expectedMaxCount int
	}{
		{
			"ボトルネックなし",
			[]*PRMetrics{
				createEfficientPRMetrics(baseTime),
			},
			0,
			2,
		},
		{
			"複数のボトルネック",
			[]*PRMetrics{
				createBottleneckPRMetrics(baseTime),
				createLongReviewTimePRMetrics(baseTime),
			},
			2,
			6,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := analyzer.AnalyzeReviewBottlenecks(tt.metrics)
			
			count := len(result)
			if count < tt.expectedMinCount || count > tt.expectedMaxCount {
				t.Errorf("Bottlenecks count = %v, want between %v and %v", 
					count, tt.expectedMinCount, tt.expectedMaxCount)
			}
			
			// ボトルネックの内容確認
			for _, bottleneck := range result {
				if bottleneck.Type == "" {
					t.Error("Bottleneck.Type should not be empty")
				}
				if bottleneck.Description == "" {
					t.Error("Bottleneck.Description should not be empty")
				}
			}
		})
	}
}

func TestReviewTimeAnalyzer_CalculateQualityMetrics(t *testing.T) {
	analyzer := NewReviewTimeAnalyzer()
	
	baseTime := time.Date(2024, 1, 15, 9, 0, 0, 0, time.UTC)
	
	pr := PullRequest{
		Additions: 100,
		Deletions: 50,
	}
	
	events := []ReviewEvent{
		{Type: ReviewEventTypeRequested, CreatedAt: baseTime, Actor: "author"},
		{Type: ReviewEventTypeCommented, CreatedAt: baseTime.Add(2 * time.Hour), Actor: "reviewer1"},
		{Type: ReviewEventTypeChangesRequested, CreatedAt: baseTime.Add(4 * time.Hour), Actor: "reviewer2"},
		{Type: ReviewEventTypeApproved, CreatedAt: baseTime.Add(6 * time.Hour), Actor: "reviewer1"},
	}
	
	result := analyzer.CalculateQualityMetrics(pr, events)
	
	// 基本的なカウントの確認
	if result.ReviewCommentCount == 0 {
		t.Error("Expected non-zero review comment count")
	}
	
	if result.ReviewRoundCount == 0 {
		t.Error("Expected non-zero review round count")
	}
	
	if result.ReviewerCount == 0 {
		t.Error("Expected non-zero reviewer count")
	}
	
	if len(result.ReviewersInvolved) == 0 {
		t.Error("Expected non-empty reviewers list")
	}
}

func TestReviewTimeAnalyzer_countReviewComments(t *testing.T) {
	analyzer := NewReviewTimeAnalyzer()
	
	events := []ReviewEvent{
		{Type: ReviewEventTypeRequested},
		{Type: ReviewEventTypeCommented},
		{Type: ReviewEventTypeChangesRequested},
		{Type: ReviewEventTypeApproved},
		{Type: ReviewEventTypeCommented},
	}
	
	result := analyzer.countReviewComments(events)
	
	// コメントと変更要求の合計を期待
	expected := 3 // コメント2つ + 変更要求1つ
	if result != expected {
		t.Errorf("countReviewComments() = %v, want %v", result, expected)
	}
}

func TestReviewTimeAnalyzer_calculateReviewRounds(t *testing.T) {
	analyzer := NewReviewTimeAnalyzer()
	baseTime := time.Now()
	
	tests := []struct {
		name     string
		events   []ReviewEvent
		expected int
	}{
		{
			"変更要求なし",
			[]ReviewEvent{
				{Type: ReviewEventTypeCommented, CreatedAt: baseTime},
				{Type: ReviewEventTypeApproved, CreatedAt: baseTime.Add(time.Hour)},
			},
			1,
		},
		{
			"1回の変更要求",
			[]ReviewEvent{
				{Type: ReviewEventTypeChangesRequested, CreatedAt: baseTime},
			},
			1,
		},
		{
			"複数の変更要求（時間間隔あり）",
			[]ReviewEvent{
				{Type: ReviewEventTypeChangesRequested, CreatedAt: baseTime},
				{Type: ReviewEventTypeChangesRequested, CreatedAt: baseTime.Add(3 * time.Hour)},
			},
			2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := analyzer.calculateReviewRounds(tt.events)
			
			if result != tt.expected {
				t.Errorf("calculateReviewRounds() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestReviewTimeAnalyzer_EdgeCases(t *testing.T) {
	analyzer := NewReviewTimeAnalyzer()
	
	t.Run("空のメトリクス", func(t *testing.T) {
		result := analyzer.AnalyzeReviewBottlenecks([]*PRMetrics{})
		
		if len(result) != 0 {
			t.Errorf("Expected 0 bottlenecks for empty metrics, got %d", len(result))
		}
	})
	
	t.Run("空のイベント", func(t *testing.T) {
		metrics := []*PRMetrics{
			{
				QualityMetrics: PRQualityMetrics{
					ReviewRoundCount:   1,
					ReviewCommentCount: 1,
					ReviewerCount:      1,
					ReviewersInvolved:  []string{"reviewer1"},
					ApproversInvolved:  []string{"reviewer1"},
				},
			},
		}
		
		result := analyzer.AnalyzeReviewEfficiency(metrics)
		
		// 基本的な効率性分析は実行される
		if result.TotalPRs != 1 {
			t.Errorf("Expected TotalPRs to be 1, got %v", result.TotalPRs)
		}
	})
	
	t.Run("極端な値", func(t *testing.T) {
		metrics := []*PRMetrics{
			{
				QualityMetrics: PRQualityMetrics{
					ReviewRoundCount:   100, // 極端に多い
					ReviewCommentCount: 1000, // 極端に多い
					ReviewerCount:      0,    // 0人
					ReviewersInvolved:  []string{},
					ApproversInvolved:  []string{},
				},
			},
		}
		
		result := analyzer.AnalyzeReviewEfficiency(metrics)
		
		// 基本的な分析は実行される
		if result.TotalPRs != 1 {
			t.Errorf("Expected TotalPRs to be 1, got %v", result.TotalPRs)
		}
	})
}

// ヘルパー関数

func createEfficientPRMetrics(baseTime time.Time) *PRMetrics {
	return &PRMetrics{
		PRID:      "efficient-pr",
		CreatedAt: baseTime,
		TimeMetrics: PRTimeMetrics{
			TimeToFirstReview: durationPtr(1 * time.Hour),
			TimeToApproval:    durationPtr(2 * time.Hour),
			TotalCycleTime:    durationPtr(4 * time.Hour),
		},
		QualityMetrics: PRQualityMetrics{
			ReviewRoundCount:   1,
			ReviewCommentCount: 2,
			ReviewerCount:      2,
		},
		SizeMetrics: PRSizeMetrics{
			LinesChanged: 50,
		},
	}
}

func createBottleneckPRMetrics(baseTime time.Time) *PRMetrics {
	return &PRMetrics{
		PRID:      "bottleneck-pr",
		CreatedAt: baseTime,
		TimeMetrics: PRTimeMetrics{
			TimeToFirstReview: durationPtr(48 * time.Hour), // 遅い
			TimeToApproval:    durationPtr(120 * time.Hour), // 非常に遅い
			TotalCycleTime:    durationPtr(200 * time.Hour), // 8日以上
		},
		QualityMetrics: PRQualityMetrics{
			ReviewRoundCount:   8, // 多い
			ReviewCommentCount: 50, // 多い
			ReviewerCount:      1, // 少ない
		},
		SizeMetrics: PRSizeMetrics{
			LinesChanged: 800, // 大きい
		},
	}
}

func createLongReviewTimePRMetrics(baseTime time.Time) *PRMetrics {
	return &PRMetrics{
		PRID:      "long-review-pr",
		CreatedAt: baseTime,
		TimeMetrics: PRTimeMetrics{
			TimeToFirstReview: durationPtr(72 * time.Hour), // 3日
			TimeToApproval:    durationPtr(168 * time.Hour), // 1週間
			TotalCycleTime:    durationPtr(240 * time.Hour), // 10日
		},
		QualityMetrics: PRQualityMetrics{
			ReviewRoundCount:   3,
			ReviewCommentCount: 15,
			ReviewerCount:      2,
		},
		SizeMetrics: PRSizeMetrics{
			LinesChanged: 300,
		},
	}
}

func createPatternPRMetrics(baseTime time.Time, hour int) *PRMetrics {
	createdTime := time.Date(baseTime.Year(), baseTime.Month(), baseTime.Day(), 
		hour, 0, 0, 0, baseTime.Location())
	
	return &PRMetrics{
		PRID:      "pattern-pr",
		CreatedAt: createdTime,
		TimeMetrics: PRTimeMetrics{
			TimeToFirstReview: durationPtr(2 * time.Hour),
			TimeToApproval:    durationPtr(4 * time.Hour),
			CreatedHour:       hour,
		},
		QualityMetrics: PRQualityMetrics{
			ReviewRoundCount:   2,
			ReviewCommentCount: 5,
			ReviewerCount:      2,
		},
	}
}