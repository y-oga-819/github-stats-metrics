package todo

import (
	"context"
)

// Repository はTodoデータアクセスの抽象化
type Repository interface {
	// GetTodos は全てのTodoを取得
	GetTodos(ctx context.Context) ([]Todo, error)
}