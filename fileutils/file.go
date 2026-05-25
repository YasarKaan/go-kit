package fileutils

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
)

// MultipartFile represents the memory representation of a file,
// mimicking Spring's MultipartFile in Go.
type MultipartFile struct {
	Name        string
	Filename    string
	ContentType string
	Size        int64
	Content     []byte
}

func (m *MultipartFile) GetOriginalFilename() string {
	return m.Filename
}

func (m *MultipartFile) GetContentType() string {
	return m.ContentType
}

func (m *MultipartFile) IsEmpty() bool {
	return len(m.Content) == 0
}

func (m *MultipartFile) GetSize() int64 {
	return m.Size
}

func (m *MultipartFile) GetBytes() []byte {
	return m.Content
}

func (m *MultipartFile) GetReader() io.Reader {
	return bytes.NewReader(m.Content)
}

// FromFile creates a MultipartFile from a local file path.
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
