package loggerutils

import (
	"os"
	"path/filepath"
	"testing"
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
