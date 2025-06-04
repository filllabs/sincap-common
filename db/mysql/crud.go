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

// TableNamer interface for models that can provide their table name
type TableNamer interface {
	TableName() string
}

// IDGetter interface for models that can provide their ID
type IDGetter interface {
	GetID() interface{}
}

// IDSetter interface for models that can set their ID
type IDSetter interface {
	SetID(id interface{}) error
}

// FieldMapper interface for models that can provide their field mappings
type FieldMapper interface {
	GetFieldMap() map[string]interface{}
}

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

// CreateWithFieldMap creates a record using a field map (no reflection)
func CreateWithFieldMap(DB *sqlx.DB, tableName string, fieldMap map[string]interface{}, idSetter IDSetter) error {
	if len(fieldMap) == 0 {
		return fmt.Errorf("no fields provided for insert")
	}

	var columns []string
	var placeholders []string
	var values []interface{}

	for column, value := range fieldMap {
		// Skip ID field (assuming auto-increment)
		if strings.ToLower(column) == "id" {
			continue
		}
		columns = append(columns, safeMySQLNaming(column))
		placeholders = append(placeholders, "?")
		values = append(values, value)
	}

	if len(columns) == 0 {
		return fmt.Errorf("no valid fields for insert")
	}

	query := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)",
		safeMySQLNaming(tableName),
		strings.Join(columns, ", "),
		strings.Join(placeholders, ", "))

	result, err := DB.Exec(query, values...)
	if err != nil {
		logging.Logger.Error("CreateWithFieldMap error", zap.String("table", tableName), zap.Error(err))
		return err
	}

	// Set the ID if possible
	if idSetter != nil {
		if id, err := result.LastInsertId(); err == nil {
			idSetter.SetID(uint64(id))
		}
	}

	return nil
}

// Create Record (with reduced reflection)
func Create(DB *sqlx.DB, record any) error {
	// Try interface-based approach first
	if fieldMapper, ok := record.(FieldMapper); ok {
		var tableName string
		if tableNamer, ok := record.(TableNamer); ok {
			tableName = tableNamer.TableName()
		} else {
			_, tableName = queryapi.GetTableName(record)
		}

		var idSetter IDSetter
		if setter, ok := record.(IDSetter); ok {
			idSetter = setter
		}

		return CreateWithFieldMap(DB, tableName, fieldMapper.GetFieldMap(), idSetter)
	}

	// Fallback to reflection-based approach
	return createWithReflection(DB, record)
}

