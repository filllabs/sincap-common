package types_test

import (
	"testing"

	"github.com/filllabs/sincap-common/types"
)

func TestToString(t *testing.T) {
	type args struct {
		from interface{}
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{name: "string", args: args{from: "String"}, want: "String"},
		{name: "int", args: args{from: 1}, want: "1"},
		{name: "int8", args: args{from: int8(1)}, want: "1"},
		{name: "int16", args: args{from: int16(1)}, want: "1"},
		{name: "int32", args: args{from: int32(1)}, want: "1"},
		{name: "int64", args: args{from: int64(1)}, want: "1"},
		{name: "uint", args: args{from: uint(1)}, want: "1"},
		{name: "uint8", args: args{from: uint8(1)}, want: "1"},
		{name: "uint16", args: args{from: uint16(1)}, want: "1"},
		{name: "uint32", args: args{from: uint32(1)}, want: "1"},
		{name: "uint64", args: args{from: uint64(1)}, want: "1"},
		{name: "float32", args: args{from: float32(1)}, want: "1.000000"},
		{name: "float64", args: args{from: float64(1)}, want: "1.000000"},
		{name: "bool", args: args{from: true}, want: "true"},
		{name: "bool", args: args{from: false}, want: "false"},
		{name: "other", args: args{from: struct{ name string }{name: "Name"}}, want: "{Name}"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := types.ToString(tt.args.from)
			if (err != nil) != tt.wantErr {
				t.Errorf("ToString() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ToString() = %v, want %v", got, tt.want)
			}
		})
	}
}
