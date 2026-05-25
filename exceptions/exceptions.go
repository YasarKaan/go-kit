package exceptions

import (
	"fmt"
)

// CustomWebServerException represents an HTTP-friendly web server exception.
type CustomWebServerException struct {
	ErrorCode int
	Message   string
	Metadata  map[string]any
}

func (e *CustomWebServerException) Error() string {
	if len(e.Metadata) > 0 {
		return fmt.Sprintf("Web Server Error (Code: %d): %s, Metadata: %v", e.ErrorCode, e.Message, e.Metadata)
	}
	return fmt.Sprintf("Web Server Error (Code: %d): %s", e.ErrorCode, e.Message)
}

func NewCustomWebServerException(code int, msg string, metadata map[string]any) *CustomWebServerException {
	return &CustomWebServerException{
		ErrorCode: code,
		Message:   msg,
		Metadata:  metadata,
	}
}

// ResourceNotFoundException represents a resource not found error.
type ResourceNotFoundException struct {
	Message string
}

func (e *ResourceNotFoundException) Error() string {
	return e.Message
}

func NewResourceNotFoundException(msg string) *ResourceNotFoundException {
	return &ResourceNotFoundException{
		Message: msg,
	}
}

// HttpRequestException represents an error during an HTTP request.
type HttpRequestException struct {
	Message string
}

func (e *HttpRequestException) Error() string {
	return e.Message
}

func NewHttpRequestException(msg string) *HttpRequestException {
	return &HttpRequestException{
		Message: msg,
	}
}
