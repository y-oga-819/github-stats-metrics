package logging

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"runtime"
	"strings"
	"time"
)

// LogLevel はログレベルを表す
type LogLevel int

const (
	DEBUG LogLevel = iota
	INFO
	WARN
	ERROR
	FATAL
)

// String はLogLevelの文字列表現を返す
func (l LogLevel) String() string {
	switch l {
	case DEBUG:
		return "DEBUG"
	case INFO:
		return "INFO"
	case WARN:
		return "WARN"
	case ERROR:
		return "ERROR"
	case FATAL:
		return "FATAL"
	default:
		return "UNKNOWN"
	}
}

// LogEntry は構造化ログエントリ
type LogEntry struct {
	Timestamp string                 `json:"timestamp"`
	Level     string                 `json:"level"`
	Message   string                 `json:"message"`
	Service   string                 `json:"service"`
	Version   string                 `json:"version"`
	TraceID   string                 `json:"trace_id,omitempty"`
	Fields    map[string]interface{} `json:"fields,omitempty"`
	Caller    string                 `json:"caller,omitempty"`
	Error     string                 `json:"error,omitempty"`
}

// StructuredLogger は構造化ログを出力するロガー
type StructuredLogger struct {
	output      io.Writer
	level       LogLevel
	service     string
	version     string
	enableCaller bool
}

// NewStructuredLogger は新しい構造化ロガーを作成
func NewStructuredLogger(level LogLevel, service, version string) *StructuredLogger {
	return &StructuredLogger{
		output:       os.Stdout,
		level:        level,
		service:      service,
		version:      version,
		enableCaller: true,
	}
}

// SetOutput は出力先を設定
func (l *StructuredLogger) SetOutput(w io.Writer) {
	l.output = w
}

// SetLevel はログレベルを設定
func (l *StructuredLogger) SetLevel(level LogLevel) {
	l.level = level
}

// EnableCaller は呼び出し元情報の記録を有効化
func (l *StructuredLogger) EnableCaller(enable bool) {
	l.enableCaller = enable
}

// log は基底のログメソッド
func (l *StructuredLogger) log(ctx context.Context, level LogLevel, message string, fields map[string]interface{}, err error) {
	if level < l.level {
		return
	}

	entry := LogEntry{
		Timestamp: time.Now().UTC().Format(time.RFC3339Nano),
		Level:     level.String(),
		Message:   message,
		Service:   l.service,
		Version:   l.version,
		Fields:    fields,
	}

	// トレースIDを取得（存在する場合）
	if traceID := getTraceID(ctx); traceID != "" {
		entry.TraceID = traceID
	}

	// エラー情報を追加
	if err != nil {
		entry.Error = err.Error()
	}

	// 呼び出し元情報を追加
	if l.enableCaller {
		if caller := getCaller(); caller != "" {
			entry.Caller = caller
		}
	}

	// JSONエンコードして出力
	data, _ := json.Marshal(entry)
	fmt.Fprintf(l.output, "%s\n", data)
}

// Debug はデバッグレベルのログを出力
func (l *StructuredLogger) Debug(ctx context.Context, message string, fields ...map[string]interface{}) {
	l.log(ctx, DEBUG, message, mergeFields(fields...), nil)
}

// Info は情報レベルのログを出力
func (l *StructuredLogger) Info(ctx context.Context, message string, fields ...map[string]interface{}) {
	l.log(ctx, INFO, message, mergeFields(fields...), nil)
}

// Warn は警告レベルのログを出力
func (l *StructuredLogger) Warn(ctx context.Context, message string, fields ...map[string]interface{}) {
	l.log(ctx, WARN, message, mergeFields(fields...), nil)
}

// Error はエラーレベルのログを出力
func (l *StructuredLogger) Error(ctx context.Context, message string, err error, fields ...map[string]interface{}) {
	l.log(ctx, ERROR, message, mergeFields(fields...), err)
}

