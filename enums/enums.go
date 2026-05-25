package enums

import (
	"fmt"
	"strings"
)

type Priority string

const (
	PriorityHigh   Priority = "High"
	PriorityMedium Priority = "Medium"
	PriorityNormal Priority = "Normal"
	PriorityLow    Priority = "Low"
)

type Severity string

const (
	SeverityCritical Severity = "CRITICAL"
	SeverityMajor    Severity = "MAJOR"
	SeverityMinor    Severity = "MINOR"
	SeverityWarning  Severity = "WARNING"
	SeverityNormal   Severity = "NORMAL"
)

type HttpMethod string

const (
	MethodGet    HttpMethod = "GET"
	MethodPost   HttpMethod = "POST"
	MethodPut    HttpMethod = "PUT"
	MethodPatch  HttpMethod = "PATCH"
	MethodDelete HttpMethod = "DELETE"
)

type LogLevel string

const (
	LevelDebug LogLevel = "DEBUG"
	LevelInfo  LogLevel = "INFO"
	LevelWarn  LogLevel = "WARN"
	LevelError LogLevel = "ERROR"
	LevelFatal LogLevel = "FATAL"
)

type OrderByDirection string

const (
	OrderAsc  OrderByDirection = "ASC"
	OrderDesc OrderByDirection = "DESC"
)

type ContentType string

const (
	ContentTypeJSON           ContentType = "application/json"
	ContentTypeFormUrlEncoded ContentType = "application/x-www-form-urlencoded"
	ContentTypeTextPlain      ContentType = "text/plain"
)

func (c ContentType) Value() string {
	return string(c)
}

func ContentTypeFromString(contentType string) (ContentType, error) {
	if contentType == "" {
		return "", fmt.Errorf("content type is empty")
	}

	normalized := strings.TrimSpace(strings.Split(strings.ToLower(contentType), ";")[0])

	switch normalized {
	case "application/json":
		return ContentTypeJSON, nil
	case "application/x-www-form-urlencoded":
		return ContentTypeFormUrlEncoded, nil
	case "text/plain":
		return ContentTypeTextPlain, nil
	default:
		return "", fmt.Errorf("unsupported content type: %s", contentType)
	}
}
