package mysql

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/filllabs/sincap-common/db/interfaces"
	"github.com/filllabs/sincap-common/db/mysql/translations"
	"github.com/filllabs/sincap-common/db/queryapi"
	"github.com/filllabs/sincap-common/db/types"
	"github.com/filllabs/sincap-common/db/util"
	"github.com/filllabs/sincap-common/logging"
	"github.com/filllabs/sincap-common/middlewares/qapi"
	"github.com/gofiber/fiber/v2"

	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

// Helper functions to eliminate code duplication

// getTableName gets table name using optimized interface approach with fallback
func getTableName(record any) string {
	if tableNamer, ok := record.(interfaces.TableNamer); ok {
		return tableNamer.TableName()
	}
	_, tableName := queryapi.GetTableName(record)
	return tableName
}

// getRecordID gets record ID using optimized interface approach with fallback
func getRecordID(record any) (interface{}, error) {
	if idGetter, ok := record.(interfaces.IDGetter); ok {
		return idGetter.GetID(), nil
	}

	// Fallback to reflection
	recordValue := reflect.ValueOf(record)
	if recordValue.Kind() == reflect.Ptr {
		recordValue = recordValue.Elem()
	}
	idField := recordValue.FieldByName("ID")
	if !idField.IsValid() {
		return nil, fmt.Errorf("no ID field found")
	}
	return idField.Interface(), nil
}

// setRecordID sets record ID using optimized interface approach with fallback
func setRecordID(record any, id uint64) {
	if idSetter, ok := record.(interfaces.IDSetter); ok {
		idSetter.SetID(id)
		return
	}

	// Fallback to reflection
	recordValue := reflect.ValueOf(record)
	if recordValue.Kind() == reflect.Ptr {
		recordValue = recordValue.Elem()
	}
	idField := recordValue.FieldByName("ID")
	if idField.IsValid() && idField.CanSet() {
		idField.SetUint(id)
	}
}

// processFieldValue handles JSON conversion and driver.Valuer interface
func processFieldValue(value interface{}) interface{} {
	switch v := value.(type) {
	case map[string]any, []any:
		j := types.JSON{}
		j.Marshal(v)
		return j
	default:
		// Check if the value implements driver.Valuer
		if valuer, ok := value.(driver.Valuer); ok {
			if dbValue, err := valuer.Value(); err == nil && dbValue != nil {
				return dbValue
			}
		}
		return value
	}
}

// processFieldValueWithReflection handles JSON conversion and driver.Valuer interface with reflection fallback
func processFieldValueWithReflection(value interface{}, fieldValue reflect.Value) interface{} {
	switch v := value.(type) {
	case map[string]any, []any:
		j := types.JSON{}
		j.Marshal(v)
		return j
	default:
		// Check if the value implements driver.Valuer
		if valuer, ok := value.(driver.Valuer); ok {
			if dbValue, err := valuer.Value(); err == nil && dbValue != nil {
				return dbValue
			}
		} else if fieldValue.CanAddr() {
			// Check if the pointer to the value implements driver.Valuer
			if valuer, ok := fieldValue.Addr().Interface().(driver.Valuer); ok {
				if dbValue, err := valuer.Value(); err == nil && dbValue != nil {
					return dbValue
				}
			}
		}
		return value
	}
}

// setTimestampFields sets CreatedAt and UpdatedAt fields in a field map
func setTimestampFields(fieldMap map[string]interface{}, recordType reflect.Type, isCreate bool) {
	now := time.Now()

	if isCreate {
		// Set CreatedAt if not provided and field exists
		if _, exists := fieldMap["CreatedAt"]; !exists {
			if field, found := recordType.FieldByName("CreatedAt"); found {
				fieldType := field.Type
				if fieldType == reflect.TypeOf(time.Time{}) || fieldType == reflect.TypeOf(&time.Time{}) {
					fieldMap["CreatedAt"] = now
				}
			}
		}
	}

	// Set UpdatedAt if not provided and field exists
	if _, exists := fieldMap["UpdatedAt"]; !exists {
		if field, found := recordType.FieldByName("UpdatedAt"); found {
			fieldType := field.Type
			if fieldType == reflect.TypeOf(time.Time{}) || fieldType == reflect.TypeOf(&time.Time{}) {
				fieldMap["UpdatedAt"] = now
			}
		}
	}
}

