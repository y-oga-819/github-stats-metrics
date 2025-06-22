package todo

import (
	"context"
	"log"

	domain "github-stats-metrics/domain/todo"
)

// UseCase はTodo関連のビジネスロジックを統括
type UseCase struct {
	repo domain.Repository
}

// NewUseCase はUseCaseのコンストラクタ（依存性注入）
func NewUseCase(repo domain.Repository) *UseCase {
	return &UseCase{
		repo: repo,
	}
}

// GetTodos はTodosを取得し、ビジネスルールを適用
func (uc *UseCase) GetTodos(ctx context.Context) ([]domain.Todo, error) {
	log.Printf("Fetching todos")

	// リポジトリから取得（抽象に依存）
	todos, err := uc.repo.GetTodos(ctx)
	if err != nil {
		return nil, err
	}

	log.Printf("Retrieved %d todos", len(todos))
	return todos, nil
}