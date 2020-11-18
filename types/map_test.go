package types

import (
	"reflect"
	"sort"
	"testing"
)

type args struct {
	data *map[string]interface{}
}

type test struct {
	name string
	args args
	want []string
}

func TestMapToStrings(t *testing.T) {
	data := make(map[string]interface{})
	data["a"] = "a"
	data["b"] = "b"
	data["c"] = 1
	tests := []test{
		{
			name: "hybrid",
			args: args{
				data: &data,
			},
			want: []string{"1", "a", "b"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ValuesToStrings(tt.args.data)
			sort.Strings(got)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MapToStrings() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestKeysString(t *testing.T) {
	data := make(map[string]interface{})
	data["a"] = "a"
	data["b"] = "b"
	data["c"] = 1
	tests := []test{
		{name: "string", args: args{data: &data}, want: []string{"a", "b", "c"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := KeysString(tt.args.data)
			sort.Strings(got)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Keys() = %v, want %v", got, tt.want)
			}
		})
	}
}
