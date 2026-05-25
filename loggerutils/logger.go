package loggerutils

import (
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/YasarKaan/go-kit/enums"
)

var (
	sensitiveJsonPattern = regexp.MustCompile(`(?i)"([^"]*(?:password|token|secret|credential|apikey|api_key|authorization|private_key|otp|pin|cvv)[^"]*)"\s*:\s*(?:"[^"]*"|[^,\}\]\s]+)`)
	sensitiveKvPattern   = regexp.MustCompile(`(?i)(password|token|secret|credential|apikey|api_key|authorization|private_key|otp|pin|cvv)\s*[=:]\s*(?:"[^"]*"|'[^']*'|[^\s,;&]+)`)
	crlfPattern          = regexp.MustCompile(`[\r\n]+`)
)

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

	return result
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
	log.Println("[DEBUG] " + formatMessage(message, args))
}

// Info logs an info level message.
func Info(message string, args ...any) {
	log.Println("[INFO] " + formatMessage(message, args))
}

// Warn logs a warning level message.
func Warn(message string, args ...any) {
	log.Println("[WARN] " + formatMessage(message, args))
}

// Error logs an error level message.
func Error(message string, args ...any) {
	log.Println("[ERROR] " + formatMessage(message, args))
}

// ErrorWithThrowable logs an error with a details error struct.
func ErrorWithThrowable(message string, err error) {
	if err != nil {
		log.Printf("[ERROR] %s: %v\n", Sanitize(message), err)
	} else {
		log.Println("[ERROR] " + Sanitize(message))
	}
}

// PushLog is a backward-compatible method matching the Java signature.
func PushLog(level enums.LogLevel, message string, args ...any) {
	formatted := formatMessage(message, args)
	switch level {
	case enums.LevelDebug:
		log.Println("[DEBUG] " + formatted)
	case enums.LevelInfo:
		log.Println("[INFO] " + formatted)
	case enums.LevelWarn:
		log.Println("[WARN] " + formatted)
	case enums.LevelError:
		log.Println("[ERROR] " + formatted)
	case enums.LevelFatal:
		log.Println("[FATAL] " + formatted)
	default:
		log.Printf("[WARN] Unknown log level %s. Msg: %s\n", level, formatted)
	}
}
