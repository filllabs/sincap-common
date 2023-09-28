package types_test

import (
	"reflect"
	"sort"
	"testing"

	"github.com/filllabs/sincap-common/types"
)

type args struct {
	data *map[string]interface{}
}

type test struct {
	name string
	args args
	want []string
}

func TestMapValuesAsStrings(t *testing.T) {
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
			got := types.MapValuesAsStrings(tt.args.data)
			sort.Strings(got)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MapValuesAsStrings() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestKeysAsStrings(t *testing.T) {
	data := make(map[string]interface{})
	data["a"] = "a"
	data["b"] = "b"
	data["c"] = 1
	tests := []test{
		{name: "string", args: args{data: &data}, want: []string{"a", "b", "c"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := types.MapKeysAsStrings(tt.args.data)
			sort.Strings(got)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Keys() = %v, want %v", got, tt.want)
			}
		})
	}
}
