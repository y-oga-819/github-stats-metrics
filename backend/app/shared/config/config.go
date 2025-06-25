package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// Config はアプリケーション設定を管理
type Config struct {
	GitHub   GitHubConfig
	Server   ServerConfig
	Security SecurityConfig
	Logging  LoggingConfig
}

// GitHubConfig はGitHub関連の設定
type GitHubConfig struct {
	Token        string
	Repositories []string
	Timeout      time.Duration
}

// ServerConfig はサーバー関連の設定
type ServerConfig struct {
	Port            int
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	ShutdownTimeout time.Duration
}

// SecurityConfig はセキュリティ関連の設定
type SecurityConfig struct {
	AllowedOrigins []string
}

// LoggingConfig はログ関連の設定
type LoggingConfig struct {
	Level  string
	Format string
}

// NewConfig は環境変数から設定を読み込み
func NewConfig() (*Config, error) {
	config := &Config{}
	
	// GitHub設定
	if err := config.loadGitHubConfig(); err != nil {
		return nil, fmt.Errorf("failed to load GitHub config: %w", err)
	}
	
	// サーバー設定
	if err := config.loadServerConfig(); err != nil {
		return nil, fmt.Errorf("failed to load server config: %w", err)
	}
	
	// セキュリティ設定
	if err := config.loadSecurityConfig(); err != nil {
		return nil, fmt.Errorf("failed to load security config: %w", err)
	}
	
	// ログ設定
	if err := config.loadLoggingConfig(); err != nil {
		return nil, fmt.Errorf("failed to load logging config: %w", err)
	}
	
	// 設定の検証
	if err := config.validate(); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}
	
	return config, nil
}

// loadGitHubConfig はGitHub関連の設定を読み込み
func (c *Config) loadGitHubConfig() error {
	// 必須: GitHubトークン
	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		return fmt.Errorf("GITHUB_TOKEN environment variable is required")
	}
	c.GitHub.Token = token
	
	// 必須: 対象リポジトリ
	repoStr := os.Getenv("GITHUB_GRAPHQL_SEARCH_QUERY_TARGET_REPOSITORIES")
	if repoStr == "" {
		return fmt.Errorf("GITHUB_GRAPHQL_SEARCH_QUERY_TARGET_REPOSITORIES environment variable is required")
	}
	c.GitHub.Repositories = strings.Split(repoStr, ",")
	
	// オプション: タイムアウト（デフォルト30秒）
	timeoutStr := os.Getenv("GITHUB_API_TIMEOUT")
	if timeoutStr == "" {
		c.GitHub.Timeout = 30 * time.Second
	} else {
		timeout, err := time.ParseDuration(timeoutStr)
		if err != nil {
			return fmt.Errorf("invalid GITHUB_API_TIMEOUT: %w", err)
		}
		c.GitHub.Timeout = timeout
	}
	
	return nil
}

