package util

import (
	"reflect"

	"github.com/filllabs/sincap-common/reflection"
)

// GetMigrationTableNames extracts table names from model structs for migration purposes
func GetMigrationTableNames(models ...interface{}) []string {
	var tables []string

	for _, model := range models {
		typ := reflection.ExtractRealTypeField(reflect.TypeOf(model))

		// Add main table name
		if m, hasName := typ.MethodByName("TableName"); hasName {
			res := m.Func.Call([]reflect.Value{reflect.ValueOf(model)})
			tables = append(tables, res[0].String())
		} else {
			tables = append(tables, typ.Name())
		}
	}

	return tables
}
