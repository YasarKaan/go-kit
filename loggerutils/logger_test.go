package loggerutils

import (
	"bytes"
	"context"
	"encoding/json"
	"log"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/YasarKaan/go-kit/enums"
)

func TestSanitize(t *testing.T) {
	inputJson := `{"username":"kaan","password":"mysecretpassword","token":"eyJ..."}`
	expectedJson := `{"username":"kaan","password":"***","token":"***"}`
	sanitizedJson := Sanitize(inputJson)
	if sanitizedJson != expectedJson {
		t.Errorf("expected masked json: %s, got: %s", expectedJson, sanitizedJson)
	}

	inputKv := "user=kaan password=secret token=abc123"
	expectedKv := "user=kaan password=*** token=***"
	sanitizedKv := Sanitize(inputKv)
	if sanitizedKv != expectedKv {
		t.Errorf("expected masked key-value: %s, got: %s", expectedKv, sanitizedKv)
	}

	// CRLF injection test
	inputCrlf := "hello\r\nworld\nlog injection"
	expectedCrlf := "hello world log injection"
	sanitizedCrlf := Sanitize(inputCrlf)
	if sanitizedCrlf != expectedCrlf {
		t.Errorf("expected CR/LF removed: %s, got: %s", expectedCrlf, sanitizedCrlf)
	}
}

func TestInitLogger(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "go-kit-log-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	logFilePath := filepath.Join(tempDir, "app.log")
	
	// Test initialization
	InitLogger(logFilePath, 10, 3, 7, false, false)

	Info("Test log entry to verify writing to file")

	// Verify file exists
	if _, err := os.Stat(logFilePath); os.IsNotExist(err) {
		t.Error("expected log file to be created, but it does not exist")
	}
}

func TestLogLevelFilteringAndJSON(t *testing.T) {
	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer log.SetOutput(os.Stderr) // Reset

	// Configure level to Warn
	SetLogLevel(enums.LevelWarn)

	// info should be filtered out
	Info("This info message should not appear")

	// warn should be printed
	Warn("This warn message should appear")

	output := buf.String()
	if strings.Contains(output, "This info message should not appear") {
		t.Error("Info message printed despite minimum level set to WARN")
	}

	if !strings.Contains(output, "This warn message should appear") {
		t.Fatal("Warn message not printed when level was set to WARN")
	}

	// Verify it's valid JSON structured format
	var entry map[string]any
	err := json.Unmarshal([]byte(strings.TrimSpace(output)), &entry)
	if err != nil {
		t.Fatalf("logged output is not valid JSON: %v. Output: %s", err, output)
	}

	if entry["level"] != string(enums.LevelWarn) {
		t.Errorf("expected log level WARN, got: %s", entry["level"])
	}

	if entry["message"] != "This warn message should appear" {
		t.Errorf("expected log message match, got: %s", entry["message"])
	}

	if entry["time"] == nil || entry["time"] == "" {
		t.Error("expected non-empty timestamp field")
	}
}

func TestStructuredLoggingFields(t *testing.T) {
	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer log.SetOutput(os.Stderr) // Reset

	// Reset log level
	SetLogLevel(enums.LevelDebug)

	// Create context with fields
	ctx := context.Background()
	ctx = WithFields(ctx, map[string]any{
		"requestId": "req-12345",
		"traceId":   "trace-9999",
		"userId":    "user-kaan",
	})

	InfoContext(ctx, "Accessing dashboard")

	output := buf.String()

	var data map[string]any
	err := json.Unmarshal([]byte(strings.TrimSpace(output)), &data)
	if err != nil {
		t.Fatalf("logged output is not valid JSON: %v", err)
	}

	if data["message"] != "Accessing dashboard" {
		t.Errorf("expected message 'Accessing dashboard', got: %v", data["message"])
	}

	if data["requestId"] != "req-12345" {
		t.Errorf("expected requestId 'req-12345', got: %v", data["requestId"])
	}

	if data["traceId"] != "trace-9999" {
		t.Errorf("expected traceId 'trace-9999', got: %v", data["traceId"])
	}

	if data["userId"] != "user-kaan" {
		t.Errorf("expected userId 'user-kaan', got: %v", data["userId"])
	}

	if data["level"] != "INFO" {
		t.Errorf("expected level 'INFO', got: %v", data["level"])
	}
}
