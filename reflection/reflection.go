package reflection

import "reflect"

// ExtractRealType helps to extract the real type of the given field.
// All combinations of pointers and slices returns the inner type
func ExtractRealType(field reflect.Type) reflect.Type {
	if field.Kind() == reflect.Ptr || field.Kind() == reflect.Slice {
		return ExtractRealType(field.Elem())
	}
	return field
}

// Depointer helps to extract depointer the given type
func Depointer(field reflect.Type) reflect.Type {
	if field.Kind() == reflect.Ptr {
		return Depointer(field.Elem())
	}
	return field
}