// Fatal は致命的エラーレベルのログを出力してプログラムを終了
func (l *StructuredLogger) Fatal(ctx context.Context, message string, err error, fields ...map[string]interface{}) {
	l.log(ctx, FATAL, message, mergeFields(fields...), err)
	os.Exit(1)
}

// WithFields はフィールド付きのロガーコンテキストを作成
func (l *StructuredLogger) WithFields(fields map[string]interface{}) *LoggerContext {
	return &LoggerContext{
		logger: l,
		fields: fields,
	}
}

// LoggerContext はフィールド付きのロガーコンテキスト
type LoggerContext struct {
	logger *StructuredLogger
	fields map[string]interface{}
}

// Debug はデバッグレベルのログを出力（フィールド付き）
func (lc *LoggerContext) Debug(ctx context.Context, message string, additionalFields ...map[string]interface{}) {
	fields := mergeFields(lc.fields)
	for _, f := range additionalFields {
		for k, v := range f {
			fields[k] = v
		}
	}
	lc.logger.log(ctx, DEBUG, message, fields, nil)
}

// Info は情報レベルのログを出力（フィールド付き）
func (lc *LoggerContext) Info(ctx context.Context, message string, additionalFields ...map[string]interface{}) {
	fields := mergeFields(lc.fields)
	for _, f := range additionalFields {
		for k, v := range f {
			fields[k] = v
		}
	}
	lc.logger.log(ctx, INFO, message, fields, nil)
}

// Warn は警告レベルのログを出力（フィールド付き）
func (lc *LoggerContext) Warn(ctx context.Context, message string, additionalFields ...map[string]interface{}) {
	fields := mergeFields(lc.fields)
	for _, f := range additionalFields {
		for k, v := range f {
			fields[k] = v
		}
	}
	lc.logger.log(ctx, WARN, message, fields, nil)
}

// Error はエラーレベルのログを出力（フィールド付き）
func (lc *LoggerContext) Error(ctx context.Context, message string, err error, additionalFields ...map[string]interface{}) {
	fields := mergeFields(lc.fields)
	for _, f := range additionalFields {
		for k, v := range f {
			fields[k] = v
		}
	}
	lc.logger.log(ctx, ERROR, message, fields, err)
}

// ヘルパー関数

// mergeFields は複数のフィールドマップをマージ
func mergeFields(fieldMaps ...map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	for _, fields := range fieldMaps {
		for k, v := range fields {
			result[k] = v
		}
	}
	return result
}

// getCaller は呼び出し元の情報を取得
func getCaller() string {
	// スタックを3つ上に遡る（getCaller -> log -> 実際の呼び出し）
	_, file, line, ok := runtime.Caller(3)
	if !ok {
		return ""
	}

	// ファイルパスから相対パスを取得
	parts := strings.Split(file, "/")
	if len(parts) > 2 {
		file = strings.Join(parts[len(parts)-2:], "/")
	}

	return fmt.Sprintf("%s:%d", file, line)
}

// getTraceID はコンテキストからトレースIDを取得
func getTraceID(ctx context.Context) string {
	if ctx == nil {
		return ""
	}

	// 実際のトレーシングライブラリ（OpenTelemetry等）に応じて実装
	// ここでは簡単な実装例
	if traceID := ctx.Value("trace_id"); traceID != nil {
		if id, ok := traceID.(string); ok {
			return id
		}
	}

	return ""
}

// ParseLogLevel は文字列からLogLevelを解析
func ParseLogLevel(level string) (LogLevel, error) {
	switch strings.ToUpper(level) {
	case "DEBUG":
		return DEBUG, nil
	case "INFO":
		return INFO, nil
	case "WARN", "WARNING":
		return WARN, nil
	case "ERROR":
		return ERROR, nil
	case "FATAL":
		return FATAL, nil
	default:
		return INFO, fmt.Errorf("unknown log level: %s", level)
	}
}