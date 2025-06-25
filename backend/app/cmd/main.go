package main

import (
	"log"

	"github-stats-metrics/server"
	"github-stats-metrics/shared/config"

	"github.com/joho/godotenv"
)

func main() {
	loadErr := godotenv.Load()
	if loadErr != nil {
		log.Printf("Warning: Could not load .env file: %v", loadErr)
	}

	// 設定を読み込み
	cfg, err := config.NewConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// server パッケージから StartWebServer を呼び出す
	err = server.StartWebServer(cfg)
	if err != nil {
		log.Fatal(err)
	}
}
