package stringutils

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/spf13/cast"

	"github.com/YasarKaan/go-kit/exceptions"
)

// GetParameterMap merges query, body, and path parameter maps into a single map.
func GetParameterMap(queryParams, bodyParams, pathParams map[string]any) map[string]any {
	paramMap := make(map[string]any)
	for k, v := range queryParams {
		paramMap[k] = v
	}
	for k, v := range bodyParams {
		paramMap[k] = v
	}
	for k, v := range pathParams {
		paramMap[k] = v
	}
	return paramMap
}

// GetStringValueFromMap extracts string value from map.
func GetStringValueFromMap(m map[string]any, key string) string {
	val, ok := m[key]
	if !ok || val == nil {
		return ""
	}
	parsed, err := cast.ToStringE(val)
	if err != nil {
		return ""
	}
	return parsed
}

// GetStringValueFromMapWithDefault extracts string value with a default.
func GetStringValueFromMapWithDefault(m map[string]any, key string, defaultValue string) string {
	val, ok := m[key]
	if !ok || val == nil {
		return defaultValue
	}
	parsed, err := cast.ToStringE(val)
	if err != nil {
		return defaultValue
	}
	return parsed
}

// GetLongValueFromMap extracts int64 value from map.
func GetLongValueFromMap(m map[string]any, key string) *int64 {
	val, ok := m[key]
	if !ok || val == nil {
		return nil
	}
	parsed, err := cast.ToInt64E(val)
	if err != nil {
		return nil
	}
	return &parsed
}

// GetLongValueFromMapWithDefault extracts int64 value with a default.
func GetLongValueFromMapWithDefault(m map[string]any, key string, defaultValue int64) int64 {
	val, ok := m[key]
	if !ok || val == nil {
		return defaultValue
	}
	parsed, err := cast.ToInt64E(val)
	if err != nil {
		return defaultValue
	}
	return parsed
}

// GetDoubleValueFromMap extracts float64 value from map.
func GetDoubleValueFromMap(m map[string]any, key string) *float64 {
	val, ok := m[key]
	if !ok || val == nil {
		return nil
	}
	parsed, err := cast.ToFloat64E(val)
	if err != nil {
		return nil
	}
	return &parsed
}

// GetDoubleValueFromMapWithDefault extracts float64 value with a default.
func GetDoubleValueFromMapWithDefault(m map[string]any, key string, defaultValue float64) float64 {
	val, ok := m[key]
	if !ok || val == nil {
		return defaultValue
	}
	parsed, err := cast.ToFloat64E(val)
	if err != nil {
		return defaultValue
	}
	return parsed
}

// GetBooleanValueFromMap extracts bool value from map.
func GetBooleanValueFromMap(m map[string]any, key string) *bool {
	val, ok := m[key]
	if !ok || val == nil {
		return nil
	}
	parsed, err := cast.ToBoolE(val)
	if err != nil {
		return nil
	}
	return &parsed
}

// GetBooleanValueFromMapWithDefault extracts bool value with default.
func GetBooleanValueFromMapWithDefault(m map[string]any, key string, defaultValue bool) bool {
	val, ok := m[key]
	if !ok || val == nil {
		return defaultValue
	}
	parsed, err := cast.ToBoolE(val)
	if err != nil {
		return defaultValue
	}
	return parsed
}

// GetMapValueFromMap extracts nested map[string]any.
func GetMapValueFromMap(m map[string]any, key string) map[string]any {
	val, ok := m[key]
	if !ok || val == nil {
		return nil
	}
	parsed, err := cast.ToStringMapE(val)
	if err != nil {
		return nil
	}
	return parsed
}

// GetMapValueFromMapWithDefault extracts nested map[string]any with default.
func GetMapValueFromMapWithDefault(m map[string]any, key string, defaultValue map[string]any) map[string]any {
	val, ok := m[key]
	if !ok || val == nil {
		return defaultValue
	}
	parsed, err := cast.ToStringMapE(val)
	if err != nil {
		return defaultValue
	}
	return parsed
}

// GetArrayListValueFromMap extracts slice of any.
func GetArrayListValueFromMap(m map[string]any, key string) []any {
	val, ok := m[key]
	if !ok || val == nil {
		return nil
	}
	parsed, err := cast.ToSliceE(val)
	if err != nil {
		return nil
	}
	return parsed
}

// GetArrayListValueFromMapWithDefault extracts slice of any with default.
func GetArrayListValueFromMapWithDefault(m map[string]any, key string, defaultValue []any) []any {
	val, ok := m[key]
	if !ok || val == nil {
		return defaultValue
	}
	parsed, err := cast.ToSliceE(val)
	if err != nil {
		return defaultValue
	}
	return parsed
}

// GetIntegerValueFromMap extracts int value from map.
func GetIntegerValueFromMap(m map[string]any, key string) *int {
	val, ok := m[key]
	if !ok || val == nil {
		return nil
	}
	parsed, err := cast.ToIntE(val)
	if err != nil {
		return nil
	}
	return &parsed
}

// GetIntegerValueFromMapWithDefault extracts int value with default.
func GetIntegerValueFromMapWithDefault(m map[string]any, key string, defaultValue int) int {
	val, ok := m[key]
	if !ok || val == nil {
		return defaultValue
	}
	parsed, err := cast.ToIntE(val)
	if err != nil {
		return defaultValue
	}
	return parsed
}

// Helper to create 422 CustomWebServerException errors.
func err422(msg string) error {
	return exceptions.NewCustomWebServerException(422, msg, nil)
}

