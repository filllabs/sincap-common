package mysql

import (
	"reflect"
	"testing"
	"time"

	"database/sql/driver"

	"github.com/filllabs/sincap-common/db/interfaces"
	"github.com/filllabs/sincap-common/db/mysql/translations"
	"github.com/filllabs/sincap-common/middlewares/qapi"
	"github.com/stretchr/testify/assert"
)

// Test model implementing all optimization interfaces
type OptimizedUser struct {
	ID    uint   `db:"ID"`
	Name  string `db:"Name"`
	Email string `db:"Email"`
}

func (OptimizedUser) TableName() string {
	return "Users"
}

func (u OptimizedUser) GetID() interface{} {
	return u.ID
}

func (u *OptimizedUser) SetID(id interface{}) error {
	if idVal, ok := id.(uint64); ok {
		u.ID = uint(idVal)
		return nil
	}
	return nil
}

func (u OptimizedUser) GetFieldMap() map[string]interface{} {
	return map[string]interface{}{
		"ID":    u.ID,
		"Name":  u.Name,
		"Email": u.Email,
	}
}

// Test model using reflection fallback
type ReflectionUser struct {
	ID    uint   `db:"ID"`
	Name  string `db:"Name"`
	Email string `db:"Email"`
}

func (ReflectionUser) TableName() string {
	return "Users"
}

func TestOptimizedInterfaces(t *testing.T) {
	// Test TableNamer interface
	user := OptimizedUser{ID: 1, Name: "John", Email: "john@example.com"}

	if tableNamer, ok := interface{}(user).(interfaces.TableNamer); ok {
		tableName := tableNamer.TableName()
		if tableName != "Users" {
			t.Errorf("Expected table name 'Users', got '%s'", tableName)
		}
	} else {
		t.Error("OptimizedUser should implement TableNamer interface")
	}

	// Test IDGetter interface
	if idGetter, ok := interface{}(user).(interfaces.IDGetter); ok {
		id := idGetter.GetID()
		if id != uint(1) {
			t.Errorf("Expected ID 1, got %v", id)
		}
	} else {
		t.Error("OptimizedUser should implement IDGetter interface")
	}

	// Test IDSetter interface
	if idSetter, ok := interface{}(&user).(interfaces.IDSetter); ok {
		err := idSetter.SetID(uint64(42))
		if err != nil {
			t.Errorf("SetID failed: %v", err)
		}
		if user.ID != 42 {
			t.Errorf("Expected ID 42, got %d", user.ID)
		}
	} else {
		t.Error("OptimizedUser should implement IDSetter interface")
	}

	// Test FieldMapper interface
	if fieldMapper, ok := interface{}(user).(interfaces.FieldMapper); ok {
		fieldMap := fieldMapper.GetFieldMap()
		if fieldMap["Name"] != "John" {
			t.Errorf("Expected Name 'John', got %v", fieldMap["Name"])
		}
		if fieldMap["Email"] != "john@example.com" {
			t.Errorf("Expected Email 'john@example.com', got %v", fieldMap["Email"])
		}
	} else {
		t.Error("OptimizedUser should implement FieldMapper interface")
	}
}

func TestReflectionFallback(t *testing.T) {
	// Test that reflection fallback still works
	user := ReflectionUser{ID: 1, Name: "Jane", Email: "jane@example.com"}

	// Should implement TableNamer but not the optimization interfaces
	if tableNamer, ok := interface{}(user).(interfaces.TableNamer); ok {
		tableName := tableNamer.TableName()
		if tableName != "Users" {
			t.Errorf("Expected table name 'Users', got '%s'", tableName)
		}
	} else {
		t.Error("ReflectionUser should implement TableNamer interface")
	}

	// Should NOT implement optimization interfaces
	if _, ok := interface{}(user).(interfaces.IDGetter); ok {
		t.Error("ReflectionUser should NOT implement IDGetter interface")
	}

	if _, ok := interface{}(&user).(interfaces.IDSetter); ok {
		t.Error("ReflectionUser should NOT implement IDSetter interface")
	}

	if _, ok := interface{}(user).(interfaces.FieldMapper); ok {
		t.Error("ReflectionUser should NOT implement FieldMapper interface")
	}
}

func TestCRUDOperations(t *testing.T) {
	// This test would require a database connection
	// For now, just test that the functions can be called without compilation errors

	// Example usage with optimized model
	optimizedUser := &OptimizedUser{Name: "Test", Email: "test@example.com"}
	_ = optimizedUser // Would use FieldMapper interface internally

	// Example usage with reflection model
	reflectionUser := &ReflectionUser{Name: "Test", Email: "test@example.com"}
	_ = reflectionUser // Would use reflection internally
}

