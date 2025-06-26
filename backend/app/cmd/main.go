package main

import (
	"context"
	"log"
	"os"

	"github-stats-metrics/server"
	"github-stats-metrics/shared/config"
	"github-stats-metrics/shared/logging"
	"github-stats-metrics/shared/metrics"
	"github-stats-metrics/shared/monitoring"

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

	// 構造化ロガーを初期化
	logLevel, err := logging.ParseLogLevel(cfg.Logging.Level)
	if err != nil {
		log.Printf("Invalid log level %s, using INFO", cfg.Logging.Level)
		logLevel = logging.INFO
	}
	
	logger := logging.NewStructuredLogger(logLevel, "github-stats-metrics", "1.0.0")
	ctx := context.Background()

	// メトリクスコレクターを初期化
	metricsCollector := metrics.NewMetricsCollector()
	if err := metricsCollector.Register(); err != nil {
		logger.Fatal(ctx, "Failed to register metrics", err)
	}

	// アプリケーション情報を設定
	metricsCollector.SetApplicationInfo("1.0.0", "2024-06-24", "latest")

	// ヘルスチェッカーを初期化
	healthChecker := monitoring.NewHealthChecker(cfg)

	logger.Info(ctx, "Starting GitHub Stats Metrics application", map[string]interface{}{
		"version":      "1.0.0",
		"log_level":    cfg.Logging.Level,
		"listen_address": cfg.GetListenAddress(),
	})

	// server パッケージから StartWebServer を呼び出す
	// TODO: 実際のデータベース接続を設定する
	err = server.StartWebServer(cfg, logger, metricsCollector, healthChecker, nil)
	if err != nil {
		logger.Fatal(ctx, "Failed to start web server", err)
		os.Exit(1)
	}
}
