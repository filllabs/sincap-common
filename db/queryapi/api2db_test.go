package queryapi

import (
	"reflect"
	"testing"
)

type Sample struct {
	ID       uint    `db:"ID"`
	Name     string  `qapi:"q:%*%;" db:"Name"`
	InnerFID uint    `db:"InnerFID"`
	InnerF   *Inner1 `qapi:"q:*"`
}

type SamplePoly struct {
	ID       uint    `db:"ID"`
	Name     string  `qapi:"q:*" db:"Name"`
	InnerFID uint    `db:"InnerFID"`
	InnerF   *Inner1 `qapi:"q:*"` // Polymorphic relationship - configured via JoinRegistry
}

type SampleM2M struct {
	ID        uint      `db:"ID"`
	Name      string    `db:"Name"`
	Inner2sID uint      `db:"Inner2sID"`
	Inner2s   []*Inner2 `qapi:"q:*"` // Many-to-many relationship - configured via JoinRegistry
}

type Inner1 struct {
	ID        uint   `db:"ID"`
	Name      string `qapi:"q:%*;" db:"Name"`
	Inner2FID uint   `db:"Inner2FID"`
	Inner2F   *Inner2
	Inner2PID uint    `db:"Inner2PID"`
	Inner2P   *Inner2 `qapi:"q:*;"` // Polymorphic relationship - configured via JoinRegistry
}

type Inner2 struct {
	ID   uint   `db:"ID"`
	Name string `qapi:"q:*%;" db:"Name"`
	Age  uint   `db:"Age"`
}

type SampleName struct {
	ID   uint   `db:"ID"`
	Name string `qapi:"q:*%;" db:"Name"`
	Age  uint   `db:"Age"`
}

func (SampleName) TableName() string {
	return "Sample"
}

func Test_getTableName(t *testing.T) {
	type args struct {
		e any
	}
	tests := []struct {
		name  string
		args  args
		want  reflect.Type
		want1 string
	}{
		{name: "name from type", args: args{e: Sample{}}, want: reflect.TypeOf(Sample{}), want1: "Sample"},
		{name: "name from type", args: args{e: SampleName{}}, want: reflect.TypeOf(SampleName{}), want1: "Sample"},
		{name: "name from primitive", args: args{e: "Test"}, want: reflect.TypeOf(""), want1: "string"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := GetTableName(tt.args.e)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getTableName() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("getTableName() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}
