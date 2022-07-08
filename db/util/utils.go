package util

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	"gitlab.com/sincap/sincap-common/middlewares/qapi"
)

var timeKind = reflect.TypeOf(time.Time{}).Kind()

// GetMany2Many tries to read the table name of the gorm tag "many2many" from the given field.
func GetMany2Many(f *reflect.StructField) (string, bool) {
	// get gorm tag
	if tag, ok := f.Tag.Lookup("gorm"); ok {
		props := strings.Split(tag, ";")
		// find many2many info
		for _, prop := range props {
			if strings.HasPrefix(prop, "many2many:") {
				return strings.TrimPrefix(prop, "many2many:"), true
			}
		}
	}
	return "", false
}

// GetPolymorphic tries to read the table name of the gorm tag "polymorphic" from the given field.
func GetPolymorphic(f *reflect.StructField) (string, bool) {
	// get gorm tag
	if tag, ok := f.Tag.Lookup("gorm"); ok {
		props := strings.Split(tag, ";")
		// find polymorphic info
		for _, prop := range props {
			if strings.HasPrefix(prop, "polymorphic:") {
				return strings.TrimPrefix(prop, "polymorphic:"), true
			}
		}
	}
	return "", false
}

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
