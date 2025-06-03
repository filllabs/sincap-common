package mysql

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/filllabs/sincap-common/db/queryapi"
	"github.com/filllabs/sincap-common/db/types"
	"github.com/filllabs/sincap-common/logging"
	"github.com/filllabs/sincap-common/middlewares/qapi"

	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

// List calls ListByQuery or ListAll according to the query parameter
func List(DB *sqlx.DB, records any, query *qapi.Query) (int, error) {
	value := reflect.ValueOf(records)
	if value.Kind() != reflect.Pointer {
		return 0, fmt.Errorf("records must be a pointer")
	}

	elem := value.Elem()
	if elem.Kind() != reflect.Slice {
		return 0, fmt.Errorf("records must be a pointer to slice")
	}

	// Generate SQL query
	queryResult, err := queryapi.GenerateSQL(query, records)
	if err != nil {
		return 0, err
	}

	// Get count if pagination is needed
	var count int64 = -1
	calculateCount := query.Offset > 0 || query.Limit > 0

	if calculateCount {
		err = DB.Get(&count, queryResult.CountQuery, queryResult.CountArgs...)
		if err != nil {
			return 0, err
		}
	}

	// Execute main query
	err = DB.Select(records, queryResult.Query, queryResult.Args...)
	if err != nil {
		return 0, err
	}

	if !calculateCount {
		return elem.Len(), nil
	}
	return int(count), nil
}

// Create Record
func Create(DB *sqlx.DB, record any) error {
	// Get table name and struct info
	typ, tableName := queryapi.GetTableName(record)

	// Build INSERT query using reflection
	var columns []string
	var placeholders []string
	var values []interface{}

	recordValue := reflect.ValueOf(record)
	if recordValue.Kind() == reflect.Ptr {
		recordValue = recordValue.Elem()
	}

	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		fieldValue := recordValue.Field(i)

		// Skip ID field (assuming auto-increment)
		if strings.ToLower(field.Name) == "id" {
			continue
		}

		// Skip unexported fields
		if !fieldValue.CanInterface() {
			continue
		}

		columns = append(columns, safeMySQLNaming(field.Name))
		placeholders = append(placeholders, "?")
		values = append(values, fieldValue.Interface())
	}

	query := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)",
		safeMySQLNaming(tableName),
		strings.Join(columns, ", "),
		strings.Join(placeholders, ", "))

	result, err := DB.Exec(query, values...)
	if err != nil {
		logging.Logger.Error("Create error", zap.Any("Model", reflect.TypeOf(record)), zap.Error(err), zap.Any("record", record))
		return err
	}

	// Set the ID if it's an auto-increment field
	if id, err := result.LastInsertId(); err == nil {
		idField := recordValue.FieldByName("ID")
		if idField.IsValid() && idField.CanSet() {
			idField.SetUint(uint64(id))
		}
	}

	return nil
}

// Read Record
func Read(DB *sqlx.DB, record any, id any, preloads ...string) error {
	// Note: preloads are ignored in sqlx implementation
	// Users need to handle relationships manually

	_, tableName := queryapi.GetTableName(record)

	// Build SELECT query
	query := fmt.Sprintf("SELECT * FROM %s WHERE id = ?", safeMySQLNaming(tableName))

	err := DB.Get(record, query, id)
	if err != nil {
		logging.Logger.Error("Read error", zap.Any("Model", reflect.TypeOf(record)), zap.Error(err), zap.Any("id", id))
	}
	return err
}

