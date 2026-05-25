package fileutils

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
)

// MultipartFile represents the memory representation or streaming source of a file,
// mimicking Spring's MultipartFile in Go.
type MultipartFile struct {
	Name        string
	Filename    string
	ContentType string
	Size        int64
	Content     []byte
	Reader      io.Reader // Optional streaming source to avoid loading entire file into memory
}

func (m *MultipartFile) GetOriginalFilename() string {
	return m.Filename
}

func (m *MultipartFile) GetContentType() string {
	return m.ContentType
}

func (m *MultipartFile) IsEmpty() bool {
	return m.Size == 0
}

func (m *MultipartFile) GetSize() int64 {
	return m.Size
}

func (m *MultipartFile) GetBytes() []byte {
	if len(m.Content) == 0 && m.Reader != nil {
		// Fallback: if it's a stream, read it once (note: this consumes the stream)
		if closer, ok := m.Reader.(io.Closer); ok {
			defer closer.Close()
		}
		bytes, err := io.ReadAll(m.Reader)
		if err == nil {
			m.Content = bytes
		}
	}
	return m.Content
}

func (m *MultipartFile) GetReader() io.Reader {
	if m.Reader != nil {
		return m.Reader
	}
	return bytes.NewReader(m.Content)
}

// SaveToFile saves the multipart file content to the target destination path using streaming io.Copy.
// This is memory-safe (OOM-free) even for large files.
func (m *MultipartFile) SaveToFile(dstPath string) error {
	if err := os.MkdirAll(filepath.Dir(dstPath), 0755); err != nil {
		return err
	}

	dstFile, err := os.Create(dstPath)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	srcReader := m.GetReader()
	if closer, ok := srcReader.(io.Closer); ok {
		defer closer.Close()
	}

	_, err = io.Copy(dstFile, srcReader)
	return err
}

// FromFile creates a memory-loaded MultipartFile from a local file path.
// WARNING: This loads the entire file content into memory. For large files,
// use FromFileStream instead to avoid Out-Of-Memory (OOM) exceptions.
func FromFile(filePath string, contentType string) (*MultipartFile, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		return nil, err
	}

	content, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}

	return &MultipartFile{
		Name:        stat.Name(),
		Filename:    filepath.Base(filePath),
		ContentType: contentType,
		Size:        stat.Size(),
		Content:     content,
	}, nil
}

// FromFileStream creates a streaming (memory-safe) MultipartFile from a local file path.
// Callers should note that the underlying file remains open for reading via GetReader().
func FromFileStream(filePath string, contentType string) (*MultipartFile, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}

	stat, err := file.Stat()
	if err != nil {
		file.Close()
		return nil, err
	}

	return &MultipartFile{
		Name:        stat.Name(),
		Filename:    filepath.Base(filePath),
		ContentType: contentType,
		Size:        stat.Size(),
		Reader:      file,
	}, nil
}
