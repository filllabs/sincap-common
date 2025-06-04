package util

import (
	"fmt"
	"reflect"
	"strconv"
	"time"

	"github.com/filllabs/sincap-common/middlewares/qapi"
)

var timeKind = reflect.TypeOf(time.Time{}).Kind()

// ValueConverter interface for types that can convert themselves
type ValueConverter interface {
	ConvertFromString(value string) (interface{}, error)
}

// ConvertValueOptimized converts values with reduced reflection usage
func ConvertValueOptimized(filter qapi.Filter, values []interface{}, value interface{}) ([]interface{}, error) {
	if value == "NULL" || value == "null" || value == "nil" {
		// Do not add anything
		return values, nil
	}

	// Try string conversion first (most common case)
	if strValue, ok := value.(string); ok {
		return convertStringValue(values, strValue)
	}

	// For non-string values, add directly
	values = append(values, value)
	return values, nil
}

// convertStringValue handles string to type conversion without reflection
func convertStringValue(values []interface{}, value string) ([]interface{}, error) {
	// Try common type conversions without reflection

	// Integer types
	if i, err := strconv.Atoi(value); err == nil {
		// Could be any integer type, but int is most common
		values = append(values, i)
		return values, nil
	}

	// Unsigned integer types
	if i, err := strconv.ParseUint(value, 10, 64); err == nil {
		values = append(values, i)
		return values, nil
	}

	// Float types
	if f, err := strconv.ParseFloat(value, 64); err == nil {
		values = append(values, f)
		return values, nil
	}

	// Boolean
	if value == "true" || value == "false" {
		values = append(values, value == "true")
		return values, nil
	}

	// Time (Unix timestamp in milliseconds)
	if i, err := strconv.ParseInt(value, 10, 64); err == nil {
		// Assume it's a timestamp if it's a large number
		if i > 1000000000 { // Rough check for timestamp
			values = append(values, time.Unix(0, i*int64(time.Millisecond)))
			return values, nil
		}
	}

	// Default to string
	values = append(values, value)
	return values, nil
}

// ConvertValue is the original reflection-based function (kept for backward compatibility)
func ConvertValue(filter qapi.Filter, typ reflect.Type, kind reflect.Kind, values []interface{}, value interface{}) ([]interface{}, error) {
	if value == "NULL" || value == "null" || value == "nil" {
		// Do not add anything
		return values, nil
	}
	switch kind {
	case reflect.String:
		values = append(values, value)
	case reflect.Int,
		reflect.Int8,
		reflect.Int16,
		reflect.Int32,
		reflect.Int64:
		if i, e := strconv.Atoi(value.(string)); e == nil {
			values = append(values, i)
		}
	case reflect.Uint,
		reflect.Uint8,
		reflect.Uint16,
		reflect.Uint32,
		reflect.Uint64:
		if i, e := strconv.ParseUint(value.(string), 10, 64); e == nil {
			values = append(values, i)
		}
	case reflect.Float32,
		reflect.Float64:
		if i, e := strconv.ParseFloat(value.(string), 64); e == nil {
			values = append(values, i)
		}
	case reflect.Bool:
		values = append(values, value.(string) == "true")
	case timeKind:
		i, err := strconv.ParseInt(value.(string), 10, 64)
		if err != nil {
			return nil, fmt.Errorf("QApi cannot parse date: %s for %s. Cause: %v", value.(string), filter.Name, err)
		}
		values = append(values, time.Unix(0, i*int64(time.Millisecond)))
	default:
		return nil, fmt.Errorf("field type not supported for QApi %s : %s", typ.Name(), filter.Name)
	}
	return values, nil
}

// ConvertValueByType converts a value based on a target type without reflection
func ConvertValueByType(value string, targetType string) (interface{}, error) {
	switch targetType {
	case "string":
		return value, nil
	case "int", "int32", "int64":
		return strconv.Atoi(value)
	case "uint", "uint32", "uint64":
		i, err := strconv.ParseUint(value, 10, 64)
		return uint64(i), err
	case "float32", "float64":
		return strconv.ParseFloat(value, 64)
	case "bool":
		return value == "true", nil
	case "time":
		i, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return nil, err
		}
		return time.Unix(0, i*int64(time.Millisecond)), nil
	default:
		return value, nil // Default to string
	}
}
