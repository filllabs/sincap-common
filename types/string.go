package types

import (
	"fmt"
	"strconv"
)

// ToString converts the given variable to string
func ToString(from interface{}) (string, error) {
	switch val := from.(type) {
	case string:
		return val, nil
	case int:
		return strconv.Itoa(val), nil
	case int8:
		return strconv.Itoa(int(val)), nil
	case int16:
		return strconv.Itoa(int(val)), nil
	case int32:
		return strconv.Itoa(int(val)), nil
	case int64:
		return strconv.Itoa(int(val)), nil
	case uint:
		return strconv.FormatUint(uint64(val), 10), nil
	case uint8:
		return strconv.FormatUint(uint64(val), 10), nil
	case uint16:
		return strconv.FormatUint(uint64(val), 10), nil
	case uint32:
		return strconv.FormatUint(uint64(val), 10), nil
	case uint64:
		return strconv.FormatUint(val, 10), nil
	case float32, float64:
		return fmt.Sprintf("%f", from), nil
	case bool:
		if val {
			return "true", nil
		}
		return "false", nil
	default:
		return fmt.Sprintf("%v", from), nil
	}

}
