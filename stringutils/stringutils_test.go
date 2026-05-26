package stringutils

import (
	"testing"

	"github.com/google/uuid"
)

func TestJSON(t *testing.T) {
	type User struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}

	u := User{Name: "Kaan", Age: 30}
	jsonStr := ObjectToJsonString(u)
	if jsonStr == "" {
		t.Fatal("expected non-empty JSON string")
	}

	// Test ObjectToJsonStringE success
	jsonStrE, err := ObjectToJsonStringE(u)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if jsonStrE != jsonStr {
		t.Errorf("expected %s, got %s", jsonStr, jsonStrE)
	}

	// Test ObjectToJsonStringE failure (channels cannot be marshalled to JSON)
	_, err = ObjectToJsonStringE(make(chan int))
	if err == nil {
		t.Error("expected error for unmarshallable object")
	}

	parsed, err := JsonStringToObject[User](jsonStr)
	if err != nil {
		t.Fatalf("unexpected error parsing JSON: %v", err)
	}

	if parsed.Name != "Kaan" || parsed.Age != 30 {
		t.Errorf("parsed object properties do not match: %+v", parsed)
	}
}

func TestConvertValue(t *testing.T) {
	source := map[string]any{"name": "FunProject", "age": 5}
	type App struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}

	app, err := ConvertValue[App](source)
	if err != nil {
		t.Fatalf("failed to convert value: %v", err)
	}

	if app.Name != "FunProject" || app.Age != 5 {
		t.Errorf("expected converted app properties to match: %+v", app)
	}
}

func TestVariableMap(t *testing.T) {
	uid := uuid.New()
	data := map[string]any{
		"str":     "hello",
		"integer": 42,
		"long":    int64(999999),
		"double":  3.14,
		"boolean": true,
		"uid":     uid.String(),
		"nested":  map[string]any{"key": "val"},
		"arr":     []any{1, 2, 3},
		"date":    "2026-05-25",
	}

	if GetStringValueFromMap(data, "str") != "hello" {
		t.Error("failed to get string")
	}

	if GetStringValueFromMapWithDefault(data, "non-existent", "default") != "default" {
		t.Error("failed to get string with default")
	}

	if *GetIntegerValueFromMap(data, "integer") != 42 {
		t.Error("failed to get integer")
	}

	if GetIntegerValueFromMapWithDefault(data, "non-existent", 10) != 10 {
		t.Error("failed to get integer with default")
	}

	if *GetLongValueFromMap(data, "long") != int64(999999) {
		t.Error("failed to get long")
	}

	if *GetDoubleValueFromMap(data, "double") != 3.14 {
		t.Error("failed to get double")
	}

	if !GetBooleanValueFromMapWithDefault(data, "boolean", false) {
		t.Error("failed to get boolean")
	}

	parsedUid, err := GetUUIDValueFromMap(data, "uid")
	if err != nil || parsedUid != uid {
		t.Errorf("failed to get UUID: %v", err)
	}

	nested := GetMapValueFromMap(data, "nested")
	if nested["key"] != "val" {
		t.Error("failed to get nested map")
	}

	arr := GetArrayListValueFromMap(data, "arr")
	if len(arr) != 3 {
		t.Error("failed to get slice")
	}

	parsedDate, err := GetDateFromMap(data, "date", "2006-01-02")
	if err != nil || parsedDate.Year() != 2026 {
		t.Errorf("failed to get date: %v", err)
	}

	// Exception methods
	val, err := GetStringValueFromMapWithException(data, "str", "error")
	if err != nil || val != "hello" {
		t.Error("failed to get string with exception")
	}

	_, err = GetStringValueFromMapWithException(data, "non-existent", "missing key error")
	if err == nil {
		t.Error("expected exception for missing key")
	}

	// Test strict casting validation (casting non-numeric "hello" to long must fail)
	_, err = GetLongValueFromMapWithException(data, "str", "invalid type error")
	if err == nil {
		t.Error("expected exception when casting non-numeric string to long")
	}
}

func TestOffsetDateTime(t *testing.T) {
	data := map[string]any{
		"offsetTime": "2026-05-25T20:00:00Z",
	}

	got := GetOffsetDateTimeFromMap(data, "offsetTime")
	if got == nil || got.Year() != 2026 || got.Hour() != 20 {
		t.Errorf("expected offset time to be parsed correctly, got: %v", got)
	}

	// Test fallback standard layout
	data2 := map[string]any{
		"offsetTime": "2026-05-25 20:00:00",
	}

	got2 := GetOffsetDateTimeFromMap(data2, "offsetTime")
	if got2 == nil || got2.Year() != 2026 || got2.Hour() != 20 {
		t.Errorf("expected standard time layout to be parsed correctly, got: %v", got2)
	}
}