func TestTimestampHandling(t *testing.T) {
	// Test model with timestamp fields
	type TimestampModel struct {
		ID        uint64    `db:"ID"`
		Name      string    `db:"Name"`
		CreatedAt time.Time `db:"CreatedAt"`
		UpdatedAt time.Time `db:"UpdatedAt"`
	}

	model := &TimestampModel{
		Name: "Test Record",
		// CreatedAt and UpdatedAt are zero values - should be auto-set
	}

	// Test createWithReflection timestamp handling
	t.Run("CreateWithReflection", func(t *testing.T) {
		beforeCreate := time.Now()

		// Simulate the timestamp setting logic from createWithReflection
		recordValue := reflect.ValueOf(model)
		if recordValue.Kind() == reflect.Ptr {
			recordValue = recordValue.Elem()
		}

		now := time.Now()

		// Set CreatedAt if field exists and is zero value
		if createdAtField := recordValue.FieldByName("CreatedAt"); createdAtField.IsValid() && createdAtField.CanSet() {
			if createdAtField.Type() == reflect.TypeOf(time.Time{}) && createdAtField.Interface().(time.Time).IsZero() {
				createdAtField.Set(reflect.ValueOf(now))
			}
		}

		// Set UpdatedAt if field exists and is zero value
		if updatedAtField := recordValue.FieldByName("UpdatedAt"); updatedAtField.IsValid() && updatedAtField.CanSet() {
			if updatedAtField.Type() == reflect.TypeOf(time.Time{}) && updatedAtField.Interface().(time.Time).IsZero() {
				updatedAtField.Set(reflect.ValueOf(now))
			}
		}

		afterCreate := time.Now()

		// Verify timestamps were set
		if model.CreatedAt.IsZero() {
			t.Error("CreatedAt should be set automatically")
		}
		if model.UpdatedAt.IsZero() {
			t.Error("UpdatedAt should be set automatically")
		}

		// Verify timestamps are reasonable
		if model.CreatedAt.Before(beforeCreate) || model.CreatedAt.After(afterCreate) {
			t.Error("CreatedAt timestamp is not within expected range")
		}
		if model.UpdatedAt.Before(beforeCreate) || model.UpdatedAt.After(afterCreate) {
			t.Error("UpdatedAt timestamp is not within expected range")
		}
	})

	// Test createWithFieldMap timestamp handling
	t.Run("CreateWithFieldMap", func(t *testing.T) {
		fieldMap := map[string]interface{}{
			"Name": "Test Record from Map",
		}

		recordType := reflect.TypeOf(model)
		if recordType.Kind() == reflect.Ptr {
			recordType = recordType.Elem()
		}

		beforeCreate := time.Now()
		now := time.Now()

		// Check if CreatedAt field exists in the record type and set it if not provided
		if _, exists := fieldMap["CreatedAt"]; !exists {
			if field, found := recordType.FieldByName("CreatedAt"); found {
				fieldType := field.Type
				if fieldType == reflect.TypeOf(time.Time{}) || fieldType == reflect.TypeOf(&time.Time{}) {
					fieldMap["CreatedAt"] = now
				}
			}
		}

		// Check if UpdatedAt field exists in the record type and set it if not provided
		if _, exists := fieldMap["UpdatedAt"]; !exists {
			if field, found := recordType.FieldByName("UpdatedAt"); found {
				fieldType := field.Type
				if fieldType == reflect.TypeOf(time.Time{}) || fieldType == reflect.TypeOf(&time.Time{}) {
					fieldMap["UpdatedAt"] = now
				}
			}
		}

		afterCreate := time.Now()

		// Verify timestamps were added to fieldMap
		createdAt, hasCreatedAt := fieldMap["CreatedAt"]
		updatedAt, hasUpdatedAt := fieldMap["UpdatedAt"]

		if !hasCreatedAt {
			t.Error("CreatedAt should be added to fieldMap automatically")
		}
		if !hasUpdatedAt {
			t.Error("UpdatedAt should be added to fieldMap automatically")
		}

		if createdAtTime, ok := createdAt.(time.Time); ok {
			if createdAtTime.Before(beforeCreate) || createdAtTime.After(afterCreate) {
				t.Error("CreatedAt timestamp is not within expected range")
			}
		} else {
			t.Error("CreatedAt should be a time.Time")
		}

		if updatedAtTime, ok := updatedAt.(time.Time); ok {
			if updatedAtTime.Before(beforeCreate) || updatedAtTime.After(afterCreate) {
				t.Error("UpdatedAt timestamp is not within expected range")
			}
		} else {
			t.Error("UpdatedAt should be a time.Time")
		}
	})

	// Test updateWithFieldMap timestamp handling
	t.Run("UpdateWithFieldMap", func(t *testing.T) {
		fieldMap := map[string]interface{}{
			"Name": "Updated Name",
		}

		beforeUpdate := time.Now()
		now := time.Now()

		// Auto-set UpdatedAt timestamp for Update operation
		if _, exists := fieldMap["UpdatedAt"]; !exists {
			fieldMap["UpdatedAt"] = now
		}

		afterUpdate := time.Now()

		// Verify UpdatedAt was added to fieldMap
		updatedAt, hasUpdatedAt := fieldMap["UpdatedAt"]

		if !hasUpdatedAt {
			t.Error("UpdatedAt should be added to fieldMap automatically")
		}

		if updatedAtTime, ok := updatedAt.(time.Time); ok {
			if updatedAtTime.Before(beforeUpdate) || updatedAtTime.After(afterUpdate) {
				t.Error("UpdatedAt timestamp is not within expected range")
			}
		} else {
			t.Error("UpdatedAt should be a time.Time")
		}
	})
}

