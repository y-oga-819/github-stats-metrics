package pull_request

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github-stats-metrics/domain/pull_request"
)

// MockPullRequestRepository はPullRequestRepositoryのモック実装
type MockPullRequestRepository struct {
	mock.Mock
}

func (m *MockPullRequestRepository) GetPullRequests(ctx context.Context, req pull_request.GetPullRequestsRequest) ([]pull_request.PullRequest, error) {
	args := m.Called(ctx, req)
	return args.Get(0).([]pull_request.PullRequest), args.Error(1)
}

func (m *MockPullRequestRepository) GetPullRequestByID(ctx context.Context, id string) (*pull_request.PullRequest, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*pull_request.PullRequest), args.Error(1)
}

func (m *MockPullRequestRepository) GetRepositories(ctx context.Context) ([]string, error) {
	args := m.Called(ctx)
	return args.Get(0).([]string), args.Error(1)
}

func (m *MockPullRequestRepository) GetDevelopers(ctx context.Context, repositories []string) ([]string, error) {
	args := m.Called(ctx, repositories)
	return args.Get(0).([]string), args.Error(1)
}

func TestUseCase_GetPullRequests(t *testing.T) {
	tests := []struct {
		name          string
		request       pull_request.GetPullRequestsRequest
		setupMock     func(*MockPullRequestRepository)
		expectedPRs   []pull_request.PullRequest
		expectedError bool
		errorType     string
	}{
		{
			name: "正常にPRを取得",
			request: pull_request.GetPullRequestsRequest{
				StartDate:  "2023-01-01",
				EndDate:    "2023-01-31",
				Developers: []string{"developer1"},
			},
			setupMock: func(mockRepo *MockPullRequestRepository) {
				mockRepo.On("GetPullRequests", mock.Anything, mock.Anything).Return([]pull_request.PullRequest{
					{
						ID:       "1",
						Title:    "Test PR",
						URL:      "https://github.com/test/repo/pull/1",
						MergedAt: timePtr(parseTime("2023-01-15T15:00:00Z")), // マージ済みにする
						Author: pull_request.Author{
							Login: "developer1",
						},
					},
				}, nil)
			},
			expectedPRs: []pull_request.PullRequest{
				{
					ID:       "1",
					Title:    "Test PR",
					URL:      "https://github.com/test/repo/pull/1",
					MergedAt: timePtr(parseTime("2023-01-15T15:00:00Z")),
					Author: pull_request.Author{
						Login: "developer1",
					},
				},
			},
			expectedError: false,
		},
		{
			name: "無効なリクエスト（バリデーションエラー）",
			request: pull_request.GetPullRequestsRequest{
				StartDate:  "",
				EndDate:    "2023-01-31",
				Developers: []string{"developer1"},
			},
			setupMock: func(mockRepo *MockPullRequestRepository) {
				// バリデーションエラーなのでリポジトリは呼ばれない
			},
			expectedPRs:   nil,
			expectedError: true,
			errorType:     ErrorTypeValidation,
		},
		{
			name: "リポジトリでエラーが発生",
			request: pull_request.GetPullRequestsRequest{
				StartDate:  "2023-01-01",
				EndDate:    "2023-01-31",
				Developers: []string{"developer1"},
			},
			setupMock: func(mockRepo *MockPullRequestRepository) {
				mockRepo.On("GetPullRequests", mock.Anything, mock.Anything).Return([]pull_request.PullRequest{}, errors.New("repository error"))
			},
			expectedPRs:   nil,
			expectedError: true,
			errorType:     ErrorTypeRepository,
		},
		{
			name: "空のPRリストを取得",
			request: pull_request.GetPullRequestsRequest{
				StartDate:  "2023-01-01",
				EndDate:    "2023-01-31",
				Developers: []string{"developer1"},
			},
			setupMock: func(mockRepo *MockPullRequestRepository) {
				mockRepo.On("GetPullRequests", mock.Anything, mock.Anything).Return([]pull_request.PullRequest{}, nil)
			},
			expectedPRs:   nil,
			expectedError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// モックの設定
			mockRepo := new(MockPullRequestRepository)
			tt.setupMock(mockRepo)

			// UseCase作成
			useCase := NewUseCase(mockRepo)
			ctx := context.Background()

			// テスト実行
			prs, err := useCase.GetPullRequests(ctx, tt.request)

			// 結果の検証
			if tt.expectedError {
				assert.Error(t, err)
				
				// エラータイプの確認
				if useCaseErr, ok := err.(UseCaseError); ok {
					assert.Equal(t, tt.errorType, useCaseErr.Type)
				}
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, tt.expectedPRs, prs)

			// モックの呼び出し確認
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestUseCaseError(t *testing.T) {
	t.Run("UseCaseErrorの動作確認", func(t *testing.T) {
		err := UseCaseError{
			Type:    ErrorTypeValidation,
			Message: "validation failed",
			Cause:   errors.New("original error"),
		}

		assert.Contains(t, err.Error(), "validation failed")
		assert.Equal(t, ErrorTypeValidation, err.Type)
		assert.Equal(t, "validation failed", err.Message)
		assert.Equal(t, "original error", err.Cause.Error())
	})

	t.Run("UseCaseError without cause", func(t *testing.T) {
		err := UseCaseError{
			Type:    ErrorTypeBusinessRule,
			Message: "business rule violation",
			Cause:   nil,
		}

		assert.Contains(t, err.Error(), "business rule violation")
	})
}

// ヘルパー関数
func parseTime(s string) time.Time {
	t, _ := time.Parse(time.RFC3339, s)
	return t
}

func timePtr(t time.Time) *time.Time {
	return &t
}