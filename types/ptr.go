package types

import "reflect"

//PtrGetElem returns the value of the given variable even if it is a ptr
func PtrGetElem(t interface{}) interface{} {
	v := reflect.ValueOf(t)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	return v.Interface()
}