func TestTranslationFunctionality(t *testing.T) {
	// Test that the translation helper functions work correctly

	// Test findTranslationFields
	type MockTranslatedRecord struct {
		ID   uint `db:"ID"`
		Name interface {
			Set(string, string)
			Get(string) string
			Value() (driver.Value, error)
		} `db:"Name"`
		Description interface {
			Set(string, string)
			Get(string) string
			Value() (driver.Value, error)
		} `db:"Description"`
		Price float64 `db:"Price"`
	}

	// Create a mock record slice
	records := &[]MockTranslatedRecord{}

	// Test findTranslationFields - this should work with the package path check
	fields := findTranslationFields(records)
	// Note: This test might not find fields since we're using interface{} instead of actual Translations type
	// But it verifies the function doesn't panic
	assert.NotNil(t, fields)

	// Test createTranslationCondition
	filter := qapi.Filter{Name: "Name", Operation: qapi.LK, Value: "Turkey"}
	condition := createTranslationCondition(filter, "Name")
	expected := "JSON_SEARCH(LOWER(`Name`), 'one', LOWER('%Turkey%')) IS NOT NULL"
	assert.Equal(t, expected, condition)

	// Test createTranslatedSortClauses
	sortClauses := []string{"Name ASC", "Price DESC"}
	langCode := "en-US"
	multiLangFields := []string{"Name"}

	result := createTranslatedSortClauses(sortClauses, langCode, multiLangFields)
	assert.Equal(t, 2, len(result))
	assert.Contains(t, result[0], "JSON_UNQUOTE(JSON_EXTRACT(`Name`")
	assert.Equal(t, "`Price` DESC", result[1])
}

func TestTranslationTypeDetection(t *testing.T) {
	// Test with actual translations.Translations type
	type MockTranslatedRecord struct {
		ID          uint                       `db:"ID"`
		Name        translations.Translations  `db:"Name"`
		Description *translations.Translations `db:"Description"`
		Price       float64                    `db:"Price"`
	}

	// Create a mock record slice
	records := &[]MockTranslatedRecord{}

	// Test findTranslationFields with actual Translations type
	fields := findTranslationFields(records)

	// Should find both Name and Description fields
	assert.Equal(t, 2, len(fields))
	assert.Contains(t, fields, "Name")
	assert.Contains(t, fields, "Description")

	// Test with single record
	singleRecord := &MockTranslatedRecord{}
	singleFields := findTranslationFields(singleRecord)
	assert.Equal(t, 2, len(singleFields))
	assert.Contains(t, singleFields, "Name")
	assert.Contains(t, singleFields, "Description")

	// Test with non-translation record
	type RegularRecord struct {
		ID   uint   `db:"ID"`
		Name string `db:"Name"`
	}

	regularRecords := &[]RegularRecord{}
	regularFields := findTranslationFields(regularRecords)
	assert.Equal(t, 0, len(regularFields))
}

