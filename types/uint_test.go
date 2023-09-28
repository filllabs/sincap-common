package types_test

import (
	"testing"

	"github.com/filllabs/sincap-common/types"
)

func TestParseUint(t *testing.T) {
	type args struct {
		val string
	}
	tests := []struct {
		name    string
		args    args
		want    uint
		wantErr bool
	}{
		{name: "valid", args: args{val: "42"}, want: 42},
		{name: "invalid", args: args{val: "life"}, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := types.ParseUint(tt.args.val)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseUint() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ParseUint() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFormatUint(t *testing.T) {
	type args struct {
		val uint
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{name: "valid", args: args{val: 42}, want: "42"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := types.FormatUint(tt.args.val); got != tt.want {
				t.Errorf("FormatUint() = %v, want %v", got, tt.want)
			}
		})
	}
}