// setTimestampFieldsReflection sets CreatedAt and UpdatedAt fields using reflection
func setTimestampFieldsReflection(recordValue reflect.Value, isCreate bool) {
	now := time.Now()

	if isCreate {
		// Set CreatedAt if field exists and is zero value
		if createdAtField := recordValue.FieldByName("CreatedAt"); createdAtField.IsValid() && createdAtField.CanSet() {
			if createdAtField.Type() == reflect.TypeOf(time.Time{}) && createdAtField.Interface().(time.Time).IsZero() {
				createdAtField.Set(reflect.ValueOf(now))
			} else if createdAtField.Type() == reflect.TypeOf(&time.Time{}) && createdAtField.IsNil() {
				createdAtField.Set(reflect.ValueOf(&now))
			}
		}
	}

	// Set UpdatedAt if field exists
	if updatedAtField := recordValue.FieldByName("UpdatedAt"); updatedAtField.IsValid() && updatedAtField.CanSet() {
		if updatedAtField.Type() == reflect.TypeOf(time.Time{}) {
			updatedAtField.Set(reflect.ValueOf(now))
		} else if updatedAtField.Type() == reflect.TypeOf(&time.Time{}) {
			updatedAtField.Set(reflect.ValueOf(&now))
		}
	}
}

func List(DB *sqlx.DB, records any, query *qapi.Query, lang ...string) (int, error) {
	value := reflect.ValueOf(records)
	if value.Kind() != reflect.Pointer {
		return 0, fmt.Errorf("records must be a pointer")
	}

	elem := value.Elem()
	if elem.Kind() != reflect.Slice {
		return 0, fmt.Errorf("records must be a pointer to slice")
	}

	// Check if we need to handle translations
	langCode := ""
	if len(lang) > 0 {
		langCode = lang[0]
	}

	useTranslations := len(langCode) > 0 && langCode != ""
	multiLangFields := findTranslationFields(records)

	// Generate SQL query with or without translation support
	var queryResult *queryapi.QueryResult
	var err error

	if useTranslations && len(multiLangFields) > 0 {
		// Use custom query generation that handles translation fields properly
		queryResult, err = generateTranslationAwareQuery(query, records, langCode, multiLangFields)
	} else {
		// Use standard query generation
		queryResult, err = queryapi.GenerateDB(query, records)
	}

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

// generateTranslationAwareQuery creates a custom query that handles translation fields without type conversion issues
func generateTranslationAwareQuery(query *qapi.Query, records any, langCode string, multiLangFields []string) (*queryapi.QueryResult, error) {
	// Separate translation and non-translation filters
	var translationFilters []qapi.Filter
	var regularFilters []qapi.Filter

	for _, filter := range query.Filter {
		isTranslationField := false
		for _, field := range multiLangFields {
			if filter.Name == field {
				translationFilters = append(translationFilters, filter)
				isTranslationField = true
				break
			}
		}
		if !isTranslationField {
			regularFilters = append(regularFilters, filter)
		}
	}

	// Create a modified query with only regular filters for standard processing
	regularQuery := &qapi.Query{
		Limit:    query.Limit,
		Offset:   query.Offset,
		Sort:     createTranslatedSortClauses(query.Sort, langCode, multiLangFields),
		Filter:   regularFilters, // Only non-translation filters
		Q:        query.Q,
		Fields:   createTranslatedFields(query.Fields, langCode, multiLangFields, records),
		Preloads: query.Preloads,
	}

	// Generate SQL for regular filters
	queryResult, err := queryapi.GenerateDB(regularQuery, records)
	if err != nil {
		return nil, err
	}

	// Add translation filter conditions manually
	if len(translationFilters) > 0 {
		translationConditions := make([]string, len(translationFilters))
		for i, filter := range translationFilters {
			if langCode == "all" {
				// For "all", search in the entire JSON field
				translationConditions[i] = createTranslationCondition(filter, filter.Name)
			} else {
				// For specific language, search in that language's value
				translationConditions[i] = createTranslationConditionForLanguage(filter, filter.Name, langCode)
			}
		}

		// Combine with existing WHERE conditions
		translationWhere := strings.Join(translationConditions, " AND ")

		// Helper function to add WHERE clause to a query
		addWhereClause := func(query string) string {
			if strings.Contains(query, "WHERE") {
				return strings.Replace(query, "WHERE", "WHERE "+translationWhere+" AND", 1)
			} else {
				// Add WHERE clause if none exists
				if strings.Contains(query, "ORDER BY") {
					return strings.Replace(query, "ORDER BY", "WHERE "+translationWhere+" ORDER BY", 1)
				} else if strings.Contains(query, "LIMIT") {
					return strings.Replace(query, "LIMIT", "WHERE "+translationWhere+" LIMIT", 1)
				} else {
					return query + " WHERE " + translationWhere
				}
			}
		}

		// Apply to both queries
		queryResult.Query = addWhereClause(queryResult.Query)
		queryResult.CountQuery = addWhereClause(queryResult.CountQuery)
	}

	return queryResult, nil
}

// createTranslatedSortClauses modifies sort clauses for translation fields
func createTranslatedSortClauses(sortClauses []string, langCode string, multiLangFields []string) []string {
	if len(sortClauses) == 0 {
		return sortClauses
	}

	var modifiedSort []string

	for _, sortClause := range sortClauses {
		var direction string
		var fieldName string

		// Parse the sort clause to extract field name and direction
		if strings.HasPrefix(sortClause, "+") {
			direction = "ASC"
			fieldName = strings.TrimPrefix(sortClause, "+")
		} else if strings.HasPrefix(sortClause, "-") {
			direction = "DESC"
			fieldName = strings.TrimPrefix(sortClause, "-")
		} else {
			// Fallback: check if it contains the field name with ASC/DESC
			parts := strings.Fields(sortClause)
			if len(parts) >= 2 {
				fieldName = parts[0]
				if strings.ToUpper(parts[1]) == "DESC" {
					direction = "DESC"
				} else {
					direction = "ASC"
				}
			} else {
				// Default to ASC if no direction specified
				fieldName = sortClause
				direction = "ASC"
			}
		}

		// Skip empty field names
		fieldName = strings.TrimSpace(fieldName)
		if fieldName == "" {
			continue
		}

		// Check if this is a translation field
		isTranslationField := false
		for _, field := range multiLangFields {
			if fieldName == field {
				isTranslationField = true
				break
			}
		}

		var sortExpression string
		if isTranslationField {
			// Handle translation fields with JSON extraction
			if langCode == "all" {
				// For "all", use DEFAULT_LANG_CODE for meaningful sorting
				jsonPath := fmt.Sprintf("$.\"%s\"", translations.DEFAULT_LANG_CODE)
				// Check if field is valid JSON before extracting, otherwise use the raw field
				sortExpression = fmt.Sprintf("CASE WHEN JSON_VALID(%s) THEN COALESCE(JSON_UNQUOTE(JSON_EXTRACT(%s, '%s')), CAST(%s AS CHAR)) ELSE CAST(%s AS CHAR) END %s",
					util.SafeMySQLNaming(fieldName), util.SafeMySQLNaming(fieldName), jsonPath, util.SafeMySQLNaming(fieldName), util.SafeMySQLNaming(fieldName), direction)
			} else {
				// For specific language, extract and sort by that language's value with NULL handling
				jsonPath := fmt.Sprintf("$.\"%s\"", langCode)
				// Check if field is valid JSON before extracting, otherwise use the raw field
				sortExpression = fmt.Sprintf("CASE WHEN JSON_VALID(%s) THEN COALESCE(JSON_UNQUOTE(JSON_EXTRACT(%s, '%s')), CAST(%s AS CHAR)) ELSE CAST(%s AS CHAR) END %s",
					util.SafeMySQLNaming(fieldName), util.SafeMySQLNaming(fieldName), jsonPath, util.SafeMySQLNaming(fieldName), util.SafeMySQLNaming(fieldName), direction)
			}
		} else {
			// For non-translation fields, just use the field name with direction and backticks
			sortExpression = fmt.Sprintf("`%s` %s", fieldName, direction)
		}

		modifiedSort = append(modifiedSort, sortExpression)
	}

	return modifiedSort
}

// createTranslationCondition creates a SQL condition for translation field filtering
func createTranslationCondition(filter qapi.Filter, fieldName string) string {
	switch filter.Operation {
	case qapi.LK: // LIKE operation - search for partial matches in all language values (case-insensitive)
		// Use JSON_SEARCH with wildcards - this works with MySQL 5.7+
		return fmt.Sprintf("JSON_SEARCH(LOWER(`%s`), 'one', LOWER('%%%s%%')) IS NOT NULL", fieldName, filter.Value)
	case qapi.EQ: // EQUAL operation - search for exact matches in all language values (case-insensitive)
		// For exact matches, search for the exact value
		return fmt.Sprintf("JSON_SEARCH(LOWER(`%s`), 'one', LOWER('%s')) IS NOT NULL", fieldName, filter.Value)
	case qapi.NEQ: // NOT EQUAL operation
		return fmt.Sprintf("JSON_SEARCH(LOWER(`%s`), 'one', LOWER('%s')) IS NULL", fieldName, filter.Value)
	case qapi.IN, qapi.IN_ALT:
		// For IN operations, check if any JSON value matches any of the IN values
		var separator string
		if filter.Operation == qapi.IN {
			separator = "|"
		} else {
			separator = "*"
		}
		values := strings.Split(filter.Value, separator)
		conditions := make([]string, len(values))
		for i, value := range values {
			conditions[i] = fmt.Sprintf("JSON_SEARCH(LOWER(`%s`), 'one', LOWER('%s')) IS NOT NULL", fieldName, strings.TrimSpace(value))
		}
		return "(" + strings.Join(conditions, " OR ") + ")"
	default:
		// For any other operations, treat as LIKE operation
		return fmt.Sprintf("JSON_SEARCH(LOWER(`%s`), 'one', LOWER('%%%s%%')) IS NOT NULL", fieldName, filter.Value)
	}
}

// createTranslationConditionForLanguage creates a SQL condition for a specific language
func createTranslationConditionForLanguage(filter qapi.Filter, fieldName string, langCode string) string {
	jsonPath := fmt.Sprintf("$.\"%s\"", langCode)

	switch filter.Operation {
	case qapi.LK: // LIKE operation
		// Handle both JSON and non-JSON values
		return fmt.Sprintf("(CASE WHEN JSON_VALID(`%s`) THEN (JSON_EXTRACT(`%s`, '%s') IS NOT NULL AND LOWER(JSON_UNQUOTE(JSON_EXTRACT(`%s`, '%s'))) LIKE LOWER('%%%s%%')) ELSE LOWER(CAST(`%s` AS CHAR)) LIKE LOWER('%%%s%%') END)",
			fieldName, fieldName, jsonPath, fieldName, jsonPath, filter.Value, fieldName, filter.Value)
	case qapi.EQ: // EQUAL operation
		// Handle both JSON and non-JSON values
		return fmt.Sprintf("(CASE WHEN JSON_VALID(`%s`) THEN (JSON_EXTRACT(`%s`, '%s') IS NOT NULL AND LOWER(JSON_UNQUOTE(JSON_EXTRACT(`%s`, '%s'))) = LOWER('%s')) ELSE LOWER(CAST(`%s` AS CHAR)) = LOWER('%s') END)",
			fieldName, fieldName, jsonPath, fieldName, jsonPath, filter.Value, fieldName, filter.Value)
	case qapi.NEQ: // NOT EQUAL operation
		// Handle both JSON and non-JSON values
		return fmt.Sprintf("(CASE WHEN JSON_VALID(`%s`) THEN (JSON_EXTRACT(`%s`, '%s') IS NULL OR LOWER(JSON_UNQUOTE(JSON_EXTRACT(`%s`, '%s'))) != LOWER('%s')) ELSE LOWER(CAST(`%s` AS CHAR)) != LOWER('%s') END)",
			fieldName, fieldName, jsonPath, fieldName, jsonPath, filter.Value, fieldName, filter.Value)
	case qapi.GT:
		return fmt.Sprintf("(CASE WHEN JSON_VALID(`%s`) THEN (JSON_EXTRACT(`%s`, '%s') IS NOT NULL AND JSON_UNQUOTE(JSON_EXTRACT(`%s`, '%s')) > '%s') ELSE CAST(`%s` AS CHAR) > '%s' END)",
			fieldName, fieldName, jsonPath, fieldName, jsonPath, filter.Value, fieldName, filter.Value)
	case qapi.GTE:
		return fmt.Sprintf("(CASE WHEN JSON_VALID(`%s`) THEN (JSON_EXTRACT(`%s`, '%s') IS NOT NULL AND JSON_UNQUOTE(JSON_EXTRACT(`%s`, '%s')) >= '%s') ELSE CAST(`%s` AS CHAR) >= '%s' END)",
			fieldName, fieldName, jsonPath, fieldName, jsonPath, filter.Value, fieldName, filter.Value)
	case qapi.LT:
		return fmt.Sprintf("(CASE WHEN JSON_VALID(`%s`) THEN (JSON_EXTRACT(`%s`, '%s') IS NOT NULL AND JSON_UNQUOTE(JSON_EXTRACT(`%s`, '%s')) < '%s') ELSE CAST(`%s` AS CHAR) < '%s' END)",
			fieldName, fieldName, jsonPath, fieldName, jsonPath, filter.Value, fieldName, filter.Value)
	case qapi.LTE:
		return fmt.Sprintf("(CASE WHEN JSON_VALID(`%s`) THEN (JSON_EXTRACT(`%s`, '%s') IS NOT NULL AND JSON_UNQUOTE(JSON_EXTRACT(`%s`, '%s')) <= '%s') ELSE CAST(`%s` AS CHAR) <= '%s' END)",
			fieldName, fieldName, jsonPath, fieldName, jsonPath, filter.Value, fieldName, filter.Value)
	case qapi.IN, qapi.IN_ALT:
		values := strings.Split(filter.Value, "|")
		jsonConditions := make([]string, len(values))
		nonJsonConditions := make([]string, len(values))
		for i, value := range values {
			jsonConditions[i] = fmt.Sprintf("LOWER(JSON_UNQUOTE(JSON_EXTRACT(`%s`, '%s'))) = LOWER('%s')", fieldName, jsonPath, value)
			nonJsonConditions[i] = fmt.Sprintf("LOWER(CAST(`%s` AS CHAR)) = LOWER('%s')", fieldName, value)
		}
		// Handle both JSON and non-JSON values
		return fmt.Sprintf("(CASE WHEN JSON_VALID(`%s`) THEN (JSON_EXTRACT(`%s`, '%s') IS NOT NULL AND (%s)) ELSE (%s) END)",
			fieldName, fieldName, jsonPath, strings.Join(jsonConditions, " OR "), strings.Join(nonJsonConditions, " OR "))
	default:
		// Default to LIKE operation with JSON validity check
		return fmt.Sprintf("(CASE WHEN JSON_VALID(`%s`) THEN (JSON_EXTRACT(`%s`, '%s') IS NOT NULL AND LOWER(JSON_UNQUOTE(JSON_EXTRACT(`%s`, '%s'))) LIKE LOWER('%%%s%%')) ELSE LOWER(CAST(`%s` AS CHAR)) LIKE LOWER('%%%s%%') END)",
			fieldName, fieldName, jsonPath, fieldName, jsonPath, filter.Value, fieldName, filter.Value)
	}
}

// createTranslatedFields modifies the fields selection for translation support
func createTranslatedFields(fields []string, langCode string, multiLangFields []string, records any) []string {
	if len(fields) == 0 {
		// If no specific fields requested, handle translation fields in the SELECT clause
		return buildTranslatedSelectFields(langCode, multiLangFields, records)
	}

	// If specific fields are requested, modify translation fields
	modifiedFields := make([]string, len(fields))
	copy(modifiedFields, fields)

	if langCode != "" {
		for i, field := range modifiedFields {
			for _, multiLangField := range multiLangFields {
				if field == multiLangField {
					var jsonPath string
					if langCode == "all" {
						// For "all", use DEFAULT_LANG_CODE for meaningful field selection
						jsonPath = fmt.Sprintf("$.\"%s\"", translations.DEFAULT_LANG_CODE)
					} else {
						// For specific language, extract only that language's value
						jsonPath = fmt.Sprintf("$.\"%s\"", langCode)
					}
					// Check if field is valid JSON before extracting, otherwise use the raw field
					modifiedFields[i] = fmt.Sprintf("CASE WHEN JSON_VALID(`%s`) THEN COALESCE(JSON_UNQUOTE(JSON_EXTRACT(`%s`, '%s')), '') ELSE CAST(`%s` AS CHAR) END AS `%s`",
						field, field, jsonPath, field, field)
				}
			}
		}
	}

	return modifiedFields
}

// buildTranslatedSelectFields builds SELECT fields for translation support
func buildTranslatedSelectFields(langCode string, multiLangFields []string, records any) []string {
	if langCode == "" {
		return nil // No translation handling needed
	}

	recordType := reflect.TypeOf(records)
	if recordType.Kind() == reflect.Ptr {
		recordType = recordType.Elem()
	}
	if recordType.Kind() == reflect.Slice {
		recordType = recordType.Elem()
		if recordType.Kind() == reflect.Ptr {
			recordType = recordType.Elem()
		}
	}

	if recordType.Kind() != reflect.Struct {
		return nil
	}

	var selectFields []string

	for i := 0; i < recordType.NumField(); i++ {
		field := recordType.Field(i)

		// Skip unexported fields
		if !field.IsExported() {
			continue
		}

		fieldName := field.Name

		// Check if it's a translation field
		isTranslationField := false
		for _, multiLangField := range multiLangFields {
			if fieldName == multiLangField {
				isTranslationField = true
				break
			}
		}

		if isTranslationField {
			if langCode == "all" {
				// For "all", use DEFAULT_LANG_CODE for meaningful field selection
				jsonPath := fmt.Sprintf("$.\"%s\"", translations.DEFAULT_LANG_CODE)
				selectFields = append(selectFields, fmt.Sprintf("CASE WHEN JSON_VALID(`%s`) THEN COALESCE(JSON_UNQUOTE(JSON_EXTRACT(`%s`, '%s')), '') ELSE CAST(`%s` AS CHAR) END AS `%s`",
					fieldName, fieldName, jsonPath, fieldName, fieldName))
			} else {
				// For specific language, extract only that language's value with NULL handling
				jsonPath := fmt.Sprintf("$.\"%s\"", langCode)
				selectFields = append(selectFields, fmt.Sprintf("CASE WHEN JSON_VALID(`%s`) THEN COALESCE(JSON_UNQUOTE(JSON_EXTRACT(`%s`, '%s')), '') ELSE CAST(`%s` AS CHAR) END AS `%s`",
					fieldName, fieldName, jsonPath, fieldName, fieldName))
			}
		} else {
			// Regular field
			selectFields = append(selectFields, fmt.Sprintf("`%s`", fieldName))
		}
	}

	return selectFields
}

// findTranslationFields finds fields that use the Translations type or JSON type for translations
func findTranslationFields(record any) []string {
	recordType := reflect.TypeOf(record)
	if recordType.Kind() == reflect.Ptr {
		recordType = recordType.Elem()
	}

	// Handle slice types to get the element type
	if recordType.Kind() == reflect.Slice {
		recordType = recordType.Elem()
		if recordType.Kind() == reflect.Ptr {
			recordType = recordType.Elem()
		}
	}

	if recordType.Kind() != reflect.Struct {
		return nil
	}

	// Get the Translations type for comparison
	translationsType := reflect.TypeOf(translations.Translations{})
	translationsPtrType := reflect.TypeOf((*translations.Translations)(nil)).Elem()

	// Also get the JSON type for comparison (sometimes used for translations)
	jsonType := reflect.TypeOf(types.JSON{})

	translationFields := []string{}
	for i := 0; i < recordType.NumField(); i++ {
		field := recordType.Field(i)
		fieldType := field.Type

		// Handle pointer types
		if fieldType.Kind() == reflect.Ptr {
			fieldType = fieldType.Elem()
		}

		// Check if field type is Translations (either value or pointer)
		if fieldType == translationsType || fieldType == translationsPtrType {
			translationFields = append(translationFields, field.Name)
		} else if fieldType == jsonType {
			// For JSON fields, check if they might be translation fields
			// We can identify them by common naming patterns or tags
			fieldName := strings.ToLower(field.Name)
			if strings.Contains(fieldName, "name") ||
				strings.Contains(fieldName, "title") ||
				strings.Contains(fieldName, "description") ||
				strings.Contains(fieldName, "label") ||
				field.Tag.Get("translation") == "true" {
				translationFields = append(translationFields, field.Name)
			}
		}
	}
	return translationFields
}

// Create Record (optimized with interfaces, fallback to reflection)
func Create(DB *sqlx.DB, record any) error {
	tableName := getTableName(record)

	// Try optimized field mapping approach
	if fieldMapper, ok := record.(interfaces.FieldMapper); ok {
		return createWithFieldMap(DB, tableName, fieldMapper.GetFieldMap(), record)
	}

	// Fallback to reflection
	return createWithReflection(DB, record, tableName)
}

// createWithFieldMap creates using field map (optimized)
func createWithFieldMap(DB *sqlx.DB, tableName string, fieldMap map[string]interface{}, record any) error {
	if len(fieldMap) == 0 {
		return fmt.Errorf("no fields provided for insert")
	}

	var columns []string
	var placeholders []string
	var values []interface{}

	// Get the record type for field type checking
	recordType := reflect.TypeOf(record)
	if recordType.Kind() == reflect.Ptr {
		recordType = recordType.Elem()
	}

	// Handle JSON fields and process field values
	for k, v := range fieldMap {
		fieldMap[k] = processFieldValue(v)
	}

	// Auto-set timestamps for Create operation
	setTimestampFields(fieldMap, recordType, true)

	for column, value := range fieldMap {
		// Skip ID field (assuming auto-increment)
		if strings.ToLower(column) == "id" {
			continue
		}
		columns = append(columns, util.SafeMySQLNaming(column))
		placeholders = append(placeholders, "?")
		values = append(values, value)
	}

	if len(columns) == 0 {
		return fmt.Errorf("no valid fields for insert")
	}

	query := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)",
		util.SafeMySQLNaming(tableName),
		strings.Join(columns, ", "),
		strings.Join(placeholders, ", "))

	result, err := DB.Exec(query, values...)
	if err != nil {
		logging.Logger.Error("Create error", zap.String("table", tableName), zap.Error(err))
		return err
	}

	// Set the ID if possible
	if id, err := result.LastInsertId(); err == nil {
		setRecordID(record, uint64(id))
	}

	return nil
}