// Update Updates the record with the given fields
func Update(DB *sqlx.DB, model any, fieldsParams ...map[string]any) error {
	typ, tableName := queryapi.GetTableName(model)

	if len(fieldsParams) == 0 {
		// Update full record - build SET clause from struct fields
		var setClauses []string
		var values []interface{}
		var whereValue interface{}

		modelValue := reflect.ValueOf(model)
		if modelValue.Kind() == reflect.Ptr {
			modelValue = modelValue.Elem()
		}

		for i := 0; i < typ.NumField(); i++ {
			field := typ.Field(i)
			fieldValue := modelValue.Field(i)

			if !fieldValue.CanInterface() {
				continue
			}

			if strings.ToLower(field.Name) == "id" {
				whereValue = fieldValue.Interface()
				continue
			}

			setClauses = append(setClauses, fmt.Sprintf("%s = ?", safeMySQLNaming(field.Name)))
			values = append(values, fieldValue.Interface())
		}

		if whereValue == nil {
			return fmt.Errorf("no ID field found for update")
		}

		query := fmt.Sprintf("UPDATE %s SET %s WHERE id = ?",
			safeMySQLNaming(tableName),
			strings.Join(setClauses, ", "))

		values = append(values, whereValue)

		_, err := DB.Exec(query, values...)
		if err != nil {
			logging.Logger.Error("Update error", zap.Any("Model", reflect.TypeOf(model)), zap.Error(err), zap.Any("record", model))
		}
		return err
	}

	// Partial update with specific fields
	if len(fieldsParams) > 1 || fieldsParams[0] == nil {
		return fmt.Errorf("update failed: invalid fields parameter")
	}

	fields := fieldsParams[0]

	// Handle JSON fields
	for k, v := range fields {
		switch v.(type) {
		case map[string]any, []any:
			j := types.JSON{}
			j.Marshal(v)
			fields[k] = j
		}
	}

	// Get ID from model
	modelValue := reflect.ValueOf(model)
	if modelValue.Kind() == reflect.Ptr {
		modelValue = modelValue.Elem()
	}

	idField := modelValue.FieldByName("ID")
	if !idField.IsValid() {
		return fmt.Errorf("no ID field found for update")
	}

	var setClauses []string
	var values []interface{}

	for k, v := range fields {
		setClauses = append(setClauses, fmt.Sprintf("%s = ?", safeMySQLNaming(k)))
		values = append(values, v)
	}

	query := fmt.Sprintf("UPDATE %s SET %s WHERE id = ?",
		safeMySQLNaming(tableName),
		strings.Join(setClauses, ", "))

	values = append(values, idField.Interface())

	_, err := DB.Exec(query, values...)
	if err != nil {
		logging.Logger.Error("Update error", zap.Any("Model", reflect.TypeOf(model)), zap.Error(err), zap.Any("record", fields))
	}
	return err
}

// Delete Record
func Delete(DB *sqlx.DB, record any) error {
	_, tableName := queryapi.GetTableName(record)

	// Get ID from record
	recordValue := reflect.ValueOf(record)
	if recordValue.Kind() == reflect.Ptr {
		recordValue = recordValue.Elem()
	}

	idField := recordValue.FieldByName("ID")
	if !idField.IsValid() {
		return fmt.Errorf("no ID field found for delete")
	}

	query := fmt.Sprintf("DELETE FROM %s WHERE id = ?", safeMySQLNaming(tableName))

	_, err := DB.Exec(query, idField.Interface())
	if err != nil {
		logging.Logger.Error("Delete error", zap.Any("Model", reflect.TypeOf(record)), zap.Error(err), zap.Any("record", record))
	}
	return err
}

// DeleteAll Records
func DeleteAll(DB *sqlx.DB, record any, ids ...any) error {
	_, tableName := queryapi.GetTableName(record)

	if len(ids) == 0 {
		return fmt.Errorf("no IDs provided for bulk delete")
	}

	// Build IN clause
	placeholders := strings.Repeat("?,", len(ids))
	placeholders = placeholders[:len(placeholders)-1] // Remove trailing comma

	query := fmt.Sprintf("DELETE FROM %s WHERE id IN (%s)",
		safeMySQLNaming(tableName),
		placeholders)

	_, err := DB.Exec(query, ids...)
	if err != nil {
		logging.Logger.Error("Delete error", zap.Any("Model", reflect.TypeOf(record)), zap.Error(err), zap.Any("ids", ids))
	}
	return err
}

// safeMySQLNaming wraps column/table names with backticks for MySQL
func safeMySQLNaming(data string) string {
	return "`" + data + "`"
}
