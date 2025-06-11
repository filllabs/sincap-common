package queryapi

import (
	"testing"

	"github.com/filllabs/sincap-common/middlewares/qapi"
	"github.com/stretchr/testify/assert"
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

// Updated test for polymorphic relationships - now falls back to simple foreign key relationship
func TestFilter2SqlPoly1Level(t *testing.T) {
	typ, tableName := GetTableName(SamplePoly{})
	q := qapi.Query{Filter: []qapi.Filter{{Name: "InnerF.Name", Operation: qapi.EQ, Value: "Osman"}}}
	where, values, err := filter2Sql(q.Filter, typ, tableName)
	assert.NoError(t, err)
	assert.Equal(t, "Osman", values[0])
	// Without GORM tags, this now falls back to simple foreign key relationship
	assert.Equal(t, "`SamplePoly`.`InnerFID` IN ( SELECT `Inner1`.`ID`  FROM `Inner1` WHERE ( `Inner1`.`Name` = ? ) )", where)
}

// Updated test for nested polymorphic relationships - now falls back to simple foreign key relationships
func TestFilter2SqlPoly2Level(t *testing.T) {
	typ, tableName := GetTableName(SamplePoly{})
	q := qapi.Query{Filter: []qapi.Filter{{Name: "InnerF.Inner2P.Name", Operation: qapi.EQ, Value: "Osman"}}}
	where, values, err := filter2Sql(q.Filter, typ, tableName)
	assert.NoError(t, err)
	assert.Equal(t, "Osman", values[0])
	// Without GORM tags, this now falls back to simple foreign key relationships
	assert.Equal(t, "`SamplePoly`.`InnerFID` IN ( SELECT `Inner1`.`ID`  FROM `Inner1` WHERE ( `Inner1`.`Inner2PID` IN ( SELECT `Inner2`.`ID`  FROM `Inner2` WHERE ( `Inner2`.`Name` = ? ) ) ) )", where)
}

// Updated test for many-to-many relationships - now falls back to simple foreign key relationship
func TestFilter2SqlPM2M(t *testing.T) {
	typ, tableName := GetTableName(SampleM2M{})
	q := qapi.Query{Filter: []qapi.Filter{{Name: "Inner2s.Name", Operation: qapi.EQ, Value: "Osman"}}}
	where, values, err := filter2Sql(q.Filter, typ, tableName)
	assert.NoError(t, err)
	assert.Equal(t, "Osman", values[0])
	// Without GORM tags, this now falls back to simple foreign key relationship
	assert.Equal(t, "`SampleM2M`.`Inner2sID` IN ( SELECT `Inner2`.`ID`  FROM `Inner2` WHERE ( `Inner2`.`Name` = ? ) )", where)
}

// New test demonstrating the join-based approach
func TestFilter2SqlWithJoins(t *testing.T) {
	// Set up join registry for proper relationship handling
	registry := NewJoinRegistry()
	registry.Register("InnerF", JoinConfig{
		Type:       OneToOne,
		Table:      "Inner1",
		LocalKey:   "ID",
		ForeignKey: "SampleID",
	})

	typ, tableName := GetTableName(Sample{})
	q := qapi.Query{Filter: []qapi.Filter{{Name: "InnerF.Name", Operation: qapi.EQ, Value: "Osman"}}}

	options := &QueryOptions{JoinRegistry: registry}
	where, values, relPaths, err := filter2SqlWithJoins(q.Filter, typ, tableName, options)

	assert.NoError(t, err)
	assert.Equal(t, "Osman", values[0])
	assert.Contains(t, relPaths, "InnerF") // Should detect the relationship path
	// With joins configured, it should generate a simpler condition
	assert.Equal(t, "`Inner1`.`Name` = ?", where)
}
