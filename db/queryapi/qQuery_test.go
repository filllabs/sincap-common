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
	assert.Equal(t, "`Sample`.`Name` LIKE ? OR `Sample`.`InnerFID` IN ( SELECT  `Inner1`.`ID`  FROM `Inner1` WHERE ( `Inner1`.`Name` LIKE ? OR `Inner1`.`Inner2PID` IN ( SELECT  `Inner2`.`ID`  FROM `Inner2` WHERE ( `Inner2`.`Name` LIKE ? ) ) ) )", where)
}
func TestQ2SqlPoly(t *testing.T) {
	typ, tableName := GetTableName(SamplePoly{})
	q := qapi.Query{Q: "seray"}
	where, values, err := q2Sql(q.Q, typ, tableName)
	assert.NoError(t, err)
	assert.Equal(t, "seray", values[0])
	assert.Equal(t, "%seray", values[1])
	assert.Equal(t, "seray%", values[2])
	assert.Equal(t, "`SamplePoly`.`Name` LIKE ? OR `SamplePoly`.`InnerFID` IN ( SELECT  `Inner1`.`ID`  FROM `Inner1` WHERE ( `Inner1`.`Name` LIKE ? OR `Inner1`.`Inner2PID` IN ( SELECT  `Inner2`.`ID`  FROM `Inner2` WHERE ( `Inner2`.`Name` LIKE ? ) ) ) )", where)
}
func TestQ2SqlM2m(t *testing.T) {
	typ, tableName := GetTableName(SampleM2M{})
	q := qapi.Query{Q: "seray"}
	where, _, err := q2Sql(q.Q, typ, tableName)
	assert.NoError(t, err)
	assert.Equal(t, "`SampleM2M`.`Inner2sID` IN ( SELECT  `Inner2`.`ID`  FROM `Inner2` WHERE ( `Inner2`.`Name` LIKE ? ) )", where)
}

func TestQ2SqlWithJoins(t *testing.T) {
	registry := NewJoinRegistry()
	registry.Register("InnerF", JoinConfig{
		Type:       OneToOne,
		Table:      "Inner1",
		LocalKey:   "ID",
		ForeignKey: "SampleID",
	})

	typ, tableName := GetTableName(Sample{})
	q := qapi.Query{Q: "seray"}

	options := &QueryOptions{JoinRegistry: registry}
	where, values, relPaths, err := q2SqlWithJoins(q.Q, typ, tableName, options)

	assert.NoError(t, err)
	assert.Equal(t, "%seray%", values[0])
	assert.Contains(t, relPaths, "InnerF")
	assert.Contains(t, where, "`Sample`.`Name` LIKE ?")
}
