package types

import (
	"reflect"

	"github.com/filllabs/sincap-common/reflection"
)

// SliceContains checks if a slice contains an element
func SliceContains(s interface{}, e interface{}) bool {
	slice := convertSliceToInterface(s)
	e = reflection.DepointerInteface(e)
	for _, a := range slice {
		a = reflection.DepointerInteface(a)
		if a == e {
			return true
		}
	}
	return false
}

// SliceContainsDeep checks if a slice contains an element with reflection
func SliceContainsDeep(s interface{}, e interface{}) bool {
	slice := convertSliceToInterface(s)
	e = reflection.DepointerInteface(e)
	for _, a := range slice {
		a = reflection.DepointerInteface(a)
		if reflect.DeepEqual(a, e) {
			return true
		}
	}
	return false
}

// convertSliceToInterface takes a slice passed in as an interface{}
// then converts the slice to a slice of interfaces
func convertSliceToInterface(s interface{}) (slice []interface{}) {
	s = reflection.DepointerInteface(s)
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

// SliceOfString creates a slice of the given item with the size of the given count.
func SliceOfString(item string, count int) []string {
	sl := make([]string, count)
	for i := 0; i < count; i++ {
		sl[i] = item
	}
	return sl
}
