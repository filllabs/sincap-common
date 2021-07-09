package queryapi

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"

	"gitlab.com/sincap/sincap-common/db/util"
	"gitlab.com/sincap/sincap-common/resources/query"
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
	Inner2s []*Inner2 `gorm:"many2many:SampleM2MInner2"`
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

func TestFilter2Sql1Level(t *testing.T) {
	typ := reflect.TypeOf(Sample{})
	q := query.Query{Filter: []query.Filter{{Name: "InnerF.Name", Operation: query.EQ, Value: "Osman"}}}
	where, values, err := filter2Sql(q.Filter, typ)
	assert.NoError(t, err)
	assert.Equal(t, "Osman", values[0])
	assert.Equal(t, "InnerFID IN ( SELECT ID FROM Inner1 WHERE ( Name = ? ) )", where)
}

func TestFilter2Sql2Level(t *testing.T) {
	typ := reflect.TypeOf(Sample{})
	q := query.Query{Filter: []query.Filter{{Name: "InnerF.Inner2F.Name", Operation: query.EQ, Value: "Osman"}}}
	where, values, err := filter2Sql(q.Filter, typ)
	assert.NoError(t, err)
	assert.Equal(t, "Osman", values[0])
	assert.Equal(t, "InnerFID IN ( SELECT ID FROM Inner1 WHERE ( Inner2FID IN ( SELECT ID FROM Inner2 WHERE ( Name = ? ) ) ) )", where)
}

func TestFilter2Sql2LevelUint(t *testing.T) {
	typ := reflect.TypeOf(Sample{})
	q := query.Query{Filter: []query.Filter{{Name: "InnerF.Inner2F.Age", Operation: query.EQ, Value: "18"}}}
	where, values, err := filter2Sql(q.Filter, typ)
	assert.NoError(t, err)
	assert.Equal(t, uint64(18), values[0])
	assert.Equal(t, "InnerFID IN ( SELECT ID FROM Inner1 WHERE ( Inner2FID IN ( SELECT ID FROM Inner2 WHERE ( Age = ? ) ) ) )", where)
}

func TestFilter2SqlPoly1Level(t *testing.T) {
	typ := reflect.TypeOf(SamplePoly{})
	q := query.Query{Filter: []query.Filter{{Name: "InnerF.Name", Operation: query.EQ, Value: "Osman"}}}
	where, values, err := filter2Sql(q.Filter, typ)
	assert.NoError(t, err)
	assert.Equal(t, "Osman", values[0])
	assert.Equal(t, "ID IN ( SELECT HolderID FROM Inner1 WHERE ( Name = ? AND HolderID = `SamplePoly`.ID AND HolderType = 'SamplePoly' ) )", where)
}

func TestFilter2SqlPoly2Level(t *testing.T) {
	typ := reflect.TypeOf(SamplePoly{})
	q := query.Query{Filter: []query.Filter{{Name: "InnerF.Inner2P.Name", Operation: query.EQ, Value: "Osman"}}}
	where, values, err := filter2Sql(q.Filter, typ)
	assert.NoError(t, err)
	assert.Equal(t, "Osman", values[0])
	assert.Equal(t, "ID IN ( SELECT HolderID FROM Inner1 WHERE ( ID IN ( SELECT HolderID FROM Inner2 WHERE ( Name = ? AND HolderID = `Inner1`.ID AND HolderType = 'Inner1' ) ) AND HolderID = `SamplePoly`.ID AND HolderType = 'SamplePoly' ) )", where)
}
func TestFilter2SqlPM2M(t *testing.T) {
	typ := reflect.TypeOf(SampleM2M{})
	q := query.Query{Filter: []query.Filter{{Name: "Inner2s.Name", Operation: query.EQ, Value: "Osman"}}}
	where, values, err := filter2Sql(q.Filter, typ)
	assert.NoError(t, err)
	assert.Equal(t, "Osman", values[0])
	assert.Equal(t, "ID IN ( SELECT SampleM2M_ID FROM SampleM2MInner2 WHERE ( Inner2_ID IN ( SELECT ID FROM Inner2 WHERE ( Name = ? ) ) ) )", where)
}
func TestQ2Sql(t *testing.T) {
	typ := reflect.TypeOf(Sample{})
	q := query.Query{Q: "seray"}
	where, values, err := q2Sql(q.Q, typ)
	assert.NoError(t, err)
	assert.Equal(t, "%seray%", values[0])
	assert.Equal(t, "%seray", values[1])
	assert.Equal(t, "seray%", values[2])
	assert.Equal(t, "Name LIKE ? OR InnerFID IN ( SELECT ID FROM Inner1 WHERE ( Name LIKE ? OR ID IN ( SELECT HolderID FROM Inner2 WHERE ( Name LIKE ? ) ) ) )", where)
}
func TestQ2SqlPoly(t *testing.T) {
	typ := reflect.TypeOf(SamplePoly{})
	q := query.Query{Q: "seray"}
	where, values, err := q2Sql(q.Q, typ)
	assert.NoError(t, err)
	assert.Equal(t, "seray", values[0])
	assert.Equal(t, "%seray", values[1])
	assert.Equal(t, "seray%", values[2])
	assert.Equal(t, "Name LIKE ? OR ID IN ( SELECT HolderID FROM Inner1 WHERE ( Name LIKE ? OR ID IN ( SELECT HolderID FROM Inner2 WHERE ( Name LIKE ? ) ) ) )", where)
}