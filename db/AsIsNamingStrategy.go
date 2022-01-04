package db

import (
	"crypto/sha1"
	"fmt"
	"strings"
	"unicode/utf8"

	"gorm.io/gorm/schema"
)

// AsIsNamingStrategy creates a namingstrategy which returns desired names as is without any modification
// like case or plural conversion
func AsIsNamingStrategy() schema.Namer {
	n := asIsNamer{}
	return &n
}

type asIsNamer struct {
}

// TableName convert string to table name
func (n *asIsNamer) TableName(table string) string { return table }

// ColumnName convert string to column name
func (n *asIsNamer) ColumnName(table, column string) string { return column }

// JoinTableName convert string to join table name
func (n *asIsNamer) JoinTableName(joinTable string) string { return joinTable }

// RelationshipFKName generate fk name for relation
func (n *asIsNamer) RelationshipFKName(rel schema.Relationship) string {
	return strings.Replace(fmt.Sprintf("fk_%s_%s", rel.Schema.Table, rel.Name), ".", "_", -1)
}

// CheckerName generate checker name
func (n *asIsNamer) CheckerName(table, column string) string {
	return strings.Replace(fmt.Sprintf("chk_%s_%s", table, column), ".", "_", -1)
}

// IndexName generate index name
func (n *asIsNamer) IndexName(table, column string) string {
	idxName := fmt.Sprintf("idx_%v_%v", table, column)
	idxName = strings.Replace(idxName, ".", "_", -1)

	if utf8.RuneCountInString(idxName) > 64 {
		h := sha1.New()
		h.Write([]byte(idxName))
		bs := h.Sum(nil)

		idxName = fmt.Sprintf("idx%v%v", table, column)[0:56] + string(bs)[:8]
	}
	return idxName
}
func (n *asIsNamer) SchemaName(table string) string {
	return strings.Split(".", table)[0]
}
