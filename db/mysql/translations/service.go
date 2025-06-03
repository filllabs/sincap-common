package translations

import (
	"fmt"
	"reflect"
	"strings"
	"sync"

	"github.com/filllabs/sincap-common/db/queryapi"
	"github.com/filllabs/sincap-common/db/util"
	"github.com/filllabs/sincap-common/logging"
	"github.com/filllabs/sincap-common/middlewares/qapi"
	"github.com/filllabs/sincap-common/reflection"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// List retrieves records from the database based on query parameters with enhanced support for translations
func List(DB *gorm.DB, records any, query *qapi.Query, lang []string) (int, error) {
	langCode := lang[0]

	value := reflect.ValueOf(records)
	if value.Kind() != reflect.Pointer {
		return 0, fmt.Errorf("records must be a pointer")
	}

	elem := value.Elem()
	if elem.Kind() != reflect.Slice {
		return 0, fmt.Errorf("records must be a pointer to slice")
	}

	// Get entity type and table name
	entityType, tableName := queryapi.GetTableName(records)
	calculateCount := query.Offset > 0 || query.Limit > 0

	// Initialize database query builder
	db := DB.Table(tableName)

	// Check if we need to handle translations
	useTranslations := len(langCode) > 0 && langCode != ""
	multiLangFields := findTranslationFields(records)

	// Build query with or without translations
	var err error
	if useTranslations && (len(multiLangFields) > 0 || len(query.Preloads) > 0) {
		path := getLanguagePath(langCode)
		db, err = generateTranslatedDB(db, query, path, entityType, multiLangFields, tableName)
	} else {
		db, err = queryapi.GenerateDB(query, db, records)
	}

	if err != nil {
		return 0, err
	}

	// Add Q parameter search support
	if len(query.Q) > 0 && useTranslations && len(multiLangFields) > 0 {
		db = addQSearch(db, query.Q, langCode, multiLangFields)
	} else if len(query.Q) > 0 {
		where, values, err := q2Sql(query.Q, entityType, tableName)
		if err != nil {
			return 0, err
		}
		db = db.Where(where, values...)
	}

	// Get total count if pagination is used
	count, err := handlePagination(db, calculateCount, query)
	if err != nil {
		return 0, err
	}

	// Add preloads with enhanced translation support
	db = addPreloads(db, query.Preloads, langCode, entityType)

	// Build optimized select clause for specific fields or translations
	db = buildOptimizedSelectClause(db, query, langCode, entityType, multiLangFields)

	// Execute the query
	result := db.Find(records)
	if result.Error != nil {
		return 0, result.Error
	}

	if !calculateCount {
		return reflect.ValueOf(records).Elem().Len(), nil
	}
	return count, nil
}

// addQSearch performs a search across all translation fields for the given query string
func addQSearch(db *gorm.DB, query string, langCode string, multiLangFields []string) *gorm.DB {
	var conditions []string
	var values []interface{}

	// Search in translation fields
	for _, field := range multiLangFields {
		jsonPath := fmt.Sprintf("$.\"%s\"", langCode)
		conditions = append(conditions,
			fmt.Sprintf("LOWER(JSON_UNQUOTE(JSON_EXTRACT(`%s`, '%s'))) LIKE LOWER(?)",
				field, jsonPath))
		values = append(values, "%"+query+"%")
	}

	if len(conditions) > 0 {
		return db.Where(strings.Join(conditions, " OR "), values...)
	}
	return db
}

// handlePagination applies pagination and returns the total count if needed
func handlePagination(db *gorm.DB, calculateCount bool, query *qapi.Query) (int, error) {
	var count int64 = -1
	if calculateCount {
		cDB := db.Count(&count)
		if cDB.Error != nil {
			return 0, cDB.Error
		}
		// Add offset and limit
		db = db.Offset(query.Offset)
		db = db.Limit(query.Limit)
	}
	return int(count), nil
}

// generateTranslatedDB handles the complex logic for queries with translations
func generateTranslatedDB(db *gorm.DB, query *qapi.Query, langCode string,
	entityType reflect.Type, multiLangFields []string, tableName string) (*gorm.DB, error) {

	// Find translation fields in preloaded models
	nestedMultiLangFields, m2mFields := findNestedTranslationFields(query.Preloads, entityType)

	// Handle sorting with translations
	db = handleTranslatedSorting(db, query, langCode, multiLangFields, nestedMultiLangFields, tableName, entityType)

	// Handle one-to-many relationship filters
	db = handleTranslatedOneToManyFilter(db, query, entityType)

	// Handle filters with translations
	db = handleTranslatedFilters(db, query, langCode, entityType, multiLangFields,
		nestedMultiLangFields, m2mFields)

	return db, nil
}

// handleTranslatedSorting applies sorting with translation field awareness
func handleTranslatedSorting(db *gorm.DB, query *qapi.Query, langCode string,
	multiLangFields []string, nestedMultiLangFields map[string][]string,
	tableName string, entityType reflect.Type) *gorm.DB {

	if len(query.Sort) == 0 {
		return db
	}

	for _, sortClause := range query.Sort {
		handled := false

		if strings.Contains(sortClause, ".") {
			// Handle sorting on related model fields
			parts := strings.SplitN(sortClause, ".", 2)
			relation := parts[0]
			fieldName := parts[1]
			sortParts := strings.Split(sortClause, " ")
			sortDirection := "ASC"
			if len(sortParts) > 1 {
				sortDirection = sortParts[1]
			}

			if fields, exists := nestedMultiLangFields[relation]; exists {
				for _, multiLangField := range fields {
					if fieldName == multiLangField {
						jsonPath := fmt.Sprintf("$.\"%s\"", langCode)
						db = db.Joins(fmt.Sprintf("JOIN %s ON %s.ID = %sID",
							relation, relation, strings.ToLower(relation))).
							Order(fmt.Sprintf("JSON_UNQUOTE(JSON_EXTRACT(%s.%s, '%s')) %s",
								relation, fieldName, jsonPath, sortDirection))
						handled = true
						break
					}
				}
			}
		} else {
			// Handle sorting on main model fields
			field := strings.Split(sortClause, " ")[0]
			sortDirection := "ASC"
			if len(strings.Split(sortClause, " ")) > 1 {
				sortDirection = strings.Split(sortClause, " ")[1]
			}

			// Check if it's a translation field
			isTranslationField := false
			for _, tf := range multiLangFields {
				if field == tf {
					isTranslationField = true
					jsonPath := fmt.Sprintf("$.\"%s\"", langCode)
					db = db.Order(fmt.Sprintf("JSON_UNQUOTE(JSON_EXTRACT(%s, '%s')) %s",
						field, jsonPath, sortDirection))
					handled = true
					break
				}
			}

			// If not a translation field, check if it's a JSON field
			if !isTranslationField && !handled {
				db = handleJsonFieldSorting(db, sortClause, tableName, entityType)
				handled = true
			}
		}

		// If not handled as a translation or JSON field, use standard ordering
		if !handled {
			db = db.Order(sortClause)
		}
	}

	return db
}

// handleJsonFieldSorting handles sorting for JSON fields
func handleJsonFieldSorting(db *gorm.DB, sortField string, tableName string, entityType reflect.Type) *gorm.DB {
	values := strings.Split(sortField, " ")
	fieldNames := strings.Split(values[0], ".")

	if len(fieldNames) < 2 {
		return db.Order(sortField)
	}

	field, isFieldFound := entityType.FieldByName(fieldNames[0])

	if isFieldFound {
		dp := reflection.DepointerField(field.Type)
		// Check if it's a JSON type
		if dp.Kind() == reflect.Map || strings.Contains(dp.String(), "json") {
			c := "CAST(" + tableName + "." + fieldNames[0] + "->" + "'$." + fieldNames[1] + "'" + "AS CHAR) " + values[1]
			return db.Order(c)
		}
	}
	return db.Order(sortField)
}

// handleTranslatedFilters applies filters with translation field awareness
func handleTranslatedFilters(db *gorm.DB, query *qapi.Query, langCode string,
	entityType reflect.Type, multiLangFields []string,
	_ map[string][]string, m2mFields map[string]string) *gorm.DB {

	if len(query.Filter) == 0 {
		return db
	}

	for _, v := range query.Filter {
		// Skip filters that contain a dot, as these are handled by handleTranslatedOneToManyFilter
		if strings.Contains(v.Name, ".") {
			continue
		}

		handled := false

		// Check if this is a polymorphic relationship
		db, handled = handlePolymorphicTranslationFilter(db, entityType, v, langCode)
		if handled {
			continue
		}

		// Check if this is a many2many field filter
		if m2mTable, exists := m2mFields[v.Name]; exists {
			field, found := entityType.FieldByName(v.Name)
			if !found {
				continue
			}
			relatedType := getRelatedModelType(field.Type)
			relatedTypeName := relatedType.Name()

			db = db.Where(fmt.Sprintf("EXISTS (SELECT 1 FROM %s m2m JOIN %s rel ON m2m.%sID = rel.ID WHERE m2m.%sID = %s.ID AND rel.ID LIKE ?)",
				m2mTable,
				relatedTypeName,
				relatedTypeName,
				entityType.Name(),
				entityType.Name()),
				"%"+v.Value+"%")
			handled = true
			continue
		}

		// Check for translation fields in main model
		if !handled {
			for _, multiLangField := range multiLangFields {
				if v.Name == multiLangField {
					jsonPath := fmt.Sprintf("$.\"%s\"", langCode)
					db = db.Where("LOWER(JSON_UNQUOTE(JSON_EXTRACT("+v.Name+", ?))) LIKE LOWER(?)",
						jsonPath, "%"+v.Value+"%")
					handled = true
					break
				}
			}
		}

		// Default handling for regular fields
		if !handled {
			db = db.Where("LOWER("+v.Name+") LIKE LOWER(?)", "%"+v.Value+"%")
		}
	}

	return db
}

// handlePolymorphicTranslationFilter handles filtering for polymorphic relationships
func handlePolymorphicTranslationFilter(db *gorm.DB, entityType reflect.Type,
	filter qapi.Filter, langCode string) (*gorm.DB, bool) {

	if !strings.Contains(filter.Name, ".") {
		return db, false
	}

	parts := strings.SplitN(filter.Name, ".", 2)
	relation := parts[0]

	field, found := entityType.FieldByName(relation)
	if !found {
		return db, false
	}

	if prefix, isPoly := util.GetPolymorphic(&field); isPoly {
		// Handle polymorphic relationship with translations
		polyID := prefix + "ID"
		relatedType := getRelatedModelType(field.Type)
		relatedTable := relatedType.Name()

		// Check if related type has translation fields
		relatedModel := reflect.New(relatedType).Interface()
		relatedMultiLangFields := findTranslationFields(relatedModel)

		if len(relatedMultiLangFields) > 0 {
			for _, multiLangField := range relatedMultiLangFields {
				if parts[1] == multiLangField {
					jsonPath := fmt.Sprintf("$.\"%s\"", langCode)
					db = db.Where(fmt.Sprintf("ID IN (SELECT %s FROM %s WHERE LOWER(JSON_UNQUOTE(JSON_EXTRACT(%s, ?))) LIKE LOWER(?))",
						polyID, relatedTable, multiLangField),
						jsonPath, "%"+filter.Value+"%")
					return db, true
				}
			}
		}
	}

	return db, false
}

// buildOptimizedSelectClause creates an optimized SELECT clause for the query
func buildOptimizedSelectClause(db *gorm.DB, query *qapi.Query, langCode string,
	entityType reflect.Type, multiLangFields []string) *gorm.DB {

	// If specific fields are requested
	if len(query.Fields) > 0 {
		if langCode != "" && len(multiLangFields) > 0 {
			// Only translate fields that are in both query.Fields and multiLangFields
			var translatedFields []string
			requestedFields := make(map[string]bool)

			for _, field := range query.Fields {
				requestedFields[field] = true
			}

			for i := 0; i < entityType.NumField(); i++ {
				field := entityType.Field(i)
				columnName := getColumnName(field)

				if !requestedFields[columnName] {
					continue
				}

				isMultiLang := false
				for _, mlField := range multiLangFields {
					if field.Name == mlField {
						isMultiLang = true
						break
					}
				}

				if isMultiLang {
					jsonPath := fmt.Sprintf("$.\"%s\"", langCode)
					translatedFields = append(translatedFields,
						fmt.Sprintf("JSON_UNQUOTE(JSON_EXTRACT(`%s`, '%s')) AS `%s`",
							columnName, jsonPath, columnName))
				} else {
					translatedFields = append(translatedFields, fmt.Sprintf("`%s`", columnName))
				}
			}

			if len(translatedFields) > 0 {
				return db.Select(strings.Join(translatedFields, ", "))
			}
		}
		return db.Select(query.Fields)
	}

	// If no specific fields, use original translation logic
	if langCode != "" && len(multiLangFields) > 0 {
		selectClause := buildTranslatedSelectClause(entityType, multiLangFields, langCode)
		if len(selectClause) > 0 {
			return db.Select(strings.Join(selectClause, ", "))
		}
	}

	return db
}

// getColumnName retrieves the column name from struct field
func getColumnName(field reflect.StructField) string {
	columnName := field.Name
	if tag, ok := field.Tag.Lookup("gorm"); ok {
		tagParts := strings.Split(tag, ";")
		for _, part := range tagParts {
			if strings.HasPrefix(part, "column:") {
				columnName = strings.TrimPrefix(part, "column:")
				break
			}
		}
	}
	return columnName
}

// addPreloads adds all preload statements to the query with enhanced translation support
func addPreloads(db *gorm.DB, preloads []string, langCode string, entityType reflect.Type) *gorm.DB {
	for _, preload := range preloads {
		field, found := entityType.FieldByName(preload)
		if !found {
			continue
		}

		relatedType := getRelatedModelType(field.Type)
		relatedModel := reflect.New(relatedType).Interface()
		nestedMultiLangFields := findTranslationFields(relatedModel)

		// Check for polymorphic relationships
		if len(nestedMultiLangFields) > 0 && langCode != "all" {
			translatedSelects := buildTranslatedSelectClause(relatedType, nestedMultiLangFields, langCode)
			db = db.Preload(preload, func(tx *gorm.DB) *gorm.DB {
				return tx.Select(translatedSelects)
			})
		} else {
			db = db.Preload(preload)
		}
	}
	return db
}

// findNestedTranslationFields finds translation fields in related models
func findNestedTranslationFields(preloads []string, entityType reflect.Type) (
	map[string][]string, map[string]string) {

	nestedMultiLangFields := make(map[string][]string)
	m2mFields := make(map[string]string)

	for _, preload := range preloads {
		field, found := entityType.FieldByName(preload)
		if !found {
			continue
		}

		// Check for many2many relationship
		if m2mTable, isM2M := util.GetMany2Many(&field); isM2M {
			m2mFields[preload] = m2mTable
		}

		// Check for polymorphic relationship
		_, isPoly := util.GetPolymorphic(&field)

		relatedModelType := getRelatedModelType(field.Type)
		relatedModel := reflect.New(relatedModelType).Interface()
		nestedFields := findTranslationFields(relatedModel)

		if len(nestedFields) > 0 {
			nestedMultiLangFields[preload] = nestedFields
		}

		// For polymorphic relationships, handle nested relationships
		if isPoly {
			// You can add recursive handling for nested polymorphic relationships here
		}
	}

	return nestedMultiLangFields, m2mFields
}

// Build SELECT clause while ignoring unwanted GORM fields
func buildTranslatedSelectClause(entityType reflect.Type, multiLangFields []string, langCode string) []string {
	var selectClause []string

	for i := 0; i < entityType.NumField(); i++ {
		field := entityType.Field(i)

		// Skip fields that should not be included in the query
		if shouldSkipField(field) {
			continue
		}

		// Determine the actual DB column name
		columnName := getColumnName(field)

		// Check if it's a multi-language field
		isMultiLang := false
		for _, multiLangField := range multiLangFields {
			if field.Name == multiLangField {
				isMultiLang = true
				break
			}
		}

		if isMultiLang {
			if langCode == "all" {
				selectClause = append(selectClause, fmt.Sprintf("`%s`", columnName))
			} else {
				selectClause = append(selectClause,
					fmt.Sprintf("JSON_OBJECT('%s', JSON_UNQUOTE(JSON_EXTRACT(`%s`, '$.\"%s\"'))) AS `%s`",
						langCode, columnName, langCode, columnName))
			}
		} else {
			selectClause = append(selectClause, fmt.Sprintf("`%s`", columnName))
		}
	}

	return selectClause
}

// Checks if a field should be skipped based on GORM tags
func shouldSkipField(field reflect.StructField) bool {
	if tag, ok := field.Tag.Lookup("gorm"); ok {
		// Split the tag by semicolons (gorm tags are typically formatted like `gorm:"foreignKey:UserID;many2many:user_roles"`)
		tags := strings.Split(tag, ";")

		// List of tags to ignore
		ignoredTags := []string{"-", "foreignKey", "many2many", "embedded", "polymorphic"}

		for _, t := range tags {
			for _, ignore := range ignoredTags {
				if strings.HasPrefix(t, ignore) {
					return true
				}
			}
		}
	}
	return false
}

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

	translationFields := []string{}
	for i := 0; i < recordType.NumField(); i++ {
		field := recordType.Field(i)

		translationsType := reflect.TypeOf((*Translations)(nil))
		if field.Type.AssignableTo(translationsType) {
			translationFields = append(translationFields, field.Name)
		}
	}
	return translationFields
}