// createWithReflection is the reflection-based implementation
func createWithReflection(DB *sqlx.DB, record any, tableName string) error {
	typ, _ := queryapi.GetTableName(record)

	var columns []string
	var placeholders []string
	var values []interface{}

	recordValue := reflect.ValueOf(record)
	if recordValue.Kind() == reflect.Ptr {
		recordValue = recordValue.Elem()
	}

	// Auto-set timestamps for Create operation
	setTimestampFieldsReflection(recordValue, true)

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

		value := fieldValue.Interface()
		value = processFieldValueWithReflection(value, fieldValue)

		columns = append(columns, util.SafeMySQLNaming(field.Name))
		placeholders = append(placeholders, "?")
		values = append(values, value)
	}

	query := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)",
		util.SafeMySQLNaming(tableName),
		strings.Join(columns, ", "),
		strings.Join(placeholders, ", "))

	result, err := DB.Exec(query, values...)
	if err != nil {
		logging.Logger.Error("Create error", zap.Any("Model", reflect.TypeOf(record)), zap.Error(err))
		return err
	}

	// Set the ID
	if id, err := result.LastInsertId(); err == nil {
		setRecordID(record, uint64(id))
	}

	return nil
}

// Read Record (optimized with interfaces, fallback to reflection)
func Read(DB *sqlx.DB, record any, id any, preloads ...string) error {
	// Note: preloads are ignored in sqlx implementation
	// Users need to handle relationships manually

	tableName := getTableName(record)

	query := fmt.Sprintf("SELECT * FROM %s WHERE ID = ?", util.SafeMySQLNaming(tableName))
	err := DB.Get(record, query, id)
	if err != nil {
		logging.Logger.Error("Read error", zap.String("table", tableName), zap.Error(err), zap.Any("id", id))
	}
	return err
}

