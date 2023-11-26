package main

import (
	"log"

	"github-stats-metrics/server"
)

func main() {
	// server パッケージから StartWebServer を呼び出す
	err := server.StartWebServer()
	if err != nil {
		log.Fatal(err)
	}
}
