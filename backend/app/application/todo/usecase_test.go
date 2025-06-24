package todo

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github-stats-metrics/domain/todo"
)

// MockTodoRepository はTodoRepositoryのモック実装
type MockTodoRepository struct {
	mock.Mock
}

func (m *MockTodoRepository) GetTodos(ctx context.Context) ([]todo.Todo, error) {
	args := m.Called(ctx)
	return args.Get(0).([]todo.Todo), args.Error(1)
}

func TestUseCase_GetTodos(t *testing.T) {
	tests := []struct {
		name           string
		setupMock      func(*MockTodoRepository)
		expectedTodos  []todo.Todo
		expectedError  bool
		expectedErrMsg string
	}{
		{
			name: "正常にTodoを取得",
			setupMock: func(mockRepo *MockTodoRepository) {
				mockRepo.On("GetTodos", mock.Anything).Return([]todo.Todo{
					{Id: 1, Title: "Test Todo 1", Completed: false},
					{Id: 2, Title: "Test Todo 2", Completed: true},
				}, nil)
			},
			expectedTodos: []todo.Todo{
				{Id: 1, Title: "Test Todo 1", Completed: false},
				{Id: 2, Title: "Test Todo 2", Completed: true},
			},
			expectedError: false,
		},
		{
			name: "リポジトリでエラーが発生",
			setupMock: func(mockRepo *MockTodoRepository) {
				mockRepo.On("GetTodos", mock.Anything).Return([]todo.Todo{}, errors.New("repository error"))
			},
			expectedTodos:  nil,
			expectedError:  true,
			expectedErrMsg: "repository error",
		},
		{
			name: "空のTodoリストを取得",
			setupMock: func(mockRepo *MockTodoRepository) {
				mockRepo.On("GetTodos", mock.Anything).Return([]todo.Todo{}, nil)
			},
			expectedTodos: []todo.Todo{},
			expectedError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// モックの設定
			mockRepo := new(MockTodoRepository)
			tt.setupMock(mockRepo)

			// UseCase作成
			useCase := NewUseCase(mockRepo)
			ctx := context.Background()

			// テスト実行
			todos, err := useCase.GetTodos(ctx)

			// 結果の検証
			if tt.expectedError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErrMsg)
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, tt.expectedTodos, todos)

			// モックの呼び出し確認
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestUseCase_Integration(t *testing.T) {
	t.Run("UseCaseの基本的な動作確認", func(t *testing.T) {
		mockRepo := new(MockTodoRepository)
		
		// 正常なレスポンスを設定
		expectedTodos := []todo.Todo{
			{Id: 1, Title: "Integration Test Todo", Completed: false},
		}
		mockRepo.On("GetTodos", mock.Anything).Return(expectedTodos, nil)

		useCase := NewUseCase(mockRepo)
		ctx := context.Background()

		todos, err := useCase.GetTodos(ctx)

		assert.NoError(t, err)
		assert.Equal(t, expectedTodos, todos)
		mockRepo.AssertExpectations(t)
	})
}