// createWithReflection is the original reflection-based implementation
func createWithReflection(DB *sqlx.DB, record any) error {
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

// ReadByID reads a record by ID (no reflection needed)
func ReadByID(DB *sqlx.DB, dest any, tableName string, id any) error {
	query := fmt.Sprintf("SELECT * FROM %s WHERE ID = ?", safeMySQLNaming(tableName))
	err := DB.Get(dest, query, id)
	if err != nil {
		logging.Logger.Error("ReadByID error", zap.String("table", tableName), zap.Error(err), zap.Any("id", id))
	}
	return err
}

// Read Record (with reduced reflection)
func Read(DB *sqlx.DB, record any, id any, preloads ...string) error {
	// Note: preloads are ignored in sqlx implementation
	// Users need to handle relationships manually

	var tableName string
	if tableNamer, ok := record.(TableNamer); ok {
		tableName = tableNamer.TableName()
	} else {
		_, tableName = queryapi.GetTableName(record)
	}

	return ReadByID(DB, record, tableName, id)
}

// UpdateWithFieldMap updates a record using a field map (no reflection)
func UpdateWithFieldMap(DB *sqlx.DB, tableName string, id interface{}, fieldMap map[string]interface{}) error {
	if len(fieldMap) == 0 {
		return fmt.Errorf("no fields provided for update")
	}

	var setClauses []string
	var values []interface{}

	// Handle JSON fields
	for k, v := range fieldMap {
		switch v.(type) {
		case map[string]any, []any:
			j := types.JSON{}
			j.Marshal(v)
			fieldMap[k] = j
		}
	}

	for k, v := range fieldMap {
		setClauses = append(setClauses, fmt.Sprintf("%s = ?", safeMySQLNaming(k)))
		values = append(values, v)
	}

	query := fmt.Sprintf("UPDATE %s SET %s WHERE ID = ?",
		safeMySQLNaming(tableName),
		strings.Join(setClauses, ", "))

	values = append(values, id)

	_, err := DB.Exec(query, values...)
	if err != nil {
		logging.Logger.Error("UpdateWithFieldMap error", zap.String("table", tableName), zap.Error(err))
	}
	return err
}

// Update Updates the record with the given fields (with reduced reflection)
func Update(DB *sqlx.DB, model any, fieldsParams ...map[string]any) error {
	var tableName string
	if tableNamer, ok := model.(TableNamer); ok {
		tableName = tableNamer.TableName()
	} else {
		_, tableName = queryapi.GetTableName(model)
	}

	// Partial update with specific fields
	if len(fieldsParams) > 0 && fieldsParams[0] != nil {
		var id interface{}
		if idGetter, ok := model.(IDGetter); ok {
			id = idGetter.GetID()
		} else {
			// Fallback to reflection
			modelValue := reflect.ValueOf(model)
			if modelValue.Kind() == reflect.Ptr {
				modelValue = modelValue.Elem()
			}
			idField := modelValue.FieldByName("ID")
			if !idField.IsValid() {
				return fmt.Errorf("no ID field found for update")
			}
			id = idField.Interface()
		}

		return UpdateWithFieldMap(DB, tableName, id, fieldsParams[0])
	}

	// Full record update - try interface-based approach first
	if fieldMapper, ok := model.(FieldMapper); ok {
		var id interface{}
		if idGetter, ok := model.(IDGetter); ok {
			id = idGetter.GetID()
		} else {
			return fmt.Errorf("model must implement IDGetter for full update")
		}

		fieldMap := fieldMapper.GetFieldMap()
		// Remove ID from field map for update
		delete(fieldMap, "ID")
		delete(fieldMap, "id")

		return UpdateWithFieldMap(DB, tableName, id, fieldMap)
	}

	// Fallback to reflection-based approach
	return updateWithReflection(DB, model, tableName)
}

// updateWithReflection is the original reflection-based implementation
func updateWithReflection(DB *sqlx.DB, model any, tableName string) error {
	typ, _ := queryapi.GetTableName(model)

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

	query := fmt.Sprintf("UPDATE %s SET %s WHERE ID = ?",
		safeMySQLNaming(tableName),
		strings.Join(setClauses, ", "))

	values = append(values, whereValue)

	_, err := DB.Exec(query, values...)
	if err != nil {
		logging.Logger.Error("Update error", zap.Any("Model", reflect.TypeOf(model)), zap.Error(err), zap.Any("record", model))
	}
	return err
}

// DeleteByID deletes a record by ID (no reflection needed)
func DeleteByID(DB *sqlx.DB, tableName string, id interface{}) error {
	query := fmt.Sprintf("DELETE FROM %s WHERE ID = ?", safeMySQLNaming(tableName))
	_, err := DB.Exec(query, id)
	if err != nil {
		logging.Logger.Error("DeleteByID error", zap.String("table", tableName), zap.Error(err), zap.Any("id", id))
	}
	return err
}

// Delete Record (with reduced reflection)
func Delete(DB *sqlx.DB, record any) error {
	var tableName string
	if tableNamer, ok := record.(TableNamer); ok {
		tableName = tableNamer.TableName()
	} else {
		_, tableName = queryapi.GetTableName(record)
	}

	var id interface{}
	if idGetter, ok := record.(IDGetter); ok {
		id = idGetter.GetID()
	} else {
		// Fallback to reflection
		recordValue := reflect.ValueOf(record)
		if recordValue.Kind() == reflect.Ptr {
			recordValue = recordValue.Elem()
		}

		idField := recordValue.FieldByName("ID")
		if !idField.IsValid() {
			return fmt.Errorf("no ID field found for delete")
		}
		id = idField.Interface()
	}

	return DeleteByID(DB, tableName, id)
}

// DeleteAll Records (no reflection needed)
func DeleteAll(DB *sqlx.DB, record any, ids ...any) error {
	var tableName string
	if tableNamer, ok := record.(TableNamer); ok {
		tableName = tableNamer.TableName()
	} else {
		_, tableName = queryapi.GetTableName(record)
	}

	if len(ids) == 0 {
		return fmt.Errorf("no IDs provided for bulk delete")
	}

	// Build IN clause
	placeholders := strings.Repeat("?,", len(ids))
	placeholders = placeholders[:len(placeholders)-1] // Remove trailing comma

	query := fmt.Sprintf("DELETE FROM %s WHERE ID IN (%s)",
		safeMySQLNaming(tableName),
		placeholders)

	_, err := DB.Exec(query, ids...)
	if err != nil {
		logging.Logger.Error("DeleteAll error", zap.String("table", tableName), zap.Error(err), zap.Any("ids", ids))
	}
	return err
}

// safeMySQLNaming wraps column/table names with backticks for MySQL
func safeMySQLNaming(data string) string {
	return "`" + data + "`"
}
