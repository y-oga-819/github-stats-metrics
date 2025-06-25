package pull_request

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestPullRequest_ReviewTime(t *testing.T) {
	tests := []struct {
		name     string
		pr       PullRequest
		expected *time.Duration
	}{
		{
			name: "レビュー時間が計算できる場合",
			pr: PullRequest{
				CreatedAt:     parseTime("2023-01-01T10:00:00Z"),
				FirstReviewed: timePtr(parseTime("2023-01-01T12:00:00Z")),
			},
			expected: durationPtr(2 * time.Hour),
		},
		{
			name: "FirstReviewedがnilの場合",
			pr: PullRequest{
				CreatedAt:     parseTime("2023-01-01T10:00:00Z"),
				FirstReviewed: nil,
			},
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.pr.ReviewTime()

			if tt.expected == nil {
				assert.Nil(t, result)
			} else {
				assert.NotNil(t, result)
				assert.Equal(t, *tt.expected, *result)
			}
		})
	}
}

func TestPullRequest_ApprovalTime(t *testing.T) {
	tests := []struct {
		name     string
		pr       PullRequest
		expected *time.Duration
	}{
		{
			name: "承認時間が計算できる場合",
			pr: PullRequest{
				FirstReviewed: timePtr(parseTime("2023-01-01T12:00:00Z")),
				LastApproved:  timePtr(parseTime("2023-01-01T14:00:00Z")),
			},
			expected: durationPtr(2 * time.Hour),
		},
		{
			name: "FirstReviewedがnilの場合",
			pr: PullRequest{
				FirstReviewed: nil,
				LastApproved:  timePtr(parseTime("2023-01-01T14:00:00Z")),
			},
			expected: nil,
		},
		{
			name: "LastApprovedがnilの場合",
			pr: PullRequest{
				FirstReviewed: timePtr(parseTime("2023-01-01T12:00:00Z")),
				LastApproved:  nil,
			},
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.pr.ApprovalTime()

			if tt.expected == nil {
				assert.Nil(t, result)
			} else {
				assert.NotNil(t, result)
				assert.Equal(t, *tt.expected, *result)
			}
		})
	}
}

func TestPullRequest_MergeTime(t *testing.T) {
	tests := []struct {
		name     string
		pr       PullRequest
		expected *time.Duration
	}{
		{
			name: "マージ時間が計算できる場合",
			pr: PullRequest{
				LastApproved: timePtr(parseTime("2023-01-01T14:00:00Z")),
				MergedAt:     timePtr(parseTime("2023-01-01T15:00:00Z")),
			},
			expected: durationPtr(1 * time.Hour),
		},
		{
			name: "LastApprovedがnilの場合",
			pr: PullRequest{
				LastApproved: nil,
				MergedAt:     timePtr(parseTime("2023-01-01T15:00:00Z")),
			},
			expected: nil,
		},
		{
			name: "MergedAtがnilの場合",
			pr: PullRequest{
				LastApproved: timePtr(parseTime("2023-01-01T14:00:00Z")),
				MergedAt:     nil,
			},
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.pr.MergeTime()

			if tt.expected == nil {
				assert.Nil(t, result)
			} else {
				assert.NotNil(t, result)
				assert.Equal(t, *tt.expected, *result)
			}
		})
	}
}

func TestPullRequest_StatusChecks(t *testing.T) {
	tests := []struct {
		name     string
		pr       PullRequest
		isReviewed bool
		isApproved bool
		isMerged   bool
	}{
		{
			name: "完全なPRライフサイクル",
			pr: PullRequest{
				FirstReviewed: timePtr(parseTime("2023-01-01T12:00:00Z")),
				LastApproved:  timePtr(parseTime("2023-01-01T14:00:00Z")),
				MergedAt:      timePtr(parseTime("2023-01-01T15:00:00Z")),
			},
			isReviewed: true,
			isApproved: true,
			isMerged:   true,
		},
		{
			name: "レビューのみのPR",
			pr: PullRequest{
				FirstReviewed: timePtr(parseTime("2023-01-01T12:00:00Z")),
			},
			isReviewed: true,
			isApproved: false,
			isMerged:   false,
		},
		{
			name: "レビューなしのPR",
			pr: PullRequest{},
			isReviewed: false,
			isApproved: false,
			isMerged:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.isReviewed, tt.pr.IsReviewed())
			assert.Equal(t, tt.isApproved, tt.pr.IsApproved())
			assert.Equal(t, tt.isMerged, tt.pr.IsMerged())
		})
	}
}

// ヘルパー関数
func parseTime(s string) time.Time {
	t, _ := time.Parse(time.RFC3339, s)
	return t
}

func timePtr(t time.Time) *time.Time {
	return &t
}

func durationPtr(d time.Duration) *time.Duration {
	return &d
}