// Update Updates the record with the given fields (optimized with interfaces, fallback to reflection)
func Update(DB *sqlx.DB, model any, fieldsParams ...map[string]any) error {
	tableName := getTableName(model)

	// Partial update with specific fields
	if len(fieldsParams) > 0 && fieldsParams[0] != nil {
		id, err := getRecordID(model)
		if err != nil {
			return fmt.Errorf("failed to get ID for update: %v", err)
		}
		return updateWithFieldMap(DB, tableName, id, fieldsParams[0])
	}

	// Full record update - try optimized approach first
	if fieldMapper, ok := model.(interfaces.FieldMapper); ok {
		id, err := getRecordID(model)
		if err != nil {
			return fmt.Errorf("failed to get ID for update: %v", err)
		}

		fieldMap := fieldMapper.GetFieldMap()
		// Remove ID from field map for update
		delete(fieldMap, "ID")
		delete(fieldMap, "id")

		return updateWithFieldMap(DB, tableName, id, fieldMap)
	}

	// Fallback to reflection
	return updateWithReflection(DB, model, tableName)
}

// updateWithFieldMap updates using field map (optimized)
func updateWithFieldMap(DB *sqlx.DB, tableName string, id interface{}, fieldMap map[string]interface{}) error {
	if len(fieldMap) == 0 {
		return fmt.Errorf("no fields provided for update")
	}

	var setClauses []string
	var values []interface{}

	// Auto-set UpdatedAt timestamp for Update operation
	if _, exists := fieldMap["UpdatedAt"]; !exists {
		fieldMap["UpdatedAt"] = time.Now()
	}

	// Handle JSON fields and process field values
	for k, v := range fieldMap {
		fieldMap[k] = processFieldValue(v)
	}

	for k, v := range fieldMap {
		setClauses = append(setClauses, fmt.Sprintf("%s = ?", util.SafeMySQLNaming(k)))
		values = append(values, v)
	}

	query := fmt.Sprintf("UPDATE %s SET %s WHERE ID = ?",
		util.SafeMySQLNaming(tableName),
		strings.Join(setClauses, ", "))

	values = append(values, id)

	_, err := DB.Exec(query, values...)
	if err != nil {
		logging.Logger.Error("Update error", zap.String("table", tableName), zap.Error(err))
	}
	return err
}

