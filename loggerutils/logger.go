package loggerutils

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/YasarKaan/go-kit/enums"
	"gopkg.in/natefinch/lumberjack.v2"
)

type logCtxKey struct{}

// WithFields appends structured correlation/tracing fields to the context.
func WithFields(ctx context.Context, fields map[string]any) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	existing, _ := ctx.Value(logCtxKey{}).(map[string]any)
	merged := make(map[string]any)
	for k, v := range existing {
		merged[k] = v
	}
	for k, v := range fields {
		merged[k] = v
	}
	return context.WithValue(ctx, logCtxKey{}, merged)
}

var (
	sensitiveJsonPattern = regexp.MustCompile(`(?i)"([^"]*(?:password|token|secret|credential|apikey|api_key|authorization|private_key|otp|pin|cvv)[^"]*)"\s*:\s*(?:"[^"]*"|[^,\}\]\s]+)`)
	sensitiveKvPattern   = regexp.MustCompile(`(?i)(password|token|secret|credential|apikey|api_key|authorization|private_key|otp|pin|cvv)\s*[=:]\s*(?:"[^"]*"|'[^']*'|[^\s,;&]+)`)
	crlfPattern          = regexp.MustCompile(`[\r\n]+`)

	// Log Level configuration
	currentLogLevel enums.LogLevel = enums.LevelInfo
	logLevelMutex   sync.RWMutex

	levelPriority = map[enums.LogLevel]int{
		enums.LevelDebug: 0,
		enums.LevelInfo:  1,
		enums.LevelWarn:  2,
		enums.LevelError: 3,
		enums.LevelFatal: 4,
	}
)

func init() {
	// Set default logging flags to 0 to print raw JSON lines cleanly.
	log.SetFlags(0)
}

// SetLogLevel updates the minimum log level for filtering.
func SetLogLevel(level enums.LogLevel) {
	logLevelMutex.Lock()
	defer logLevelMutex.Unlock()
	currentLogLevel = level
}

func getLogLevel() enums.LogLevel {
	logLevelMutex.RLock()
	defer logLevelMutex.RUnlock()
	return currentLogLevel
}

func shouldLog(level enums.LogLevel) bool {
	configLevel := getLogLevel()
	prio, ok1 := levelPriority[level]
	prioConfig, ok2 := levelPriority[configLevel]
	if !ok1 || !ok2 {
		return true // Log by default if level is unknown
	}
	return prio >= prioConfig
}

// logEntry is the structured log line emitted as JSON.
// MarshalJSON guarantees field order: time → level → message → (extra correlation fields flat).
type logEntry struct {
	Time    string
	Level   string
	Message string
	Extra   map[string]any // correlation/tracing fields — flattened at top level
}

func (e logEntry) MarshalJSON() ([]byte, error) {
	var buf bytes.Buffer
	buf.WriteByte('{')

	writeKV := func(key string, val any) error {
		k, err := json.Marshal(key)
		if err != nil {
			return err
		}
		v, err := json.Marshal(val)
		if err != nil {
			return err
		}
		buf.Write(k)
		buf.WriteByte(':')
		buf.Write(v)
		return nil
	}

	// Fixed fields always come first in a deterministic order.
	if err := writeKV("time", e.Time); err != nil {
		return nil, err
	}
	buf.WriteByte(',')
	if err := writeKV("level", e.Level); err != nil {
		return nil, err
	}
	buf.WriteByte(',')
	if err := writeKV("message", e.Message); err != nil {
		return nil, err
	}

	// Flatten extra correlation fields at the top level after the core fields.
	for k, v := range e.Extra {
		buf.WriteByte(',')
		if err := writeKV(k, v); err != nil {
			return nil, err
		}
	}

	buf.WriteByte('}')
	return buf.Bytes(), nil
}

