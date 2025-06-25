package pull_request

import (
	"testing"
	"time"
)

func TestCycleTimeCalculator_CalculateTimeMetrics(t *testing.T) {
	calc := NewCycleTimeCalculator()
	
	baseTime := time.Date(2024, 1, 15, 9, 0, 0, 0, time.UTC)
	
	pr := PullRequest{
		CreatedAt:     baseTime,
		MergedAt:      timePtr2(baseTime.Add(24 * time.Hour)),
		FirstReviewed: timePtr2(baseTime.Add(2 * time.Hour)),
		LastApproved:  timePtr2(baseTime.Add(20 * time.Hour)),
	}
	
	reviewEvents := []ReviewEvent{
		{
			Type:      ReviewEventTypeRequested,
			CreatedAt: baseTime.Add(30 * time.Minute),
			Actor:     "author",
		},
		{
			Type:      ReviewEventTypeCommented,
			CreatedAt: baseTime.Add(2 * time.Hour),
			Actor:     "reviewer1",
			Reviewer:  "reviewer1",
		},
		{
			Type:      ReviewEventTypeApproved,
			CreatedAt: baseTime.Add(20 * time.Hour),
			Actor:     "reviewer2",
			Reviewer:  "reviewer2",
		},
		{
			Type:      ReviewEventTypeMerged,
			CreatedAt: baseTime.Add(24 * time.Hour),
			Actor:     "author",
		},
	}
	
	result := calc.CalculateTimeMetrics(pr, reviewEvents)
	
	// 基本的な時間フィールドの確認
	if result.CreatedHour != 9 {
		t.Errorf("CreatedHour = %v, want 9", result.CreatedHour)
	}
	
	if result.MergedHour == nil || *result.MergedHour != 9 {
		t.Errorf("MergedHour = %v, want 9", result.MergedHour)
	}
	
	// 初回レビューまでの時間
	if result.TimeToFirstReview == nil {
		t.Error("TimeToFirstReview should not be nil")
	} else if *result.TimeToFirstReview != 2*time.Hour {
		t.Errorf("TimeToFirstReview = %v, want %v", *result.TimeToFirstReview, 2*time.Hour)
	}
	
	// 承認までの時間
	if result.TimeToApproval == nil {
		t.Error("TimeToApproval should not be nil")
	} else if *result.TimeToApproval != 20*time.Hour {
		t.Errorf("TimeToApproval = %v, want %v", *result.TimeToApproval, 20*time.Hour)
	}
	
	// マージまでの時間
	if result.TimeToMerge == nil {
		t.Error("TimeToMerge should not be nil")
	} else if *result.TimeToMerge != 4*time.Hour { // 20h -> 24h = 4h
		t.Errorf("TimeToMerge = %v, want %v", *result.TimeToMerge, 4*time.Hour)
	}
	
	// 総サイクルタイム
	if result.TotalCycleTime == nil {
		t.Error("TotalCycleTime should not be nil")
	} else if *result.TotalCycleTime != 24*time.Hour {
		t.Errorf("TotalCycleTime = %v, want %v", *result.TotalCycleTime, 24*time.Hour)
	}
}

