package stringutils

import (
	"encoding/json"
	"fmt"
)

// ObjectToJsonString converts any object to a JSON string.
// If serialization fails, it returns an empty string.
func ObjectToJsonString(obj any) string {
	if obj == nil {
		return ""
	}
	bytes, err := json.Marshal(obj)
	if err != nil {
		// In a real scenario, we will use loggerutils. For now, output or return empty.
		return ""
	}
	return string(bytes)
}

// JsonStringToObject parses a JSON string into a variable of type T.
func JsonStringToObject[T any](jsonStr string) (T, error) {
	var target T
	if jsonStr == "" {
		return target, fmt.Errorf("empty json string")
	}
	err := json.Unmarshal([]byte(jsonStr), &target)
	return target, err
}

// ConvertValue converts one type to another by serializing to JSON and deserializing.
// This mimics Jackson's ObjectMapper.convertValue in Java.
func ConvertValue[T any](source any) (T, error) {
	var target T
	if source == nil {
		return target, nil
	}
	bytes, err := json.Marshal(source)
	if err != nil {
		return target, err
	}
	err = json.Unmarshal(bytes, &target)
	return target, err
}