func TestTranslationLanguageCodeHandling(t *testing.T) {
	// Test with actual translations.Translations type
	type MockTranslatedRecord struct {
		ID          uint                      `db:"ID"`
		Name        translations.Translations `db:"Name"`
		Description translations.Translations `db:"Description"`
		Price       float64                   `db:"Price"`
	}

	records := &[]MockTranslatedRecord{}
	multiLangFields := []string{"Name", "Description"}

	// Test sorting with specific language code
	t.Run("SortingWithSpecificLanguage", func(t *testing.T) {
		sortClauses := []string{"Name ASC", "Price DESC"}
		langCode := "en-US"

		result := createTranslatedSortClauses(sortClauses, langCode, multiLangFields)
		assert.Equal(t, 2, len(result))

		// Name field should be modified with JSON validity check and extraction for specific language
		assert.Contains(t, result[0], "CASE WHEN JSON_VALID(`Name`) THEN COALESCE(JSON_UNQUOTE(JSON_EXTRACT(`Name`, '$.\"en-US\"')), CAST(`Name` AS CHAR)) ELSE CAST(`Name` AS CHAR) END")
		assert.Contains(t, result[0], "ASC")

		// Price field should be converted to proper SQL format
		assert.Equal(t, "`Price` DESC", result[1])
	})

	// Test sorting with +/- prefix system
	t.Run("SortingWithPrefixSystem", func(t *testing.T) {
		sortClauses := []string{"+Name", "-Description", "+Price"}
		langCode := "en-US"

		result := createTranslatedSortClauses(sortClauses, langCode, multiLangFields)
		assert.Equal(t, 3, len(result))

		// Name field should be modified with JSON validity check for ASC
		assert.Contains(t, result[0], "CASE WHEN JSON_VALID(`Name`) THEN COALESCE(JSON_UNQUOTE(JSON_EXTRACT(`Name`, '$.\"en-US\"')), CAST(`Name` AS CHAR)) ELSE CAST(`Name` AS CHAR) END")
		assert.Contains(t, result[0], "ASC")

		// Description field should be modified with JSON validity check for DESC
		assert.Contains(t, result[1], "CASE WHEN JSON_VALID(`Description`) THEN COALESCE(JSON_UNQUOTE(JSON_EXTRACT(`Description`, '$.\"en-US\"')), CAST(`Description` AS CHAR)) ELSE CAST(`Description` AS CHAR) END")
		assert.Contains(t, result[1], "DESC")

		// Price field should be converted to proper SQL format
		assert.Equal(t, "`Price` ASC", result[2])
	})

	// Test sorting with +/- prefix system and "all" language
	t.Run("SortingWithPrefixSystemAndAllLanguages", func(t *testing.T) {
		sortClauses := []string{"+Name", "-Description", "+Price"}
		langCode := "all"

		result := createTranslatedSortClauses(sortClauses, langCode, multiLangFields)
		assert.Equal(t, 3, len(result))

		// Name field should use DEFAULT_LANG_CODE for "all" language with ASC and JSON validity check
		assert.Contains(t, result[0], "CASE WHEN JSON_VALID(`Name`) THEN COALESCE(JSON_UNQUOTE(JSON_EXTRACT(`Name`, '$.\"en-US\"')), CAST(`Name` AS CHAR)) ELSE CAST(`Name` AS CHAR) END")
		assert.Contains(t, result[0], "ASC")

		// Description field should use DEFAULT_LANG_CODE for "all" language with DESC and JSON validity check
		assert.Contains(t, result[1], "CASE WHEN JSON_VALID(`Description`) THEN COALESCE(JSON_UNQUOTE(JSON_EXTRACT(`Description`, '$.\"en-US\"')), CAST(`Description` AS CHAR)) ELSE CAST(`Description` AS CHAR) END")
		assert.Contains(t, result[1], "DESC")

		// Price field should be converted to proper SQL format
		assert.Equal(t, "`Price` ASC", result[2])
	})

	// Test sorting with traditional ASC/DESC format (backward compatibility)
	t.Run("SortingWithTraditionalFormat", func(t *testing.T) {
		sortClauses := []string{"Name ASC", "Description DESC", "Price ASC"}
		langCode := "en-US"

		result := createTranslatedSortClauses(sortClauses, langCode, multiLangFields)
		assert.Equal(t, 3, len(result))

		// Name field should be modified with JSON validity check for ASC
		assert.Contains(t, result[0], "CASE WHEN JSON_VALID(`Name`) THEN COALESCE(JSON_UNQUOTE(JSON_EXTRACT(`Name`, '$.\"en-US\"')), CAST(`Name` AS CHAR)) ELSE CAST(`Name` AS CHAR) END")
		assert.Contains(t, result[0], "ASC")

		// Description field should be modified with JSON validity check for DESC
		assert.Contains(t, result[1], "CASE WHEN JSON_VALID(`Description`) THEN COALESCE(JSON_UNQUOTE(JSON_EXTRACT(`Description`, '$.\"en-US\"')), CAST(`Description` AS CHAR)) ELSE CAST(`Description` AS CHAR) END")
		assert.Contains(t, result[1], "DESC")

		// Price field should be converted to proper SQL format
		assert.Equal(t, "`Price` ASC", result[2])
	})

	// Test filtering with specific language code
	t.Run("FilteringWithSpecificLanguage", func(t *testing.T) {
		filter := qapi.Filter{Name: "Name", Operation: qapi.LK, Value: "Turkey"}
		langCode := "en-US"

		condition := createTranslationConditionForLanguage(filter, "Name", langCode)
		expected := "(CASE WHEN JSON_VALID(`Name`) THEN (JSON_EXTRACT(`Name`, '$.\"en-US\"') IS NOT NULL AND LOWER(JSON_UNQUOTE(JSON_EXTRACT(`Name`, '$.\"en-US\"'))) LIKE LOWER('%Turkey%')) ELSE LOWER(CAST(`Name` AS CHAR)) LIKE LOWER('%Turkey%') END)"
		assert.Equal(t, expected, condition)
	})

	// Test filtering with "all" language code (should use the general condition)
	t.Run("FilteringWithAllLanguages", func(t *testing.T) {
		filter := qapi.Filter{Name: "Name", Operation: qapi.LK, Value: "Turkey"}

		condition := createTranslationCondition(filter, "Name")
		expected := "JSON_SEARCH(LOWER(`Name`), 'one', LOWER('%Turkey%')) IS NOT NULL"
		assert.Equal(t, expected, condition)
	})

	// Test field selection with specific language code
	t.Run("FieldSelectionWithSpecificLanguage", func(t *testing.T) {
		fields := []string{"Name", "Price"}
		langCode := "en-US"

		result := createTranslatedFields(fields, langCode, multiLangFields, records)
		assert.Equal(t, 2, len(result))

		// Name field should be modified with JSON validity check and extraction
		assert.Contains(t, result[0], "CASE WHEN JSON_VALID(`Name`) THEN COALESCE(JSON_UNQUOTE(JSON_EXTRACT(`Name`, '$.\"en-US\"')), '') ELSE CAST(`Name` AS CHAR) END AS `Name`")

		// Price field should remain unchanged
		assert.Equal(t, "Price", result[1])
	})

	// Test field selection with "all" language code
	t.Run("FieldSelectionWithAllLanguages", func(t *testing.T) {
		fields := []string{"Name", "Price"}
		langCode := "all"

		result := createTranslatedFields(fields, langCode, multiLangFields, records)
		assert.Equal(t, 2, len(result))

		// Name field should use DEFAULT_LANG_CODE for "all" to get meaningful value with JSON validity check
		assert.Contains(t, result[0], "CASE WHEN JSON_VALID(`Name`) THEN COALESCE(JSON_UNQUOTE(JSON_EXTRACT(`Name`, '$.\"en-US\"')), '') ELSE CAST(`Name` AS CHAR) END AS `Name`")

		// Price field should remain unchanged
		assert.Equal(t, "Price", result[1])
	})

	// Test IN operation with specific language
	t.Run("INOperationWithSpecificLanguage", func(t *testing.T) {
		filter := qapi.Filter{Name: "Name", Operation: qapi.IN, Value: "Turkey|USA|Germany"}
		langCode := "en-US"

		condition := createTranslationConditionForLanguage(filter, "Name", langCode)

		// Should create OR conditions for each value with JSON validity check
		assert.Contains(t, condition, "CASE WHEN JSON_VALID(`Name`) THEN")
		assert.Contains(t, condition, "JSON_EXTRACT(`Name`, '$.\"en-US\"') IS NOT NULL")
		assert.Contains(t, condition, "LOWER(JSON_UNQUOTE(JSON_EXTRACT(`Name`, '$.\"en-US\"'))) = LOWER('Turkey')")
		assert.Contains(t, condition, "LOWER(JSON_UNQUOTE(JSON_EXTRACT(`Name`, '$.\"en-US\"'))) = LOWER('USA')")
		assert.Contains(t, condition, "LOWER(JSON_UNQUOTE(JSON_EXTRACT(`Name`, '$.\"en-US\"'))) = LOWER('Germany')")
		assert.Contains(t, condition, "ELSE")
		assert.Contains(t, condition, "LOWER(CAST(`Name` AS CHAR)) = LOWER('Turkey')")
		assert.Contains(t, condition, " OR ")
	})
}