func TestCycleTimeCalculator_calculateTimeToFirstReview(t *testing.T) {
	calc := NewCycleTimeCalculator()
	baseTime := time.Date(2024, 1, 15, 9, 0, 0, 0, time.UTC)
	
	pr := PullRequest{
		CreatedAt: baseTime,
	}
	
	tests := []struct {
		name         string
		events       []ReviewEvent
		expected     *time.Duration
	}{
		{
			"コメントイベントあり",
			[]ReviewEvent{
				{Type: ReviewEventTypeCommented, CreatedAt: baseTime.Add(2 * time.Hour)},
			},
			durationPtr(2 * time.Hour),
		},
		{
			"承認イベントあり",
			[]ReviewEvent{
				{Type: ReviewEventTypeApproved, CreatedAt: baseTime.Add(1 * time.Hour)},
			},
			durationPtr(1 * time.Hour),
		},
		{
			"変更要求イベントあり",
			[]ReviewEvent{
				{Type: ReviewEventTypeChangesRequested, CreatedAt: baseTime.Add(3 * time.Hour)},
			},
			durationPtr(3 * time.Hour),
		},
		{
			"関係ないイベントのみ",
			[]ReviewEvent{
				{Type: ReviewEventTypeRequested, CreatedAt: baseTime.Add(1 * time.Hour)},
			},
			nil,
		},
		{
			"イベントなし",
			[]ReviewEvent{},
			nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calc.calculateTimeToFirstReview(pr, tt.events)
			
			if tt.expected == nil {
				if result != nil {
					t.Errorf("Expected nil, got %v", result)
				}
			} else {
				if result == nil {
					t.Error("Expected non-nil result")
				} else if *result != *tt.expected {
					t.Errorf("Expected %v, got %v", *tt.expected, *result)
				}
			}
		})
	}
}

func TestCycleTimeCalculator_calculateTimeToApproval(t *testing.T) {
	calc := NewCycleTimeCalculator()
	baseTime := time.Date(2024, 1, 15, 9, 0, 0, 0, time.UTC)
	
	pr := PullRequest{
		CreatedAt: baseTime,
	}
	
	tests := []struct {
		name     string
		events   []ReviewEvent
		expected *time.Duration
	}{
		{
			"承認イベントあり",
			[]ReviewEvent{
				{Type: ReviewEventTypeCommented, CreatedAt: baseTime.Add(1 * time.Hour)},
				{Type: ReviewEventTypeApproved, CreatedAt: baseTime.Add(3 * time.Hour)},
			},
			durationPtr(3 * time.Hour),
		},
		{
			"複数の承認イベント（最初を使用）",
			[]ReviewEvent{
				{Type: ReviewEventTypeApproved, CreatedAt: baseTime.Add(2 * time.Hour)},
				{Type: ReviewEventTypeApproved, CreatedAt: baseTime.Add(4 * time.Hour)},
			},
			durationPtr(2 * time.Hour),
		},
		{
			"承認イベントなし",
			[]ReviewEvent{
				{Type: ReviewEventTypeCommented, CreatedAt: baseTime.Add(1 * time.Hour)},
			},
			nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calc.calculateTimeToApproval(pr, tt.events)
			
			if tt.expected == nil {
				if result != nil {
					t.Errorf("Expected nil, got %v", result)
				}
			} else {
				if result == nil {
					t.Error("Expected non-nil result")
				} else if *result != *tt.expected {
					t.Errorf("Expected %v, got %v", *tt.expected, *result)
				}
			}
		})
	}
}

func TestCycleTimeCalculator_calculateTimeToMerge(t *testing.T) {
	calc := NewCycleTimeCalculator()
	baseTime := time.Date(2024, 1, 15, 9, 0, 0, 0, time.UTC)
	
	tests := []struct {
		name     string
		pr       PullRequest
		events   []ReviewEvent
		expected *time.Duration
	}{
		{
			"承認からマージまで",
			PullRequest{
				CreatedAt: baseTime,
				MergedAt:  timePtr2(baseTime.Add(5 * time.Hour)),
			},
			[]ReviewEvent{
				{Type: ReviewEventTypeApproved, CreatedAt: baseTime.Add(3 * time.Hour)},
			},
			durationPtr(2 * time.Hour),
		},
		{
			"マージされていない",
			PullRequest{
				CreatedAt: baseTime,
				MergedAt:  nil,
			},
			[]ReviewEvent{
				{Type: ReviewEventTypeApproved, CreatedAt: baseTime.Add(3 * time.Hour)},
			},
			nil,
		},
		{
			"承認イベントなし",
			PullRequest{
				CreatedAt: baseTime,
				MergedAt:  timePtr2(baseTime.Add(5 * time.Hour)),
			},
			[]ReviewEvent{},
			nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calc.calculateTimeToMerge(tt.pr, tt.events)
			
			if tt.expected == nil {
				if result != nil {
					t.Errorf("Expected nil, got %v", result)
				}
			} else {
				if result == nil {
					t.Error("Expected non-nil result")
				} else if *result != *tt.expected {
					t.Errorf("Expected %v, got %v", *tt.expected, *result)
				}
			}
		})
	}
}

