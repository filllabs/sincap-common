package queryapi

import (
	"reflect"
	"testing"

	"github.com/filllabs/sincap-common/db/util"
)

type Sample struct {
	ID       uint
	Name     string `qapi:"q:%*%;"`
	InnerFID uint
	InnerF   *Inner1 `qapi:"q:*"`
}

type SamplePoly struct {
	ID     uint
	Name   string  `qapi:"q:*"`
	InnerF *Inner1 `gorm:"polymorphic:Holder;" qapi:"q:*"`
}

type SampleM2M struct {
	ID      uint
	Name    string
	Inner2s []*Inner2 `gorm:"many2many:SampleM2MInner2"  qapi:"q:*"`
}
type Inner1 struct {
	ID uint
	util.PolymorphicModel
	Name      string `qapi:"q:%*;"`
	Inner2FID uint
	Inner2F   *Inner2
	Inner2P   *Inner2 `gorm:"polymorphic:Holder;" qapi:"q:*;"`
}
type Inner2 struct {
	ID   uint
	Name string `qapi:"q:*%;"`
	Age  uint
}
type SampleName struct {
	ID   uint
	Name string `qapi:"q:*%;"`
	Age  uint
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