// GetStringValueFromMapWithException extracts string or returns exception.
func GetStringValueFromMapWithException(m map[string]any, key string, customMessage string) (string, error) {
	val, ok := m[key]
	if !ok || val == nil {
		return "", err422(customMessage)
	}
	parsed, err := cast.ToStringE(val)
	if err != nil {
		return "", err422(customMessage)
	}
	return parsed, nil
}

// GetLongValueFromMapWithException extracts int64 or returns exception.
func GetLongValueFromMapWithException(m map[string]any, key string, customMessage string) (int64, error) {
	val, ok := m[key]
	if !ok || val == nil {
		return 0, err422(customMessage)
	}
	parsed, err := cast.ToInt64E(val)
	if err != nil {
		return 0, err422(customMessage)
	}
	return parsed, nil
}

// GetDoubleValueFromMapWithException extracts float64 or returns exception.
func GetDoubleValueFromMapWithException(m map[string]any, key string, customMessage string) (float64, error) {
	val, ok := m[key]
	if !ok || val == nil {
		return 0.0, err422(customMessage)
	}
	parsed, err := cast.ToFloat64E(val)
	if err != nil {
		return 0.0, err422(customMessage)
	}
	return parsed, nil
}

// GetBooleanValueFromMapWithException extracts bool or returns exception.
func GetBooleanValueFromMapWithException(m map[string]any, key string, customMessage string) (bool, error) {
	val, ok := m[key]
	if !ok || val == nil {
		return false, err422(customMessage)
	}
	parsed, err := cast.ToBoolE(val)
	if err != nil {
		return false, err422(customMessage)
	}
	return parsed, nil
}

// GetIntegerValueFromMapWithException extracts int or returns exception.
func GetIntegerValueFromMapWithException(m map[string]any, key string, customMessage string) (int, error) {
	val, ok := m[key]
	if !ok || val == nil {
		return 0, err422(customMessage)
	}
	parsed, err := cast.ToIntE(val)
	if err != nil {
		return 0, err422(customMessage)
	}
	return parsed, nil
}

// GetMapValueFromMapWithException extracts map[string]any or returns exception.
func GetMapValueFromMapWithException(m map[string]any, key string, customMessage string) (map[string]any, error) {
	val, ok := m[key]
	if !ok || val == nil {
		return nil, err422(customMessage)
	}
	parsed, err := cast.ToStringMapE(val)
	if err != nil {
		return nil, err422(customMessage)
	}
	return parsed, nil
}

// GetArrayListValueFromMapWithException extracts []any or returns exception.
func GetArrayListValueFromMapWithException(m map[string]any, key string, customMessage string) ([]any, error) {
	val, ok := m[key]
	if !ok || val == nil {
		return nil, err422(customMessage)
	}
	parsed, err := cast.ToSliceE(val)
	if err != nil {
		return nil, err422(customMessage)
	}
	return parsed, nil
}

// GetUUIDValueFromMap extracts UUID from map.
func GetUUIDValueFromMap(m map[string]any, key string) (uuid.UUID, error) {
	val, ok := m[key]
	if !ok || val == nil {
		return uuid.Nil, fmt.Errorf("key %s not found", key)
	}
	switch v := val.(type) {
	case uuid.UUID:
		return v, nil
	case []byte:
		return uuid.Parse(string(v))
	case string:
		return uuid.Parse(v)
	default:
		return uuid.Parse(cast.ToString(val))
	}
}

// GetUUIDValueFromMapWithDefault extracts UUID with default.
func GetUUIDValueFromMapWithDefault(m map[string]any, key string, defaultValue uuid.UUID) uuid.UUID {
	uid, err := GetUUIDValueFromMap(m, key)
	if err != nil {
		return defaultValue
	}
	return uid
}

// GetUUIDValueFromMapWithException extracts UUID or returns exception.
func GetUUIDValueFromMapWithException(m map[string]any, key string, customMessage string) (uuid.UUID, error) {
	uid, err := GetUUIDValueFromMap(m, key)
	if err != nil {
		return uuid.Nil, err422(customMessage)
	}
	return uid, nil
}

// GetDateFromMap parses a date string using the specified layout (format).
func GetDateFromMap(m map[string]any, key string, layout string) (time.Time, error) {
	valStr := GetStringValueFromMap(m, key)
	if valStr == "" {
		return time.Time{}, fmt.Errorf("empty date string")
	}
	return time.Parse(layout, valStr)
}

// GetDateFromMapWithDefault parses a date string or returns a default.
func GetDateFromMapWithDefault(m map[string]any, key string, layout string, defaultValue time.Time) time.Time {
	t, err := GetDateFromMap(m, key, layout)
	if err != nil {
		return defaultValue
	}
	return t
}

// GetDateFromMapWithException parses a date or returns 422 error.
func GetDateFromMapWithException(m map[string]any, key string, layout string, customMessage string) (time.Time, error) {
	t, err := GetDateFromMap(m, key, layout)
	if err != nil {
		return time.Time{}, err422(customMessage)
	}
	return t, nil
}

// GetOffsetDateTimeFromMap parses a date time with timezone offset.
func GetOffsetDateTimeFromMap(m map[string]any, key string) *time.Time {
	val, ok := m[key]
	if !ok || val == nil {
		return nil
	}

	switch v := val.(type) {
	case time.Time:
		return &v
	case string:
		if v == "" {
			return nil
		}
		// Attempt RFC3339 parsing which is typical for OffsetDateTime (e.g. 2026-05-25T20:00:00Z)
		t, err := time.Parse(time.RFC3339, v)
		if err == nil {
			return &t
		}
		// Try standard layouts if RFC3339 fails
		t, err = time.Parse("2006-01-02 15:04:05", v)
		if err == nil {
			return &t
		}
	}
	return nil
}