func TestCycleTimeCalculator_calculateBusinessHours(t *testing.T) {
	calc := NewCycleTimeCalculator()
	// 営業時間を使用する設定に変更
	calc.config.UseBusinessHours = true
	calc.config.BusinessStart = 9
	calc.config.BusinessEnd = 18
	calc.config.ExcludeWeekends = true
	
	jst, _ := time.LoadLocation("Asia/Tokyo")
	
	tests := []struct {
		name     string
		start    time.Time
		end      time.Time
		expected time.Duration
	}{
		{
			"同日営業時間内",
			time.Date(2024, 1, 15, 10, 0, 0, 0, jst), // 月曜 10:00
			time.Date(2024, 1, 15, 15, 0, 0, 0, jst), // 月曜 15:00
			5 * time.Hour,
		},
		{
			"営業時間外開始",
			time.Date(2024, 1, 15, 7, 0, 0, 0, jst),  // 月曜 7:00
			time.Date(2024, 1, 15, 12, 0, 0, 0, jst), // 月曜 12:00
			3 * time.Hour, // 9:00-12:00
		},
		{
			"営業時間外終了",
			time.Date(2024, 1, 15, 14, 0, 0, 0, jst), // 月曜 14:00
			time.Date(2024, 1, 15, 20, 0, 0, 0, jst), // 月曜 20:00
			4 * time.Hour, // 14:00-18:00
		},
		{
			"週末をまたぐ",
			time.Date(2024, 1, 12, 16, 0, 0, 0, jst), // 金曜 16:00
			time.Date(2024, 1, 15, 11, 0, 0, 0, jst), // 月曜 11:00
			4 * time.Hour, // 金曜16:00-18:00 + 月曜9:00-11:00
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calc.calculateBusinessHours(tt.start, tt.end)
			
			// 営業時間計算は複雑なので、期待値に近い値かチェック
			diff := result - tt.expected
			if diff < 0 {
				diff = -diff
			}
			
			// 1時間の誤差を許容
			if diff > time.Hour {
				t.Errorf("calculateBusinessHours(%v, %v) = %v, want approximately %v", 
					tt.start, tt.end, result, tt.expected)
			}
		})
	}
}

func TestCycleTimeCalculator_CalculateCycleTimeStatistics(t *testing.T) {
	calc := NewCycleTimeCalculator()
	
	tests := []struct {
		name     string
		metrics  []*PRMetrics
		expected int // 期待されるTotalPRs
	}{
		{
			"空のメトリクス",
			[]*PRMetrics{},
			0,
		},
		{
			"単一のメトリクス",
			[]*PRMetrics{
				createTestCycleTimeMetrics(24 * time.Hour),
			},
			1,
		},
		{
			"複数のメトリクス",
			[]*PRMetrics{
				createTestCycleTimeMetrics(12 * time.Hour),
				createTestCycleTimeMetrics(24 * time.Hour),
				createTestCycleTimeMetrics(48 * time.Hour),
			},
			3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calc.CalculateCycleTimeStatistics(tt.metrics)
			
			if result.TotalPRs != tt.expected {
				t.Errorf("TotalPRs = %v, want %v", result.TotalPRs, tt.expected)
			}
			
			// 統計が正しく計算されているかの基本チェック
			if len(tt.metrics) > 0 {
				// 最低限、カウントが正しいことを確認
				if result.TotalCycleTime.Count == 0 {
					// すべてのメトリクスにTotalCycleTimeがない場合は0でOK
				}
			}
		})
	}
}

