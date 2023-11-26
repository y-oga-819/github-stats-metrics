package server

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
)

// 頭文字を大文字にするとパッケージ外部から呼び出しできる
func StartWebServer() error {
	fmt.Println("Start Web Server!")
	r := mux.NewRouter().StrictSlash(true)

	// URL に呼び出したい関数を登録する
	r.HandleFunc("/api/todos", getTodos).Methods("GET")

	// ポートを指定してサーバーを起動する
	return http.ListenAndServe(":8080", r)
}

type Todo struct {
	Id        int
	Title     string
	Completed bool
}

func getTodos(w http.ResponseWriter, r *http.Request) {
	// フロントエンドとバックエンドのポートが違うので許可しておく
	// （すべてを許可する設定にしているので、本番ではより制限を厳しくしておくように）
	w.Header().Set("Access-Control-Allow-Origin", "*")

	// 返却したい値を構造体で定義
	todo1 := Todo{
		Id:        1,
		Title:     "チャーハン作るよ！",
		Completed: true,
	}
	todo2 := Todo{
		Id:        2,
		Title:     "豚肉も入れるよ！",
		Completed: false,
	}

	todos := []Todo{todo1, todo2}
	// JSON にして返却
	responseBody, err := json.Marshal(todos)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write(responseBody)
}
