package reflection

import "reflect"

func ExtractRealType(field reflect.Type) reflect.Type {
	if field.Kind() == reflect.Ptr {
		return field.Elem()
	}
	return field
}
