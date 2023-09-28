package queryapi

import (
	"testing"

	"github.com/filllabs/sincap-common/middlewares/qapi"
	"github.com/stretchr/testify/assert"
)

func TestQ2Sql(t *testing.T) {
	typ, tableName := GetTableName(Sample{})
	q := qapi.Query{Q: "seray"}
	where, values, err := q2Sql(q.Q, typ, tableName)
	assert.NoError(t, err)
	assert.Equal(t, "%seray%", values[0])
	assert.Equal(t, "%seray", values[1])
	assert.Equal(t, "seray%", values[2])
	assert.Equal(t, "`Sample`.`Name` LIKE ? OR `Sample`.`InnerFID` IN ( SELECT  `Inner1`.`ID`  FROM `Inner1` WHERE ( `Inner1`.`Name` LIKE ? OR `Inner1`.`ID` IN ( SELECT `Inner2`.`HolderID` FROM `Inner2` WHERE ( `Inner2`.`Name` LIKE ? ) ) ) )", where)
}
func TestQ2SqlPoly(t *testing.T) {
	typ, tableName := GetTableName(SamplePoly{})
	q := qapi.Query{Q: "seray"}
	where, values, err := q2Sql(q.Q, typ, tableName)
	assert.NoError(t, err)
	assert.Equal(t, "seray", values[0])
	assert.Equal(t, "%seray", values[1])
	assert.Equal(t, "seray%", values[2])
	assert.Equal(t, "`SamplePoly`.`Name` LIKE ? OR `SamplePoly`.`ID` IN ( SELECT `Inner1`.`HolderID` FROM `Inner1` WHERE ( `Inner1`.`Name` LIKE ? OR `Inner1`.`ID` IN ( SELECT `Inner2`.`HolderID` FROM `Inner2` WHERE ( `Inner2`.`Name` LIKE ? ) ) ) )", where)
}
func TestQ2SqlM2m(t *testing.T) {
	typ, tableName := GetTableName(SampleM2M{})
	q := qapi.Query{Q: "seray"}
	where, _, err := q2Sql(q.Q, typ, tableName)
	assert.NoError(t, err)
	// assert.Equal(t, "%seray%", values[0])
	// assert.Equal(t, "%seray", values[1])
	// assert.Equal(t, "seray%", values[2])
	assert.Equal(t, "`SampleM2M`.`ID` IN ( SELECT `SampleM2MInner2`.`SampleM2MID` FROM `SampleM2MInner2` WHERE ( `SampleM2MInner2`.`Inner2ID` IN ( SELECT  `Inner2`.`ID`  FROM `Inner2` WHERE ( `Inner2`.`Name` LIKE ? ) ) ) )", where)
}
