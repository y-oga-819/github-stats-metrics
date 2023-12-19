package main

import (
	"log"

	"github-stats-metrics/server"

	"github.com/joho/godotenv"
)

func main() {
	loadErr := godotenv.Load()
	if loadErr != nil {
		log.Fatal("Error loading .env file")
	}

	// server パッケージから StartWebServer を呼び出す
	err := server.StartWebServer()
	if err != nil {
		log.Fatal(err)
	}
}