func getRelatedModelType(t reflect.Type) reflect.Type {
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	if t.Kind() == reflect.Slice {
		t = t.Elem()
		if t.Kind() == reflect.Ptr {
			t = t.Elem()
		}
	}
	return t
}

func getLanguagePath(langCode string) string {
	if langCode == "all" {
		return DEFAULT_LANG_CODE
	}
	return langCode
}

func q2Sql(q string, typ reflect.Type, tableName string) (string, []interface{}, error) {

	// Convert q to  where condition with OR for all fields with tag
	where, values, err := generateQQuery(typ, tableName, q)
	if err != nil {
		logging.Logger.Warn("Can't create query from q", zap.Error(err))
	}
	return strings.Join(where, " OR "), values, nil
}

func generateQQuery(structType reflect.Type, tableName string, q string) ([]string, []interface{}, error) {
	var where []string
	var values []interface{}
	taggedFields := getQapiFields(structType)
	for _, field := range *taggedFields {
		if field.Typ.Kind() != reflect.Struct {
			where = append(where, column(tableName, field.Field.Name)+" LIKE ?")
			values = append(values, strings.Replace(field.Tag, "*", q, 1))
			continue
		}
		// if its is struct generate query recursively
		w, v, err := generateQQuery(field.Typ, field.TableName, q)
		var cond []string
		if err != nil {
			logging.Logger.Warn("Can't create query from q", zap.Error(err))
			continue
		}

		if prefix, isPoly := util.GetPolymorphic(&field.Field); isPoly {
			polyID := prefix + "ID"

			cond = append(cond, column(tableName, "ID"), "IN (", "SELECT", column(field.TableName, polyID), "FROM", safeMySQLNaming(field.TableName), "WHERE (")
			cond = append(cond, strings.Join(w, " OR "))
			cond = append(cond, ") )")
			where = append(where, strings.Join(cond, " "))
		} else if m2mTable, isM2M := util.GetMany2Many(&field.Field); isM2M {
			srcRef := column(m2mTable, tableName+"ID")
			destRef := column(m2mTable, field.TableName+"ID")
			cond = append(cond, column(tableName, "ID"), "IN (", "SELECT", srcRef, "FROM", safeMySQLNaming(m2mTable), "WHERE (", destRef, "IN (", "SELECT ", column(field.TableName, "ID"), " FROM", safeMySQLNaming(field.TableName), "WHERE (")
			cond = append(cond, strings.Join(w, " OR "))
			cond = append(cond, ")", ")", ")", ")")
			where = append(where, strings.Join(cond, " "))
		} else {
			cond = append(cond, column(tableName, field.Field.Name+"ID"), "IN (", "SELECT ", column(field.TableName, "ID"), " FROM", safeMySQLNaming(field.TableName), "WHERE (")
			cond = append(cond, strings.Join(w, " OR "))
			cond = append(cond, ")", ")")
			where = append(where, strings.Join(cond, " "))
		}
		values = append(values, v...)

	}
	return where, values, nil
}

