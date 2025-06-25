package logger

import (
	"fmt"
	"log"
	"os"
	"strings"
)

// LogLevel はログレベルを表す型
type LogLevel int

const (
	DEBUG LogLevel = iota
	INFO
	WARN
	ERROR
)

var logLevelNames = map[LogLevel]string{
	DEBUG: "DEBUG",
	INFO:  "INFO",
	WARN:  "WARN",
	ERROR: "ERROR",
}

// LevelLogger はログレベル制御機能付きロガー
type LevelLogger struct {
	logger   *log.Logger
	minLevel LogLevel
}

// NewLevelLogger はLevelLoggerのコンストラクタ
func NewLevelLogger() *LevelLogger {
	minLevel := getLogLevelFromEnv()
	return &LevelLogger{
		logger:   log.New(os.Stdout, "", log.LstdFlags|log.Lshortfile),
		minLevel: minLevel,
	}
}

// getLogLevelFromEnv は環境変数からログレベルを取得
func getLogLevelFromEnv() LogLevel {
	levelStr := strings.ToUpper(os.Getenv("LOG_LEVEL"))
	
	switch levelStr {
	case "DEBUG":
		return DEBUG
	case "INFO":
		return INFO
	case "WARN":
		return WARN
	case "ERROR":
		return ERROR
	default:
		return INFO // デフォルトはINFO
	}
}

// shouldLog は指定されたレベルのログを出力すべきかを判定
func (l *LevelLogger) shouldLog(level LogLevel) bool {
	return level >= l.minLevel
}

// Debug はデバッグレベルのログを出力
func (l *LevelLogger) Debug(message string, fields ...interface{}) {
	if l.shouldLog(DEBUG) {
		l.logWithLevel(DEBUG, message, fields...)
	}
}

// Info は情報レベルのログを出力
func (l *LevelLogger) Info(message string, fields ...interface{}) {
	if l.shouldLog(INFO) {
		l.logWithLevel(INFO, message, fields...)
	}
}

// Warn は警告レベルのログを出力
func (l *LevelLogger) Warn(message string, fields ...interface{}) {
	if l.shouldLog(WARN) {
		l.logWithLevel(WARN, message, fields...)
	}
}

// Error はエラーレベルのログを出力
func (l *LevelLogger) Error(message string, fields ...interface{}) {
	if l.shouldLog(ERROR) {
		l.logWithLevel(ERROR, message, fields...)
	}
}

// logWithLevel はレベル付きログを出力
func (l *LevelLogger) logWithLevel(level LogLevel, message string, fields ...interface{}) {
	levelName := logLevelNames[level]
	logMessage := fmt.Sprintf("[%s] %s", levelName, message)
	
	// フィールドを追加（key-value ペア）
	for i := 0; i < len(fields); i += 2 {
		if i+1 < len(fields) {
			logMessage += fmt.Sprintf(" %s=%v", fields[i], fields[i+1])
		}
	}
	
	l.logger.Println(logMessage)
}

// SetLevel は最小ログレベルを動的に変更
func (l *LevelLogger) SetLevel(level LogLevel) {
	l.minLevel = level
}

// GetLevel は現在の最小ログレベルを取得
func (l *LevelLogger) GetLevel() LogLevel {
	return l.minLevel
}

// IsDebugEnabled はデバッグレベルが有効かを判定
func (l *LevelLogger) IsDebugEnabled() bool {
	return l.shouldLog(DEBUG)
}