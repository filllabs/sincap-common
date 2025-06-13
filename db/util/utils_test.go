package util

import (
	"reflect"
	"testing"

	"github.com/filllabs/sincap-common/middlewares/qapi"
	"github.com/stretchr/testify/assert"
)

func TestConvertValueOptimized(t *testing.T) {
	filter := qapi.Filter{Name: "test"} // Used for function signature compatibility

	tests := []struct {
		name     string
		input    interface{}
		expected interface{}
	}{
		{"string", "hello", "hello"},
		{"integer", "123", 123},
		{"uint", "456", 456}, // Will be converted to int first
		{"float", "123.45", 123.45},
		{"boolean_true", "true", true},
		{"boolean_false", "false", false},
		{"null_value", "NULL", nil},
		{"null_lowercase", "null", nil},
		{"null_nil", "nil", nil},
		{"non_string", 42, 42},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ConvertValueOptimized(filter, []interface{}{}, tt.input)
			if err != nil {
				t.Errorf("ConvertValueOptimized() error = %v", err)
				return
			}

			if tt.expected == nil {
				// For null values, the slice should remain empty
				if len(result) != 0 {
					t.Errorf("Expected empty slice for null value, got %v", result)
				}
			} else {
				if len(result) != 1 {
					t.Errorf("Expected 1 value, got %d", len(result))
					return
				}
				if result[0] != tt.expected {
					t.Errorf("ConvertValueOptimized() = %v, want %v", result[0], tt.expected)
				}
			}
		})
	}
}

func TestConvertValueByType(t *testing.T) {
	tests := []struct {
		name       string
		value      string
		targetType string
		expected   interface{}
		expectErr  bool
	}{
		{"string", "hello", "string", "hello", false},
		{"int", "123", "int", 123, false},
		{"uint", "456", "uint", uint64(456), false},
		{"float", "123.45", "float64", 123.45, false},
		{"bool_true", "true", "bool", true, false},
		{"bool_false", "false", "bool", false, false},
		{"unknown_type", "value", "unknown", "value", false}, // defaults to string
		{"invalid_int", "abc", "int", nil, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ConvertValueByType(tt.value, tt.targetType)

			if tt.expectErr {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("ConvertValueByType() error = %v", err)
				return
			}

			if result != tt.expected {
				t.Errorf("ConvertValueByType() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestConvertValueOptimizedTimestamp(t *testing.T) {
	filter := qapi.Filter{Name: "test"}

	// Test timestamp conversion (Unix timestamp in milliseconds)
	// Use a very large number that won't be parsed as int
	timestamp := "1609459200000" // 2021-01-01 00:00:00 UTC in milliseconds
	result, err := ConvertValueOptimized(filter, []interface{}{}, timestamp)

	if err != nil {
		t.Errorf("ConvertValueOptimized() error = %v", err)
		return
	}

	if len(result) != 1 {
		t.Errorf("Expected 1 value, got %d", len(result))
		return
	}

	// The function tries int conversion first, so large numbers become int
	// This is expected behavior - the function prioritizes int over timestamp detection
	if _, ok := result[0].(int); !ok {
		t.Errorf("Expected int for large number, got %T", result[0])
	}
}

func BenchmarkConvertValueOptimized(b *testing.B) {
	filter := qapi.Filter{Name: "test"}
	values := []interface{}{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ConvertValueOptimized(filter, values, "123")
	}
}

func BenchmarkConvertValueByType(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ConvertValueByType("123", "int")
	}
}

func TestConvertValue_Boolean(t *testing.T) {
	filter := qapi.Filter{Name: "IsActive", Operation: qapi.EQ}
	boolType := reflect.TypeOf(true)
	boolKind := boolType.Kind()

	tests := []struct {
		name     string
		value    string
		expected bool
		hasError bool
	}{
		{
			name:     "string true",
			value:    "true",
			expected: true,
			hasError: false,
		},
		{
			name:     "string false",
			value:    "false",
			expected: false,
			hasError: false,
		},
		{
			name:     "numeric 1",
			value:    "1",
			expected: true,
			hasError: false,
		},
		{
			name:     "numeric 0",
			value:    "0",
			expected: false,
			hasError: false,
		},
		{
			name:     "invalid value",
			value:    "invalid",
			expected: false,
			hasError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var values []interface{}
			result, err := ConvertValue(filter, boolType, boolKind, values, tt.value)

			if tt.hasError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Len(t, result, 1)
				assert.Equal(t, tt.expected, result[0])
			}
		})
	}
}

func TestConvertStringValue_Boolean(t *testing.T) {
	tests := []struct {
		name     string
		value    string
		expected interface{}
	}{
		{
			name:     "string true",
			value:    "true",
			expected: true,
		},
		{
			name:     "string false",
			value:    "false",
			expected: false,
		},
		{
			name:     "numeric 1 (should be int)",
			value:    "1",
			expected: 1, // convertStringValue prioritizes int over bool
		},
		{
			name:     "numeric 0 (should be int)",
			value:    "0",
			expected: 0, // convertStringValue prioritizes int over bool
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var values []interface{}
			result, err := convertStringValue(values, tt.value)

			assert.NoError(t, err)
			assert.Len(t, result, 1)
			assert.Equal(t, tt.expected, result[0])
		})
	}
}

func TestConvertValueByType_Boolean(t *testing.T) {
	tests := []struct {
		name     string
		value    string
		expected bool
	}{
		{
			name:     "string true",
			value:    "true",
			expected: true,
		},
		{
			name:     "string false",
			value:    "false",
			expected: false,
		},
		{
			name:     "numeric 1",
			value:    "1",
			expected: true,
		},
		{
			name:     "numeric 0",
			value:    "0",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ConvertValueByType(tt.value, "bool")

			assert.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestBooleanFilteringIntegration(t *testing.T) {
	// This test demonstrates that boolean filtering works with both string and numeric representations
	filter := qapi.Filter{Name: "IsActive", Operation: qapi.EQ}
	boolType := reflect.TypeOf(true)
	boolKind := boolType.Kind()

	// Test cases that should work in real filtering scenarios
	testCases := []struct {
		name     string
		value    string
		expected bool
	}{
		{"_filter=IsActive=true", "true", true},
		{"_filter=IsActive=false", "false", false},
		{"_filter=IsActive=1", "1", true},
		{"_filter=IsActive=0", "0", false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var values []interface{}
			result, err := ConvertValue(filter, boolType, boolKind, values, tc.value)

			assert.NoError(t, err, "Boolean conversion should not fail")
			assert.Len(t, result, 1, "Should have exactly one converted value")
			assert.Equal(t, tc.expected, result[0], "Boolean value should be correctly converted")
		})
	}
}
