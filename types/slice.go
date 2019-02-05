package types

import (
	"reflect"
)

// SliceContains checks if a slice contains an element
func SliceContains(s interface{}, e interface{}) bool {
	slice := convertSliceToInterface(s)
	e = PtrGetElem(e)
	for _, a := range slice {
		a = PtrGetElem(a)
		if a == e {
			return true
		}
	}
	return false
}

// convertSliceToInterface takes a slice passed in as an interface{}
// then converts the slice to a slice of interfaces
func convertSliceToInterface(s interface{}) (slice []interface{}) {
	s = PtrGetElem(s)
	v := reflect.ValueOf(s)
	if v.Kind() != reflect.Slice {
		return nil
	}

	length := v.Len()
	slice = make([]interface{}, length)
	for i := 0; i < length; i++ {
		slice[i] = v.Index(i).Interface()
	}

	return slice
}
