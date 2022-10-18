package queryapi

import (
	"reflect"
	"testing"

	"gitlab.com/sincap/sincap-common/reflection"
)

func Test_getQapiFields(t *testing.T) {
	type args struct {
		structType reflect.Type
	}
	typ := reflect.TypeOf(Sample{})
	nameField, _ := typ.FieldByName("Name")
	innerfField, _ := typ.FieldByName("InnerF")

	want := []pair{
		{
			TableName: "string",
			Typ:       reflection.ExtractRealTypeField(nameField.Type),
			Tag:       "%*%",
			Field:     nameField,
		},
		{
			TableName: "Inner1",
			Typ:       reflection.ExtractRealTypeField(innerfField.Type),
			Tag:       "*",
			Field:     innerfField,
		},
	}
	tests := []struct {
		name string
		args args
		want []pair
	}{
		{
			name: "q tagged fields",
			args: args{
				structType: typ,
			},
			want: want,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getQapiFields(tt.args.structType); reflect.DeepEqual(got, tt.want) {
				t.Errorf("getQapiFields() = %+v, want %+v", got, tt.want)
			}
			if got := getQapiFields(tt.args.structType); reflect.DeepEqual(got, tt.want) {
				t.Errorf("getQapiFields() = %+v, want %+v", got, tt.want)
			}

		})
	}
}
