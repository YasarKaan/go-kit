package exceptions

import (
	"fmt"
	"strings"

	"github.com/YasarKaan/go-kit/loggerutils"
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

// NewCustomWebServerException creates a new CustomWebServerException with sanitized metadata.
// Strips Authorization, Cookie, and Set-Cookie headers, and masks sensitive values in body payloads.
func NewCustomWebServerException(code int, msg string, metadata map[string]any) *CustomWebServerException {
	if metadata == nil {
		return &CustomWebServerException{
			ErrorCode: code,
			Message:   msg,
		}
	}

	sanitizedMetadata := make(map[string]any)
	for k, v := range metadata {
		lowerK := strings.ToLower(k)

		// Sanitize headers map
		if lowerK == "headers" {
			if headersMap, ok := v.(map[string]string); ok {
				cleanHeaders := make(map[string]string)
				for hk, hv := range headersMap {
					lowerHk := strings.ToLower(hk)
					if lowerHk == "authorization" || lowerHk == "cookie" || lowerHk == "set-cookie" {
						cleanHeaders[hk] = "***"
					} else {
						cleanHeaders[hk] = hv
					}
				}
				sanitizedMetadata[k] = cleanHeaders
			} else if headersInterfaceMap, ok := v.(map[string]any); ok {
				cleanHeaders := make(map[string]any)
				for hk, hv := range headersInterfaceMap {
					lowerHk := strings.ToLower(hk)
					if lowerHk == "authorization" || lowerHk == "cookie" || lowerHk == "set-cookie" {
						cleanHeaders[hk] = "***"
					} else {
						cleanHeaders[hk] = hv
					}
				}
				sanitizedMetadata[k] = cleanHeaders
			} else {
				sanitizedMetadata[k] = v
			}
			continue
		}

		// Sanitize body/payload strings
		if lowerK == "body" || lowerK == "payload" || lowerK == "response" {
			if strVal, ok := v.(string); ok {
				sanitizedMetadata[k] = loggerutils.Sanitize(strVal)
			} else {
				sanitizedMetadata[k] = v
			}
			continue
		}

		sanitizedMetadata[k] = v
	}

	return &CustomWebServerException{
		ErrorCode: code,
		Message:   msg,
		Metadata:  sanitizedMetadata,
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
