package logging

import (
	"reflect"

	"gitlab.com/sincap/sincap-common/reflection"
	"go.uber.org/zap/zapcore"
)

// InterfaceArrayMarshaller is a marshaller for interface arrays. It extracts type names of the given array
type InterfaceArrayMarshaller struct {
	Arr []interface{}
}

// MarshalLogArray writes the defined array
func (a *InterfaceArrayMarshaller) MarshalLogArray(enc zapcore.ArrayEncoder) error {
	for _, m := range a.Arr {
		name := reflection.ExtractRealTypeField(reflect.TypeOf(m)).Name()
		enc.AppendString(name)
	}
	return nil
}