// loadServerConfig はサーバー関連の設定を読み込み
func (c *Config) loadServerConfig() error {
	// オプション: ポート番号（デフォルト8080）
	portStr := os.Getenv("SERVER_PORT")
	if portStr == "" {
		c.Server.Port = 8080
	} else {
		port, err := strconv.Atoi(portStr)
		if err != nil {
			return fmt.Errorf("invalid SERVER_PORT: %w", err)
		}
		if port < 1 || port > 65535 {
			return fmt.Errorf("SERVER_PORT must be between 1 and 65535")
		}
		c.Server.Port = port
	}
	
	// オプション: 読み込みタイムアウト（デフォルト15秒）
	readTimeoutStr := os.Getenv("SERVER_READ_TIMEOUT")
	if readTimeoutStr == "" {
		c.Server.ReadTimeout = 15 * time.Second
	} else {
		timeout, err := time.ParseDuration(readTimeoutStr)
		if err != nil {
			return fmt.Errorf("invalid SERVER_READ_TIMEOUT: %w", err)
		}
		c.Server.ReadTimeout = timeout
	}
	
	// オプション: 書き込みタイムアウト（デフォルト15秒）
	writeTimeoutStr := os.Getenv("SERVER_WRITE_TIMEOUT")
	if writeTimeoutStr == "" {
		c.Server.WriteTimeout = 15 * time.Second
	} else {
		timeout, err := time.ParseDuration(writeTimeoutStr)
		if err != nil {
			return fmt.Errorf("invalid SERVER_WRITE_TIMEOUT: %w", err)
		}
		c.Server.WriteTimeout = timeout
	}
	
	// オプション: シャットダウンタイムアウト（デフォルト30秒）
	shutdownTimeoutStr := os.Getenv("SERVER_SHUTDOWN_TIMEOUT")
	if shutdownTimeoutStr == "" {
		c.Server.ShutdownTimeout = 30 * time.Second
	} else {
		timeout, err := time.ParseDuration(shutdownTimeoutStr)
		if err != nil {
			return fmt.Errorf("invalid SERVER_SHUTDOWN_TIMEOUT: %w", err)
		}
		c.Server.ShutdownTimeout = timeout
	}
	
	return nil
}

// loadSecurityConfig はセキュリティ関連の設定を読み込み
func (c *Config) loadSecurityConfig() error {
	// オプション: 許可オリジン（デフォルトlocalhost:3000）
	originsStr := os.Getenv("ALLOWED_ORIGINS")
	if originsStr == "" {
		c.Security.AllowedOrigins = []string{"http://localhost:3000"}
	} else {
		c.Security.AllowedOrigins = strings.Split(originsStr, ",")
	}
	
	return nil
}

// loadLoggingConfig はログ関連の設定を読み込み
func (c *Config) loadLoggingConfig() error {
	// オプション: ログレベル（デフォルトINFO）
	level := os.Getenv("LOG_LEVEL")
	if level == "" {
		c.Logging.Level = "INFO"
	} else {
		c.Logging.Level = strings.ToUpper(level)
	}
	
	// オプション: ログフォーマット（デフォルトJSON）
	format := os.Getenv("LOG_FORMAT")
	if format == "" {
		c.Logging.Format = "JSON"
	} else {
		c.Logging.Format = strings.ToUpper(format)
	}
	
	return nil
}

// validate は設定の妥当性を検証
func (c *Config) validate() error {
	// GitHubリポジトリの形式チェック
	for _, repo := range c.GitHub.Repositories {
		repo = strings.TrimSpace(repo)
		if !strings.Contains(repo, "/") {
			return fmt.Errorf("invalid repository format: %s (should be owner/repo)", repo)
		}
	}
	
	// ログレベルの検証
	validLogLevels := map[string]bool{
		"DEBUG": true,
		"INFO":  true,
		"WARN":  true,
		"ERROR": true,
	}
	if !validLogLevels[c.Logging.Level] {
		return fmt.Errorf("invalid log level: %s (valid: DEBUG, INFO, WARN, ERROR)", c.Logging.Level)
	}
	
	// ログフォーマットの検証
	validLogFormats := map[string]bool{
		"JSON": true,
		"TEXT": true,
	}
	if !validLogFormats[c.Logging.Format] {
		return fmt.Errorf("invalid log format: %s (valid: JSON, TEXT)", c.Logging.Format)
	}
	
	return nil
}

// GetListenAddress はサーバーのリスンアドレスを返す
func (c *Config) GetListenAddress() string {
	return fmt.Sprintf(":%d", c.Server.Port)
}

// IsDebugMode はデバッグモードかどうかを判定
func (c *Config) IsDebugMode() bool {
	return c.Logging.Level == "DEBUG"
}

// GetCleanRepositories は前後の空白を除去したリポジトリリストを返す
func (c *Config) GetCleanRepositories() []string {
	cleaned := make([]string, len(c.GitHub.Repositories))
	for i, repo := range c.GitHub.Repositories {
		cleaned[i] = strings.TrimSpace(repo)
	}
	return cleaned
}