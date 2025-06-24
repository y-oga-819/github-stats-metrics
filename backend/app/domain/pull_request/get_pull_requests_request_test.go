package pull_request

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetPullRequestsRequest_Validate(t *testing.T) {
	tests := []struct {
		name    string
		req     GetPullRequestsRequest
		wantErr bool
		errMsg  string
	}{
		{
			name: "有効なリクエスト",
			req: GetPullRequestsRequest{
				StartDate:  "2023-01-01",
				EndDate:    "2023-01-31",
				Developers: []string{"developer1", "developer2"},
			},
			wantErr: false,
		},
		{
			name: "開始日が空",
			req: GetPullRequestsRequest{
				StartDate:  "",
				EndDate:    "2023-01-31",
				Developers: []string{"developer1"},
			},
			wantErr: true,
			errMsg:  "start date is required",
		},
		{
			name: "終了日が空",
			req: GetPullRequestsRequest{
				StartDate:  "2023-01-01",
				EndDate:    "",
				Developers: []string{"developer1"},
			},
			wantErr: true,
			errMsg:  "end date is required",
		},
		{
			name: "開発者リストが空",
			req: GetPullRequestsRequest{
				StartDate:  "2023-01-01",
				EndDate:    "2023-01-31",
				Developers: []string{},
			},
			wantErr: true,
			errMsg:  "at least one developer is required",
		},
		{
			name: "開発者リストがnil",
			req: GetPullRequestsRequest{
				StartDate:  "2023-01-01",
				EndDate:    "2023-01-31",
				Developers: nil,
			},
			wantErr: true,
			errMsg:  "at least one developer is required",
		},
		{
			name: "日付形式が無効（開始日）",
			req: GetPullRequestsRequest{
				StartDate:  "invalid-date",
				EndDate:    "2023-01-31",
				Developers: []string{"developer1"},
			},
			wantErr: true,
			errMsg:  "invalid start date format",
		},
		{
			name: "日付形式が無効（終了日）",
			req: GetPullRequestsRequest{
				StartDate:  "2023-01-01",
				EndDate:    "invalid-date",
				Developers: []string{"developer1"},
			},
			wantErr: true,
			errMsg:  "invalid end date format",
		},
		{
			name: "開始日が終了日より後",
			req: GetPullRequestsRequest{
				StartDate:  "2023-01-31",
				EndDate:    "2023-01-01",
				Developers: []string{"developer1"},
			},
			wantErr: true,
			errMsg:  "start date must be before end date",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.req.Validate()

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestGetPullRequestsRequest_GetDates(t *testing.T) {
	tests := []struct {
		name      string
		startDate string
		endDate   string
		wantErr   bool
	}{
		{
			name:      "有効な日付",
			startDate: "2023-01-01",
			endDate:   "2023-01-31",
			wantErr:   false,
		},
		{
			name:      "無効な開始日",
			startDate: "invalid-date",
			endDate:   "2023-01-31",
			wantErr:   true,
		},
		{
			name:      "無効な終了日",
			startDate: "2023-01-01",
			endDate:   "invalid-date",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := GetPullRequestsRequest{
				StartDate: tt.startDate,
				EndDate:   tt.endDate,
			}

			_, err1 := req.GetStartDate()
			_, err2 := req.GetEndDate()

			if tt.wantErr {
				assert.True(t, err1 != nil || err2 != nil)
			} else {
				assert.NoError(t, err1)
				assert.NoError(t, err2)
			}
		})
	}
}