// Package reflection provides a set of everyday needed functions over go's reflect package
package reflection

import "reflect"

// ExtractRealTypeField helps to extract the real type of the given field.
// All combinations of pointers and slices returns the inner type
func ExtractRealTypeField(field reflect.Type) reflect.Type {
	if field.Kind() == reflect.Ptr || field.Kind() == reflect.Slice {
		return ExtractRealTypeField(field.Elem())
	}
	return field
}

// DepointerField helps to extract depointer the given field
func DepointerField(field reflect.Type) reflect.Type {
	if field.Kind() == reflect.Ptr {
		return DepointerField(field.Elem())
	}
	return field
}

// DepointerInteface helps to extract depointer the given interface
func DepointerInteface(t interface{}) interface{} {
	v := reflect.ValueOf(t)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	return v.Interface()
}
