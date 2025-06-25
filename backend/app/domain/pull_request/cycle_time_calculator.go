package pull_request

import (
	"sort"
	"time"
)

// CycleTimeCalculator はサイクルタイムを計算するサービス
type CycleTimeCalculator struct {
	config CycleTimeConfig
}

// CycleTimeConfig はサイクルタイム計算の設定
type CycleTimeConfig struct {
	// 営業時間の設定（時間計算に営業時間を考慮するか）
	UseBusinessHours bool
	BusinessStart    int // 9時
	BusinessEnd      int // 18時
	
	// 営業日の設定
	ExcludeWeekends  bool
	ExcludeHolidays  bool
	
	// タイムゾーン
	Timezone *time.Location
}

// NewCycleTimeCalculator は新しいサイクルタイム計算機を作成
func NewCycleTimeCalculator() *CycleTimeCalculator {
	return &CycleTimeCalculator{
		config: getDefaultCycleTimeConfig(),
	}
}

// getDefaultCycleTimeConfig はデフォルトの設定を返す
func getDefaultCycleTimeConfig() CycleTimeConfig {
	loc, _ := time.LoadLocation("Asia/Tokyo")
	return CycleTimeConfig{
		UseBusinessHours: false, // デフォルトは24時間計算
		BusinessStart:    9,
		BusinessEnd:      18,
		ExcludeWeekends:  false,
		ExcludeHolidays:  false,
		Timezone:        loc,
	}
}

// CalculateTimeMetrics はPRの時間メトリクスを計算
func (calc *CycleTimeCalculator) CalculateTimeMetrics(pr PullRequest, reviewEvents []ReviewEvent) PRTimeMetrics {
	timeMetrics := PRTimeMetrics{
		CreatedHour: pr.CreatedAt.Hour(),
	}
	
	if pr.MergedAt != nil {
		mergedHour := pr.MergedAt.Hour()
		timeMetrics.MergedHour = &mergedHour
	}
	
	// レビューイベントを時系列でソート
	sortedEvents := make([]ReviewEvent, len(reviewEvents))
	copy(sortedEvents, reviewEvents)
	sort.Slice(sortedEvents, func(i, j int) bool {
		return sortedEvents[i].CreatedAt.Before(sortedEvents[j].CreatedAt)
	})
	
	// 各段階の時間を計算
	timeMetrics.TimeToFirstReview = calc.calculateTimeToFirstReview(pr, sortedEvents)
	timeMetrics.TimeToApproval = calc.calculateTimeToApproval(pr, sortedEvents)
	timeMetrics.TimeToMerge = calc.calculateTimeToMerge(pr, sortedEvents)
	timeMetrics.TotalCycleTime = calc.calculateTotalCycleTime(pr)
	timeMetrics.ReviewWaitTime = calc.calculateReviewWaitTime(pr, sortedEvents)
	timeMetrics.ReviewActiveTime = calc.calculateReviewActiveTime(sortedEvents)
	timeMetrics.FirstCommitToMerge = calc.calculateFirstCommitToMerge(pr)
	
	return timeMetrics
}

// calculateTimeToFirstReview は初回レビューまでの時間を計算
func (calc *CycleTimeCalculator) calculateTimeToFirstReview(pr PullRequest, events []ReviewEvent) *time.Duration {
	// 最初のレビューイベントを探す
	for _, event := range events {
		if event.Type == ReviewEventTypeCommented || 
		   event.Type == ReviewEventTypeApproved || 
		   event.Type == ReviewEventTypeChangesRequested {
			duration := calc.calculateDuration(pr.CreatedAt, event.CreatedAt)
			return &duration
		}
	}
	
	// イベントがない場合は既存のFirstReviewedを使用
	if pr.FirstReviewed != nil {
		duration := calc.calculateDuration(pr.CreatedAt, *pr.FirstReviewed)
		return &duration
	}
	
	return nil
}

// calculateTimeToApproval は承認までの時間を計算
func (calc *CycleTimeCalculator) calculateTimeToApproval(pr PullRequest, events []ReviewEvent) *time.Duration {
	// 最初の承認イベントを探す
	for _, event := range events {
		if event.Type == ReviewEventTypeApproved {
			duration := calc.calculateDuration(pr.CreatedAt, event.CreatedAt)
			return &duration
		}
	}
	
	// イベントがない場合は既存のLastApprovedを使用
	if pr.LastApproved != nil {
		duration := calc.calculateDuration(pr.CreatedAt, *pr.LastApproved)
		return &duration
	}
	
	return nil
}

