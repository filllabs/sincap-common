package mysql

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/filllabs/sincap-common/db/queryapi"
	"github.com/filllabs/sincap-common/db/types"
	"github.com/filllabs/sincap-common/db/util"
	"github.com/filllabs/sincap-common/logging"
	"github.com/filllabs/sincap-common/middlewares/qapi"
	"github.com/filllabs/sincap-common/reflection"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

// List calls ListByQuery or ListAll according to the query parameter
func List(DB *gorm.DB, records any, query *qapi.Query) (int, error) {
	value := reflect.ValueOf(records)
	if value.Kind() != reflect.Pointer {
		return 0, fmt.Errorf("records must be a pointer")
	}

	elem := value.Elem()
	if elem.Kind() != reflect.Slice {
		return 0, fmt.Errorf("records must be a pointer to slice")
	}

	entityType, tableName := queryapi.GetTableName(records)

	// CHECK: since entity used no need to manually add
	_, hasDeletedAt := entityType.FieldByName("DeletedAt")
	calculateCount := query.Offset > 0 || query.Limit > 0

	// Get count
	var count int64 = -1
	db, err := queryapi.GenerateDB(query, DB, records)
	if err != nil {
		return 0, err
	}

	db = db.Table(tableName)

	cDB := db
	// CHECK: since entity used no need to manually add
	if hasDeletedAt {
		cDB = cDB.Where("`" + tableName + "`.`DeletedAt` IS NULL")
	}

	// check if the count is needed as seperate query (if there is a pagination)
	if calculateCount {
		cDB = cDB.Count(&count)
		if cDB.Error != nil {
			return 0, cDB.Error
		}
		// Add Offset and limit
		db = db.Offset(query.Offset)
		db = db.Limit(query.Limit)
	}

	db = addPreloads(entityType, db, query.Preloads)
	if len(query.Fields) > 0 {
		db = db.Select(query.Fields)
	}
	result := db.Find(records)
	if result.Error != nil {
		return 0, result.Error
	}
	if !calculateCount {
		return elem.Len(), nil
	}
	return int(count), nil
}

// Create Record
func Create(DB *gorm.DB, record any) error {
	result := DB.Model(record).Create(record)
	if result.Error != nil {
		logging.Logger.Error("Create error", zap.Any("Model", reflect.TypeOf(record)), zap.Error(result.Error), zap.Any("record", record))
	}
	return result.Error
}

// Read Record
func Read(DB *gorm.DB, record any, id any, preloads ...string) error {
	if len(preloads) > 0 {
		DB = addPreloads(reflection.DepointerField(reflect.TypeOf(record)), DB, preloads)
	}

	// if id is string so add ID=? to query in order to support (gorm wants qull cond if id is string, if number it works by default)
	if _, ok := id.(string); ok {
		_, tableName := queryapi.GetTableName(record)
		id = fmt.Sprintf("`%s`.`ID`='%s'", tableName, id)
	}

	result := DB.First(record, id)
	if result.Error != nil {
		logging.Logger.Error("Read error", zap.Any("Model", reflect.TypeOf(record)), zap.Error(result.Error), zap.Any("id", id))
	}
	return result.Error
}

// Update Updates the record with the given fields
// If fields is empty, updates the record with all fields
// fields is a map of field names to values
// examples:
//
//		Update(DB, &User{Name: "John Doe"}) => Updates all users with name "John Doe"
//		Update(DB, &User{Name: "John Doe", Age: 30}) => Updates all users with name "John Doe" and age 30
//	  Update(DB, &User{ID:1 , Name: "John Doe"}) => Updates user with ID 1 and Name "John Doe"
//		Update(DB, &User{Name: "John Doe", Age: 30}, map[string]any{"Age": 41}) => Updates Age column to 41 for all users with Name "John Doe" and Age 30
//		Update(DB, &User{ID:1} , map[string]any{"Age": 41}) => Updates Age column to 41 for user with ID 1
func Update(DB *gorm.DB, model any, fieldsParams ...map[string]any) error {
	if len(fieldsParams) == 0 {
		// update full record
		result := DB.Save(model)
		if result.Error != nil {
			logging.Logger.Error("Update error", zap.Any("Model", reflect.TypeOf(model)), zap.Error(result.Error), zap.Any("record", model))
		}
		return result.Error
	}
	// error if fields more than 1 or first element is not a map
	if len(fieldsParams) > 1 || fieldsParams[0] == nil {
		return fmt.Errorf("update failed: invalid fields parameter")
	}

	fields := fieldsParams[0]
	// check fields and if inner map convert it to json
	for k, v := range fields {
		switch v.(type) {
		case map[string]any, []any:
			j := types.JSON{}
			j.Marshal(v)
			fields[k] = j
		}
	}
	result := DB.Model(model).Updates(fields)
	if result.Error != nil {
		logging.Logger.Error("Update error", zap.Any("Model", reflect.TypeOf(model)), zap.Error(result.Error), zap.Any("record", fields))
	}
	return result.Error
}

// Delete Record
func Delete(DB *gorm.DB, record any) error {
	result := DB.Delete(record)
	if result.Error != nil {
		logging.Logger.Error("Delete error", zap.Any("Model", reflect.TypeOf(record)), zap.Error(result.Error), zap.Any("record", record))
	}
	return result.Error
}

// DeleteAll Record
func DeleteAll(DB *gorm.DB, record any, ids ...any) error {
	result := DB.Delete(record, ids...)
	if result.Error != nil {
		logging.Logger.Error("Delete error", zap.Any("Model", reflect.TypeOf(record)), zap.Error(result.Error), zap.Any("record", record))
	}
	return result.Error
}

// Associations opens "save_associations" for the given DB
func Associations(DB *gorm.DB) *gorm.DB {
	return DB.Session(&gorm.Session{FullSaveAssociations: true})
}

// AddPreloads helps you to add preloads to the given DB
func AddPreloads(typ reflect.Type, db *gorm.DB, preloads ...string) *gorm.DB {
	return addPreloads(typ, db, preloads)
}
func addPreloads(typ reflect.Type, db *gorm.DB, preloads []string) *gorm.DB {
	for _, field := range preloads {
		isNested := strings.Contains(field, ".")
		if isNested {
			db = db.Preload(field)
			continue
		}
		fType, ok := typ.FieldByName(field)
		if !ok {
			db = db.Joins(field)
			continue
		}
		rFt := reflection.DepointerField(fType.Type)
		isSlice := rFt.Kind() == reflect.Slice

		if isSlice {
			db = db.Preload(field)
		} else if _, isM2m := util.GetMany2Many(&fType); isM2m { // many2many does not support joins.
			db = db.Preload(field)
		} else if _, isPoly := util.GetPolymorphic(&fType); isPoly { // polymorphic does not support joins.
			db = db.Preload(field)
		} else {
			db = db.Joins(field)
		}

		// "JOIN emails ON emails.user_id = users.id AND emails.email = ?"
	}
	return db
}
