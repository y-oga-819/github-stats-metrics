package logger

import (
	"fmt"
	"log"
	"os"
)

// StandardLogger は標準的なログ出力を提供
type StandardLogger struct {
	logger *log.Logger
}

// NewStandardLogger はStandardLoggerのコンストラクタ
func NewStandardLogger() *StandardLogger {
	return &StandardLogger{
		logger: log.New(os.Stdout, "", log.LstdFlags|log.Lshortfile),
	}
}

// Error はエラーレベルのログを出力
func (l *StandardLogger) Error(message string, fields ...interface{}) {
	l.logWithFields("ERROR", message, fields...)
}

// Warn は警告レベルのログを出力
func (l *StandardLogger) Warn(message string, fields ...interface{}) {
	l.logWithFields("WARN", message, fields...)
}

// Info は情報レベルのログを出力
func (l *StandardLogger) Info(message string, fields ...interface{}) {
	l.logWithFields("INFO", message, fields...)
}

// logWithFields はフィールド付きログを出力
func (l *StandardLogger) logWithFields(level, message string, fields ...interface{}) {
	logMessage := level + ": " + message
	
	// フィールドを追加（key-value ペア）
	for i := 0; i < len(fields); i += 2 {
		if i+1 < len(fields) {
			logMessage += " " + fields[i].(string) + "=" + formatValue(fields[i+1])
		}
	}
	
	l.logger.Println(logMessage)
}

// formatValue は値を文字列形式に変換
func formatValue(v interface{}) string {
	switch val := v.(type) {
	case string:
		return val
	case int:
		return fmt.Sprintf("%d", val)
	case int64:
		return fmt.Sprintf("%d", val)
	case float64:
		return fmt.Sprintf("%.2f", val)
	default:
		return fmt.Sprintf("%v", val)
	}
}