// calculateTimeToMerge は承認からマージまでの時間を計算
func (calc *CycleTimeCalculator) calculateTimeToMerge(pr PullRequest, events []ReviewEvent) *time.Duration {
	if pr.MergedAt == nil {
		return nil
	}
	
	// 最後の承認イベントを探す
	var lastApproval *time.Time
	for i := len(events) - 1; i >= 0; i-- {
		if events[i].Type == ReviewEventTypeApproved {
			lastApproval = &events[i].CreatedAt
			break
		}
	}
	
	if lastApproval != nil {
		duration := calc.calculateDuration(*lastApproval, *pr.MergedAt)
		return &duration
	}
	
	// イベントがない場合は既存のLastApprovedを使用
	if pr.LastApproved != nil {
		duration := calc.calculateDuration(*pr.LastApproved, *pr.MergedAt)
		return &duration
	}
	
	return nil
}

// calculateTotalCycleTime は全体のサイクルタイムを計算
func (calc *CycleTimeCalculator) calculateTotalCycleTime(pr PullRequest) *time.Duration {
	if pr.MergedAt == nil {
		return nil
	}
	
	duration := calc.calculateDuration(pr.CreatedAt, *pr.MergedAt)
	return &duration
}

// calculateReviewWaitTime はレビュー待ち時間を計算
func (calc *CycleTimeCalculator) calculateReviewWaitTime(pr PullRequest, events []ReviewEvent) *time.Duration {
	if len(events) == 0 {
		return nil
	}
	
	totalWaitTime := time.Duration(0)
	lastEventTime := pr.CreatedAt
	
	for _, event := range events {
		// レビュー要求からレビュー実施までの時間
		if event.Type == ReviewEventTypeRequested {
			lastEventTime = event.CreatedAt
		} else if event.Type == ReviewEventTypeCommented || 
				  event.Type == ReviewEventTypeApproved || 
				  event.Type == ReviewEventTypeChangesRequested {
			waitTime := calc.calculateDuration(lastEventTime, event.CreatedAt)
			totalWaitTime += waitTime
			lastEventTime = event.CreatedAt
		}
	}
	
	return &totalWaitTime
}

// calculateReviewActiveTime はレビュー実施時間を計算（推定）
func (calc *CycleTimeCalculator) calculateReviewActiveTime(events []ReviewEvent) *time.Duration {
	if len(events) <= 1 {
		return nil
	}
	
	// レビューイベント間の時間から推定
	// これは簡易的な計算で、実際のレビュー時間とは異なる場合がある
	totalActiveTime := time.Duration(0)
	var lastReviewTime *time.Time
	
	for _, event := range events {
		if event.Type == ReviewEventTypeCommented || 
		   event.Type == ReviewEventTypeApproved || 
		   event.Type == ReviewEventTypeChangesRequested {
			if lastReviewTime != nil {
				// 連続するレビューイベント間の時間（最大2時間まで）
				duration := event.CreatedAt.Sub(*lastReviewTime)
				if duration <= 2*time.Hour {
					totalActiveTime += duration
				}
			}
			lastReviewTime = &event.CreatedAt
		}
	}
	
	// 最低5分は見積もる
	if totalActiveTime < 5*time.Minute {
		minTime := 5 * time.Minute
		return &minTime
	}
	
	return &totalActiveTime
}

// calculateFirstCommitToMerge は最初のコミットからマージまでの時間を計算
func (calc *CycleTimeCalculator) calculateFirstCommitToMerge(pr PullRequest) *time.Duration {
	if pr.MergedAt == nil {
		return nil
	}
	
	// PR作成時刻を最初のコミット時刻として近似
	// 実際の実装では、GitHubのコミット情報から最初のコミット時刻を取得する
	duration := calc.calculateDuration(pr.CreatedAt, *pr.MergedAt)
	return &duration
}

// calculateDuration は営業時間を考慮して時間を計算
func (calc *CycleTimeCalculator) calculateDuration(start, end time.Time) time.Duration {
	if !calc.config.UseBusinessHours {
		return end.Sub(start)
	}
	
	// 営業時間のみを計算（簡易版）
	return calc.calculateBusinessHours(start, end)
}