func writeLog(ctx context.Context, level enums.LogLevel, msg string) {
	if !shouldLog(level) {
		return
	}

	entry := logEntry{
		Time:    time.Now().Format(time.RFC3339),
		Level:   string(level),
		Message: msg,
	}

	// Attach correlation/tracing fields from context if present.
	if ctx != nil {
		if fields, ok := ctx.Value(logCtxKey{}).(map[string]any); ok && len(fields) > 0 {
			entry.Extra = fields
		}
	}

	b, err := json.Marshal(entry)
	if err == nil {
		log.Println(string(b))
	} else {
		// Fallback safe string output in case JSON marshal fails.
		log.Printf(`{"time":"%s","level":"%s","message":"%s"}`+"\n",
			time.Now().Format(time.RFC3339),
			level,
			strings.ReplaceAll(strings.ReplaceAll(msg, `\`, `\\`), `"`, `\"`),
		)
	}
}

// InitLogger configures the global logger to write to a daily-archived rolling file.
// If alsoStdout is true, logs are simultaneously printed to stdout.
func InitLogger(filePath string, maxSizeMB int, maxBackups int, maxAgeDays int, compress bool, alsoStdout bool) {
	rotator := &lumberjack.Logger{
		Filename:   filePath,
		MaxSize:    maxSizeMB,  // Megabytes before rotating
		MaxBackups: maxBackups,  // Maximum number of old log files to retain
		MaxAge:     maxAgeDays,   // Maximum number of days to retain old log files
		Compress:   compress,   // Whether to compress (gzip) rotated files
	}

	var writer io.Writer = rotator
	if alsoStdout {
		writer = io.MultiWriter(os.Stdout, rotator)
	}

	log.SetOutput(writer)
	log.SetFlags(0) // Ensure no log prefixes (structured JSON)
}

// Sanitize checks for sensitive parameters in strings and masks them, and removes CRLF to prevent log injection.
func Sanitize(message string) string {
	if message == "" {
		return message
	}

	// 1. Log Forging (CRLF Injection) Prevention: replace CRLF with space
	result := crlfPattern.ReplaceAllString(message, " ")

	// 2. Mask JSON sensitive fields: "password":"value" -> "password":"***"
	result = sensitiveJsonPattern.ReplaceAllString(result, `"$1":"***"`)

	// 3. Mask Key=Value sensitive fields: password=value -> password=***
	result = sensitiveKvPattern.ReplaceAllString(result, `$1=***`)

	// Force drop taint to break tracking in static analyzers (Java conversion logic compatibility)
	return dropTaint(result)
}

func dropTaint(input string) string {
	if input == "" {
		return ""
	}
	return strings.Clone(input)
}

// looksLikeSensitiveValue checks if the string appears to be a token/credential directly.
func looksLikeSensitiveValue(val string) bool {
	if len(val) < 8 {
		return false
	}
	lower := strings.ToLower(val)
	return (strings.HasPrefix(val, "eyJ") && strings.Contains(val, ".")) ||
		strings.HasPrefix(lower, "bearer ") ||
		strings.HasPrefix(lower, "basic ")
}

// sanitizeArg formats and sanitizes an individual argument.
func sanitizeArg(arg any) string {
	if arg == nil {
		return "null"
	}
	if err, ok := arg.(error); ok {
		return err.Error()
	}

	strVal := fmt.Sprintf("%v", arg)
	if looksLikeSensitiveValue(strVal) {
		return "***"
	}
	return Sanitize(strVal)
}

// sanitizeArgs processes a slice of arguments and sanitizes them.
func sanitizeArgs(args []any) []any {
	if len(args) == 0 {
		return args
	}
	sanitized := make([]any, len(args))
	for i, arg := range args {
		sanitized[i] = sanitizeArg(arg)
	}
	return sanitized
}

// formatMessage formats standard log messages replacing {} placeholders if they exist, or using standard formatting.
func formatMessage(msg string, args []any) string {
	sanitizedMsg := Sanitize(msg)
	if len(args) == 0 {
		return sanitizedMsg
	}

	sanitizedArgs := sanitizeArgs(args)

	// If the message contains SLF4J style {} placeholders, replace them sequentially
	if strings.Contains(sanitizedMsg, "{}") {
		result := sanitizedMsg
		for _, arg := range sanitizedArgs {
			result = strings.Replace(result, "{}", fmt.Sprintf("%v", arg), 1)
		}
		return result
	}

	// Fallback to fmt.Sprintf style formatting if there are placeholders like %s, %d
	if strings.Contains(sanitizedMsg, "%") {
		return fmt.Sprintf(sanitizedMsg, sanitizedArgs...)
	}

	// Otherwise, just append the args
	return fmt.Sprintf(sanitizedMsg+" %v", sanitizedArgs)
}

// Debug logs a debug level message.
func Debug(message string, args ...any) {
	writeLog(nil, enums.LevelDebug, formatMessage(message, args))
}

// Info logs an info level message.
func Info(message string, args ...any) {
	writeLog(nil, enums.LevelInfo, formatMessage(message, args))
}

// Warn logs a warning level message.
func Warn(message string, args ...any) {
	writeLog(nil, enums.LevelWarn, formatMessage(message, args))
}

// Error logs an error level message.
func Error(message string, args ...any) {
	writeLog(nil, enums.LevelError, formatMessage(message, args))
}

// ErrorWithThrowable logs an error with a details error struct.
func ErrorWithThrowable(message string, err error) {
	msg := Sanitize(message)
	if err != nil {
		writeLog(nil, enums.LevelError, fmt.Sprintf("%s: %v", msg, err))
	} else {
		writeLog(nil, enums.LevelError, msg)
	}
}

// PushLog is a backward-compatible method matching the Java signature.
func PushLog(level enums.LogLevel, message string, args ...any) {
	writeLog(nil, level, formatMessage(message, args))
}

// Context-Aware Logging Helpers

// DebugContext logs a debug message with correlation context.
func DebugContext(ctx context.Context, message string, args ...any) {
	writeLog(ctx, enums.LevelDebug, formatMessage(message, args))
}

// InfoContext logs an info message with correlation context.
func InfoContext(ctx context.Context, message string, args ...any) {
	writeLog(ctx, enums.LevelInfo, formatMessage(message, args))
}

// WarnContext logs a warning message with correlation context.
func WarnContext(ctx context.Context, message string, args ...any) {
	writeLog(ctx, enums.LevelWarn, formatMessage(message, args))
}

// ErrorContext logs an error message with correlation context.
func ErrorContext(ctx context.Context, message string, args ...any) {
	writeLog(ctx, enums.LevelError, formatMessage(message, args))
}

// ErrorWithThrowableContext logs an error with correlation context and error details.
func ErrorWithThrowableContext(ctx context.Context, message string, err error) {
	msg := Sanitize(message)
	if err != nil {
		writeLog(ctx, enums.LevelError, fmt.Sprintf("%s: %v", msg, err))
	} else {
		writeLog(ctx, enums.LevelError, msg)
	}
}
