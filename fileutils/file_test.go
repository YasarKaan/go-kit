package fileutils

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestMultipartFileMemory(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "go-kit-file-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	filePath := filepath.Join(tempDir, "test.txt")
	content := []byte("hello world multipart memory")
	if err := os.WriteFile(filePath, content, 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	mf, err := FromFile(filePath, "text/plain")
	if err != nil {
		t.Fatalf("failed to load file: %v", err)
	}

	if mf.IsEmpty() {
		t.Error("expected file to not be empty")
	}

	if mf.GetSize() != int64(len(content)) {
		t.Errorf("expected size %d, got: %d", len(content), mf.GetSize())
	}

	if !bytes.Equal(mf.GetBytes(), content) {
		t.Errorf("expected content %q, got: %q", string(content), string(mf.GetBytes()))
	}
}

func TestMultipartFileStreamAndSave(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "go-kit-file-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	srcPath := filepath.Join(tempDir, "src.txt")
	content := []byte("hello world multipart stream copying is awesome")
	if err := os.WriteFile(srcPath, content, 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	mf, err := FromFileStream(srcPath, "text/plain")
	if err != nil {
		t.Fatalf("failed to create stream file: %v", err)
	}

	if mf.IsEmpty() {
		t.Error("expected stream file to not be empty")
	}

	dstPath := filepath.Join(tempDir, "dst.txt")
	if err := mf.SaveToFile(dstPath); err != nil {
		t.Fatalf("failed to save file: %v", err)
	}

	savedContent, err := os.ReadFile(dstPath)
	if err != nil {
		t.Fatalf("failed to read saved file: %v", err)
	}

	if !bytes.Equal(savedContent, content) {
		t.Errorf("saved content %q does not match original %q", string(savedContent), string(content))
	}
}
