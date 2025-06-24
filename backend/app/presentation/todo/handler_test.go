package todo

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github-stats-metrics/domain/todo"
	usecase "github-stats-metrics/application/todo"
)

// MockTodoRepository は実際のTodoリポジトリのモック実装
type MockTodoRepository struct {
	mock.Mock
}

func (m *MockTodoRepository) GetTodos(ctx context.Context) ([]todo.Todo, error) {
	args := m.Called(ctx)
	return args.Get(0).([]todo.Todo), args.Error(1)
}

func TestHandler_GetTodos(t *testing.T) {
	tests := []struct {
		name           string
		setupMock      func(*MockTodoRepository)
		expectedStatus int
		expectedBody   interface{}
	}{
		{
			name: "正常にTodoを取得",
			setupMock: func(mockRepo *MockTodoRepository) {
				mockRepo.On("GetTodos", mock.Anything).Return([]todo.Todo{
					{Id: 1, Title: "Test Todo 1", Completed: false},
					{Id: 2, Title: "Test Todo 2", Completed: true},
				}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody: []todo.Todo{
				{Id: 1, Title: "Test Todo 1", Completed: false},
				{Id: 2, Title: "Test Todo 2", Completed: true},
			},
		},
		{
			name: "空のTodoリストを取得",
			setupMock: func(mockRepo *MockTodoRepository) {
				mockRepo.On("GetTodos", mock.Anything).Return([]todo.Todo{}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   []todo.Todo{},
		},
		{
			name: "UseCaseでエラーが発生",
			setupMock: func(mockRepo *MockTodoRepository) {
				mockRepo.On("GetTodos", mock.Anything).Return([]todo.Todo{}, errors.New("usecase error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody: map[string]interface{}{
				"error":   "Failed to get todos",
				"details": "usecase error",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// モックの設定
			mockRepo := new(MockTodoRepository)
			tt.setupMock(mockRepo)

			// UseCaseとハンドラー作成
			todoUseCase := usecase.NewUseCase(mockRepo)
			handler := NewHandler(todoUseCase)

			// HTTPリクエスト作成
			req := httptest.NewRequest(http.MethodGet, "/api/todos", nil)
			rec := httptest.NewRecorder()

			// テスト実行
			handler.GetTodos(rec, req)

			// ステータスコード確認
			assert.Equal(t, tt.expectedStatus, rec.Code)

			// Content-Type確認
			assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))

			// レスポンスボディ確認
			var actualBody interface{}
			err := json.NewDecoder(rec.Body).Decode(&actualBody)
			assert.NoError(t, err)

			expectedJSON, _ := json.Marshal(tt.expectedBody)
			actualJSON, _ := json.Marshal(actualBody)
			assert.JSONEq(t, string(expectedJSON), string(actualJSON))

			// モックの呼び出し確認
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestHandler_Integration(t *testing.T) {
	t.Run("HTTPハンドラーの基本的な動作確認", func(t *testing.T) {
		mockRepo := new(MockTodoRepository)
		expectedTodos := []todo.Todo{
			{Id: 1, Title: "Integration Test Todo", Completed: false},
		}
		mockRepo.On("GetTodos", mock.Anything).Return(expectedTodos, nil)

		todoUseCase := usecase.NewUseCase(mockRepo)
		handler := NewHandler(todoUseCase)
		req := httptest.NewRequest(http.MethodGet, "/api/todos", nil)
		rec := httptest.NewRecorder()

		handler.GetTodos(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))

		var todos []todo.Todo
		err := json.NewDecoder(rec.Body).Decode(&todos)
		assert.NoError(t, err)
		assert.Equal(t, expectedTodos, todos)

		mockRepo.AssertExpectations(t)
	})
}