// updateWithReflection is the reflection-based implementation
func updateWithReflection(DB *sqlx.DB, model any, tableName string) error {
	typ, _ := queryapi.GetTableName(model)

	var setClauses []string
	var values []interface{}
	var whereValue interface{}

	modelValue := reflect.ValueOf(model)
	if modelValue.Kind() == reflect.Ptr {
		modelValue = modelValue.Elem()
	}

	// Auto-set UpdatedAt timestamp for Update operation
	setTimestampFieldsReflection(modelValue, false)

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

		value := fieldValue.Interface()
		value = processFieldValueWithReflection(value, fieldValue)

		setClauses = append(setClauses, fmt.Sprintf("%s = ?", util.SafeMySQLNaming(field.Name)))
		values = append(values, value)
	}

	if whereValue == nil {
		return fmt.Errorf("no ID field found for update")
	}

	query := fmt.Sprintf("UPDATE %s SET %s WHERE ID = ?",
		util.SafeMySQLNaming(tableName),
		strings.Join(setClauses, ", "))

	values = append(values, whereValue)

	_, err := DB.Exec(query, values...)
	if err != nil {
		logging.Logger.Error("Update error", zap.Any("Model", reflect.TypeOf(model)), zap.Error(err))
	}
	return err
}

// DeleteAll Records - handles multiple records by ID
func DeleteAll(DB *sqlx.DB, record any, ids ...any) error {
	if len(ids) == 0 {
		return fmt.Errorf("no IDs provided")
	}

	// Collect all individual ID values
	var idValues []any

	for _, id := range ids {
		// Handle pointer types by dereferencing
		if reflect.TypeOf(id).Kind() == reflect.Ptr {
			id = reflect.ValueOf(id).Elem().Interface()
		}

		switch v := id.(type) {
		case *fiber.Ctx:
			// Handle Fiber context - parse JSON body
			var jsonArray []any
			if err := json.Unmarshal(v.Body(), &jsonArray); err != nil {
				return fmt.Errorf("failed to parse JSON: %v", err)
			}
			idValues = append(idValues, jsonArray...)
		default:
			// Handle slices using reflection
			val := reflect.ValueOf(v)
			if val.Kind() == reflect.Slice {
				// Expand any slice type into individual values
				for i := 0; i < val.Len(); i++ {
					idValues = append(idValues, val.Index(i).Interface())
				}
			} else {
				// Single value
				idValues = append(idValues, v)
			}
		}
	}

	if len(idValues) == 0 {
		return fmt.Errorf("no valid IDs found")
	}

	tableName := getTableName(record)

	// Build query with placeholders
	placeholders := strings.Repeat("?,", len(idValues))
	placeholders = placeholders[:len(placeholders)-1]
	query := fmt.Sprintf("DELETE FROM %s WHERE ID IN (%s)", util.SafeMySQLNaming(tableName), placeholders)

	// Execute with individual values (not slices)
	_, err := DB.Exec(query, idValues...)
	return err
}

// Delete Record (optimized with interfaces, fallback to reflection)
func Delete(DB *sqlx.DB, record any) error {
	tableName := getTableName(record)

	id, err := getRecordID(record)
	if err != nil {
		return fmt.Errorf("failed to get ID for delete: %v", err)
	}

	query := fmt.Sprintf("DELETE FROM %s WHERE ID = ?", util.SafeMySQLNaming(tableName))
	_, err = DB.Exec(query, id)
	if err != nil {
		logging.Logger.Error("Delete error", zap.String("table", tableName), zap.Error(err), zap.Any("id", id))
	}
	return err
}