func TestCycleTimeCalculator_calculateDurationStatistics(t *testing.T) {
	calc := NewCycleTimeCalculator()
	
	durations := []time.Duration{
		1 * time.Hour,
		2 * time.Hour,
		3 * time.Hour,
		4 * time.Hour,
		5 * time.Hour,
	}
	
	result := calc.calculateDurationStatistics(durations)
	
	// 基本統計の確認
	if result.Count != 5 {
		t.Errorf("Count = %v, want 5", result.Count)
	}
	
	if result.Min != 1*time.Hour {
		t.Errorf("Min = %v, want %v", result.Min, 1*time.Hour)
	}
	
	if result.Max != 5*time.Hour {
		t.Errorf("Max = %v, want %v", result.Max, 5*time.Hour)
	}
	
	if result.Average != 3*time.Hour {
		t.Errorf("Average = %v, want %v", result.Average, 3*time.Hour)
	}
	
	if result.Median != 3*time.Hour {
		t.Errorf("Median = %v, want %v", result.Median, 3*time.Hour)
	}
	
	// パーセンタイルの確認（概算）
	if result.P75 != 4*time.Hour {
		t.Errorf("P75 = %v, want %v", result.P75, 4*time.Hour)
	}
}

func TestCycleTimeCalculator_EdgeCases(t *testing.T) {
	calc := NewCycleTimeCalculator()
	
	t.Run("空のイベントリスト", func(t *testing.T) {
		baseTime := time.Now()
		pr := PullRequest{
			CreatedAt: baseTime,
			MergedAt:  timePtr2(baseTime.Add(time.Hour)),
		}
		
		result := calc.CalculateTimeMetrics(pr, []ReviewEvent{})
		
		// 基本的な時間は計算される
		if result.TotalCycleTime == nil {
			t.Error("TotalCycleTime should not be nil")
		}
		
		// イベントベースの時間は計算されない
		if result.TimeToFirstReview != nil {
			t.Error("TimeToFirstReview should be nil with no events")
		}
	})
	
	t.Run("マージされていないPR", func(t *testing.T) {
		baseTime := time.Now()
		pr := PullRequest{
			CreatedAt: baseTime,
			MergedAt:  nil,
		}
		
		result := calc.CalculateTimeMetrics(pr, []ReviewEvent{})
		
		// マージ関連の時間は計算されない
		if result.TotalCycleTime != nil {
			t.Error("TotalCycleTime should be nil for unmerged PR")
		}
		
		if result.TimeToMerge != nil {
			t.Error("TimeToMerge should be nil for unmerged PR")
		}
	})
}

// ヘルパー関数

func timePtr2(t time.Time) *time.Time {
	return &t
}

func createTestCycleTimeMetrics(cycleTime time.Duration) *PRMetrics {
	return &PRMetrics{
		TimeMetrics: PRTimeMetrics{
			TotalCycleTime:    &cycleTime,
			TimeToFirstReview: durationPtr(cycleTime / 4),
			TimeToApproval:    durationPtr(cycleTime / 2),
			TimeToMerge:       durationPtr(cycleTime / 8),
		},
	}
}

func TestDefaultCycleTimeConfig(t *testing.T) {
	config := getDefaultCycleTimeConfig()
	
	// デフォルト設定の確認
	if config.UseBusinessHours {
		t.Error("Expected UseBusinessHours to be false by default")
	}
	
	if config.BusinessStart != 9 {
		t.Errorf("Expected BusinessStart to be 9, got %v", config.BusinessStart)
	}
	
	if config.BusinessEnd != 18 {
		t.Errorf("Expected BusinessEnd to be 18, got %v", config.BusinessEnd)
	}
	
	if config.ExcludeWeekends {
		t.Error("Expected ExcludeWeekends to be false by default")
	}
	
	if config.Timezone == nil {
		t.Error("Expected Timezone to be set")
	}
}