func column(tableName string, columnName string) string {
	return safeMySQLNaming(tableName) + "." + safeMySQLNaming(columnName)
}

func safeMySQLNaming(data string) string {
	return "`" + data + "`"
}

var qapiFields sync.Map = sync.Map{}

type pair struct {
	Field     reflect.StructField
	Typ       reflect.Type
	Tag       string
	TableName string
}

func getQapiFields(structType reflect.Type) *[]pair {
	fields, cached := qapiFields.Load(structType.Name())
	if !cached {
		// load and parse tags
		taggedFields := []pair{}
		for i := 0; i < structType.NumField(); i++ {
			field := structType.Field(i)
			tag, hasTag := getQapiQPrefix(&field)
			fieldTyp := field.Type
			val := reflect.New(fieldTyp)
			realType, tableName := queryapi.GetTableName(val.Interface())
			if hasTag {
				taggedFields = append(taggedFields, pair{
					Field:     field,
					Typ:       realType,
					Tag:       tag,
					TableName: tableName,
				})
			}
		}
		qapiFields.Store(structType.Name(), &taggedFields)
		return &taggedFields
	}
	return fields.(*[]pair)
}
func getQapiQPrefix(f *reflect.StructField) (string, bool) {
	// get gorm tag
	if tag, ok := f.Tag.Lookup("qapi"); ok {
		props := strings.Split(tag, ";")
		// find qaip q info
		for _, prop := range props {
			if strings.HasPrefix(prop, "q:") {
				return strings.TrimPrefix(prop, "q:"), true
			}
		}
	}
	return "", false
}

// Add this function to handle one-to-many relationships properly
func handleTranslatedOneToManyFilter(db *gorm.DB, query *qapi.Query, entityType reflect.Type) *gorm.DB {
	for _, v := range query.Filter {
		if strings.Contains(v.Name, ".") {
			parts := strings.SplitN(v.Name, ".", 2)
			relation := parts[0]
			fieldName := parts[1]

			field, found := entityType.FieldByName(relation)
			if !found {
				continue
			}

			// Check if this is a one-to-many relationship
			if field.Type.Kind() == reflect.Slice ||
				(field.Type.Kind() == reflect.Ptr && field.Type.Elem().Kind() == reflect.Slice) {
				// Get the related table name
				relatedType := getRelatedModelType(field.Type)
				relatedTable := relatedType.Name()

				// Build a subquery that finds parent IDs where children match the filter
				db = db.Where(fmt.Sprintf("ID IN (SELECT %sID FROM %s WHERE %s = ?)",
					entityType.Name(), relatedTable, fieldName), v.Value)
			}
		}
	}
	return db
}
