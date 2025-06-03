package util

import (
	"fmt"
	"reflect"

	"github.com/filllabs/sincap-common/db"
	"github.com/filllabs/sincap-common/logging"
	"github.com/filllabs/sincap-common/reflection"
	"github.com/jmoiron/sqlx"
)

// DropAll drops all tables at the database
// Deprecated: Use golang-migrate or goose for database migrations instead
func DropAll(DB *sqlx.DB, models ...interface{}) error {
	logging.Logger.Warn("DropAll is deprecated. Use golang-migrate or goose for database migrations instead.")
	return fmt.Errorf("DropAll is deprecated. Use golang-migrate or goose for database migrations instead")
}

// DropRelationTables drops all relation tables at the database
// Deprecated: Use golang-migrate or goose for database migrations instead
func DropRelationTables(DB *sqlx.DB, models ...interface{}) error {
	logging.Logger.Warn("DropRelationTables is deprecated. Use golang-migrate or goose for database migrations instead.")
	return fmt.Errorf("DropRelationTables is deprecated. Use golang-migrate or goose for database migrations instead")
}

// AutoMigrate migrates all models defined
// Deprecated: Use golang-migrate or goose for database migrations instead
func AutoMigrate(command string, dbconfig db.Config, DB *sqlx.DB, models ...interface{}) error {
	logging.Logger.Warn("AutoMigrate is deprecated. Use golang-migrate or goose for database migrations instead.")
	return fmt.Errorf("AutoMigrate is deprecated. Use golang-migrate or goose for database migrations instead")
}

// GetMigrationTableNames extracts table names from models for migration purposes
// This can be used to help generate migration files manually
func GetMigrationTableNames(models ...interface{}) []string {
	var tables []string
	var relationTables []string

	for _, model := range models {
		typ := reflection.ExtractRealTypeField(reflect.TypeOf(model))

		// Add main table name
		if m, hasName := typ.MethodByName("TableName"); hasName {
			res := m.Func.Call([]reflect.Value{reflect.ValueOf(model)})
			tables = append(tables, res[0].String())
		} else {
			tables = append(tables, typ.Name())
		}

		// Add relation tables
		for i := 0; i < typ.NumField(); i++ {
			f := typ.Field(i)
			if name, isM2M := GetMany2Many(&f); isM2M {
				relationTables = append(relationTables, name)
			}
		}
	}

	// Combine main tables and relation tables
	allTables := append(tables, relationTables...)
	return allTables
}
