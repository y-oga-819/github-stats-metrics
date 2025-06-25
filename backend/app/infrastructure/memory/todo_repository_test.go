package memory

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github-stats-metrics/domain/todo"
)

func TestTodoRepository_GetAllTodos(t *testing.T) {
	tests := []struct {
		name     string
		expected []todo.Todo
	}{
		{
			name: "Todoリストを取得",
			expected: []todo.Todo{
				{
					Id:        1,
					Title:     "チャーハン作るよ！",
					Completed: true,
				},
				{
					Id:        2,
					Title:     "豚肉も入れるよ！",
					Completed: false,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := NewTodoRepository()
			ctx := context.Background()

			todos, err := repo.GetTodos(ctx)

			assert.NoError(t, err)
			assert.Equal(t, len(tt.expected), len(todos))

			for i, expectedTodo := range tt.expected {
				assert.Equal(t, expectedTodo.Id, todos[i].Id)
				assert.Equal(t, expectedTodo.Title, todos[i].Title)
				assert.Equal(t, expectedTodo.Completed, todos[i].Completed)
			}
		})
	}
}

func TestTodoRepository_Integration(t *testing.T) {
	t.Run("リポジトリの基本的な動作確認", func(t *testing.T) {
		repo := NewTodoRepository()
		ctx := context.Background()

		// データ取得
		todos, err := repo.GetTodos(ctx)
		assert.NoError(t, err)
		assert.NotEmpty(t, todos)

		// データの一貫性確認
		for _, todo := range todos {
			assert.NotZero(t, todo.Id)
			assert.NotEmpty(t, todo.Title)
		}
	})
}