// calculateBusinessHours は営業時間のみを計算
func (calc *CycleTimeCalculator) calculateBusinessHours(start, end time.Time) time.Duration {
	// タイムゾーンを設定
	start = start.In(calc.config.Timezone)
	end = end.In(calc.config.Timezone)
	
	if start.After(end) {
		return 0
	}
	
	totalDuration := time.Duration(0)
	current := start
	
	for current.Before(end) {
		// 当日の営業時間を計算
		dayStart := time.Date(current.Year(), current.Month(), current.Day(), 
			calc.config.BusinessStart, 0, 0, 0, calc.config.Timezone)
		dayEnd := time.Date(current.Year(), current.Month(), current.Day(), 
			calc.config.BusinessEnd, 0, 0, 0, calc.config.Timezone)
		
		// 週末をスキップ
		if calc.config.ExcludeWeekends && (current.Weekday() == time.Saturday || current.Weekday() == time.Sunday) {
			current = current.AddDate(0, 0, 1)
			current = time.Date(current.Year(), current.Month(), current.Day(), 0, 0, 0, 0, calc.config.Timezone)
			continue
		}
		
		// その日の作業時間を計算
		workStart := current
		if workStart.Before(dayStart) {
			workStart = dayStart
		}
		
		workEnd := end
		if workEnd.After(dayEnd) {
			workEnd = dayEnd
		}
		
		if workStart.Before(workEnd) && workStart.Day() == workEnd.Day() {
			totalDuration += workEnd.Sub(workStart)
		}
		
		// 次の日に進む
		current = current.AddDate(0, 0, 1)
		current = time.Date(current.Year(), current.Month(), current.Day(), 0, 0, 0, 0, calc.config.Timezone)
	}
	
	return totalDuration
}

// CalculateCycleTimeStatistics は複数PRのサイクルタイム統計を計算
func (calc *CycleTimeCalculator) CalculateCycleTimeStatistics(metrics []*PRMetrics) *CycleTimeStatistics {
	if len(metrics) == 0 {
		return &CycleTimeStatistics{}
	}
	
	stats := &CycleTimeStatistics{
		TotalPRs: len(metrics),
	}
	
	var totalCycleTimes []time.Duration
	var reviewTimes []time.Duration
	var approvalTimes []time.Duration
	var mergeTimes []time.Duration
	
	for _, metric := range metrics {
		if metric.TimeMetrics.TotalCycleTime != nil {
			totalCycleTimes = append(totalCycleTimes, *metric.TimeMetrics.TotalCycleTime)
		}
		if metric.TimeMetrics.TimeToFirstReview != nil {
			reviewTimes = append(reviewTimes, *metric.TimeMetrics.TimeToFirstReview)
		}
		if metric.TimeMetrics.TimeToApproval != nil {
			approvalTimes = append(approvalTimes, *metric.TimeMetrics.TimeToApproval)
		}
		if metric.TimeMetrics.TimeToMerge != nil {
			mergeTimes = append(mergeTimes, *metric.TimeMetrics.TimeToMerge)
		}
	}
	
	// 統計値を計算
	stats.TotalCycleTime = calc.calculateDurationStatistics(totalCycleTimes)
	stats.TimeToFirstReview = calc.calculateDurationStatistics(reviewTimes)
	stats.TimeToApproval = calc.calculateDurationStatistics(approvalTimes)
	stats.TimeToMerge = calc.calculateDurationStatistics(mergeTimes)
	
	return stats
}

// calculateDurationStatistics は時間の統計値を計算
func (calc *CycleTimeCalculator) calculateDurationStatistics(durations []time.Duration) DurationStatistics {
	if len(durations) == 0 {
		return DurationStatistics{}
	}
	
	// ソート
	sort.Slice(durations, func(i, j int) bool {
		return durations[i] < durations[j]
	})
	
	stats := DurationStatistics{
		Count: len(durations),
		Min:   durations[0],
		Max:   durations[len(durations)-1],
	}
	
	// 平均値
	total := time.Duration(0)
	for _, d := range durations {
		total += d
	}
	stats.Average = total / time.Duration(len(durations))
	
	// 中央値
	if len(durations)%2 == 0 {
		mid := len(durations) / 2
		stats.Median = (durations[mid-1] + durations[mid]) / 2
	} else {
		stats.Median = durations[len(durations)/2]
	}
	
	// パーセンタイル
	stats.P75 = durations[int(float64(len(durations))*0.75)]
	stats.P90 = durations[int(float64(len(durations))*0.90)]
	stats.P95 = durations[int(float64(len(durations))*0.95)]
	
	return stats
}

// CycleTimeStatistics はサイクルタイムの統計情報
type CycleTimeStatistics struct {
	TotalPRs int `json:"totalPRs"`
	
	TotalCycleTime    DurationStatistics `json:"totalCycleTime"`
	TimeToFirstReview DurationStatistics `json:"timeToFirstReview"`
	TimeToApproval    DurationStatistics `json:"timeToApproval"`
	TimeToMerge       DurationStatistics `json:"timeToMerge"`
}

// DurationStatistics は時間の統計情報
type DurationStatistics struct {
	Count   int           `json:"count"`
	Average time.Duration `json:"average"`
	Median  time.Duration `json:"median"`
	Min     time.Duration `json:"min"`
	Max     time.Duration `json:"max"`
	P75     time.Duration `json:"p75"`
	P90     time.Duration `json:"p90"`
	P95     time.Duration `json:"p95"`
}