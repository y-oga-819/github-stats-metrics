package memory

import (
	"context"

	domain "github-stats-metrics/domain/todo"
)

// todoRepository はメモリ内Todoデータの実装
type todoRepository struct{}

// NewTodoRepository はメモリ内Todo Repositoryを作成
func NewTodoRepository() domain.Repository {
	return &todoRepository{}
}

// GetTodos は全てのTodoを取得（ハードコードされたデータ）
func (r *todoRepository) GetTodos(ctx context.Context) ([]domain.Todo, error) {
	// 返却したい値を構造体で定義
	todo1 := domain.Todo{
		Id:        1,
		Title:     "チャーハン作るよ！",
		Completed: true,
	}
	todo2 := domain.Todo{
		Id:        2,
		Title:     "豚肉も入れるよ！",
		Completed: false,
	}

	todos := []domain.Todo{todo1, todo2}
	return todos, nil
}