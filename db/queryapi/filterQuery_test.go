package queryapi

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gitlab.com/sincap/sincap-common/middlewares/qapi"
)

func TestFilter2Sql1Level(t *testing.T) {
	typ, tableName := GetTableName(Sample{})
	q := qapi.Query{Filter: []qapi.Filter{{Name: "InnerF.Name", Operation: qapi.EQ, Value: "Osman"}}}
	where, values, err := filter2Sql(q.Filter, typ, tableName)
	assert.NoError(t, err)
	assert.Equal(t, "Osman", values[0])
	assert.Equal(t, "`Sample`.`InnerFID` IN ( SELECT `Inner1`.`ID`  FROM `Inner1` WHERE ( `Inner1`.`Name` = ? ) )", where)
}

func TestFilter2Sql2Level(t *testing.T) {
	typ, tableName := GetTableName(Sample{})
	q := qapi.Query{Filter: []qapi.Filter{{Name: "InnerF.Inner2F.Name", Operation: qapi.EQ, Value: "Osman"}}}
	where, values, err := filter2Sql(q.Filter, typ, tableName)
	assert.NoError(t, err)
	assert.Equal(t, "Osman", values[0])
	assert.Equal(t, "`Sample`.`InnerFID` IN ( SELECT `Inner1`.`ID`  FROM `Inner1` WHERE ( `Inner1`.`Inner2FID` IN ( SELECT `Inner2`.`ID`  FROM `Inner2` WHERE ( `Inner2`.`Name` = ? ) ) ) )", where)
}

func TestFilter2Sql2LevelUint(t *testing.T) {
	typ, tableName := GetTableName(Sample{})
	q := qapi.Query{Filter: []qapi.Filter{{Name: "InnerF.Inner2F.Age", Operation: qapi.EQ, Value: "18"}}}
	where, values, err := filter2Sql(q.Filter, typ, tableName)
	assert.NoError(t, err)
	assert.Equal(t, uint64(18), values[0])
	assert.Equal(t, "`Sample`.`InnerFID` IN ( SELECT `Inner1`.`ID`  FROM `Inner1` WHERE ( `Inner1`.`Inner2FID` IN ( SELECT `Inner2`.`ID`  FROM `Inner2` WHERE ( `Inner2`.`Age` = ? ) ) ) )", where)
}

func TestFilter2SqlPoly1Level(t *testing.T) {
	typ, tableName := GetTableName(SamplePoly{})
	q := qapi.Query{Filter: []qapi.Filter{{Name: "InnerF.Name", Operation: qapi.EQ, Value: "Osman"}}}
	where, values, err := filter2Sql(q.Filter, typ, tableName)
	assert.NoError(t, err)
	assert.Equal(t, "Osman", values[0])
	assert.Equal(t, "`SamplePoly`.`ID` IN ( SELECT `Inner1`.`HolderID` FROM `Inner1` WHERE ( `Inner1`.`Name` = ? AND `Inner1`.`HolderID` = `SamplePoly`.`ID` AND `Inner1`.`HolderType` = 'SamplePoly' ) )", where)
}

func TestFilter2SqlPoly2Level(t *testing.T) {
	typ, tableName := GetTableName(SamplePoly{})
	q := qapi.Query{Filter: []qapi.Filter{{Name: "InnerF.Inner2P.Name", Operation: qapi.EQ, Value: "Osman"}}}
	where, values, err := filter2Sql(q.Filter, typ, tableName)
	assert.NoError(t, err)
	assert.Equal(t, "Osman", values[0])
	assert.Equal(t, "`SamplePoly`.`ID` IN ( SELECT `Inner1`.`HolderID` FROM `Inner1` WHERE ( `Inner1`.`ID` IN ( SELECT `Inner2`.`HolderID` FROM `Inner2` WHERE ( `Inner2`.`Name` = ? AND `Inner2`.`HolderID` = `Inner1`.`ID` AND `Inner2`.`HolderType` = 'Inner1' ) ) AND `Inner1`.`HolderID` = `SamplePoly`.`ID` AND `Inner1`.`HolderType` = 'SamplePoly' ) )", where)
}
func TestFilter2SqlPM2M(t *testing.T) {
	typ, tableName := GetTableName(SampleM2M{})
	q := qapi.Query{Filter: []qapi.Filter{{Name: "Inner2s.Name", Operation: qapi.EQ, Value: "Osman"}}}
	where, values, err := filter2Sql(q.Filter, typ, tableName)
	assert.NoError(t, err)
	assert.Equal(t, "Osman", values[0])
	assert.Equal(t, "`SampleM2M`.ID IN ( SELECT `SampleM2MID` FROM `SampleM2MInner2` WHERE ( `Inner2ID` IN ( SELECT ID FROM `Inner2` WHERE ( `Name` = ? ) ) ) )", where)
}
