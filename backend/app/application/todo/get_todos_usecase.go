package todo

import (
	"net/http"

	todoDomain "github-stats-metrics/domain/todo"
	todoPresenter "github-stats-metrics/presentation/todo"
)

func GetTodos(w http.ResponseWriter, r *http.Request) {
	// 返却したい値を構造体で定義
	todo1 := todoDomain.Todo{
		Id:        1,
		Title:     "チャーハン作るよ！",
		Completed: true,
	}
	todo2 := todoDomain.Todo{
		Id:        2,
		Title:     "豚肉も入れるよ！",
		Completed: false,
	}

	todos := []todoDomain.Todo{todo1, todo2}

	todoPresenter.Success(w, todos)
}
