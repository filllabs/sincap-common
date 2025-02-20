package crud

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
func List(DB *gorm.DB, records []any, query *qapi.Query) ([]any, int, error) {

	entityType, tableName := queryapi.GetTableName(records)

	// CHECK: since entity used no need to manually add
	// _, hasDeletedAt := entityType.FieldByName("DeletedAt")
	calculateCount := query.Offset > 0 || query.Limit > 0

	// Get count
	var count int64 = -1
	db, err := queryapi.GenerateDB(query, DB, records)
	if err != nil {
		return records, 0, err
	}

	db = db.Table(tableName)

	cDB := db
	// CHECK: since entity used no need to manually add
	// if hasDeletedAt {
	// 	cDB.Where("`" + tableName + "`.`DeletedAt` is NULL")
	// }

	// check if the count is needed as seperate query (if there is a pagination)
	if calculateCount {
		cDB = cDB.Count(&count)
		if cDB.Error != nil {
			return records, 0, cDB.Error
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
		return records, 0, result.Error
	}
	if !calculateCount {
		return records, len(records), nil
	}
	return records, int(count), nil
}

// Create Record
func Create(DB *gorm.DB, record interface{}) error {
	result := DB.Model(record).Create(record)
	if result.Error != nil {
		logging.Logger.Error("Create error", zap.Any("Model", reflect.TypeOf(record)), zap.Error(result.Error), zap.Any("record", record))
	}
	return result.Error
}

// Read Record
func Read(DB *gorm.DB, record interface{}, id any, preloads ...string) error {
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

// Update Record
func Update(DB *gorm.DB, record interface{}) error {
	result := DB.Save(record)
	if result.Error != nil {
		logging.Logger.Error("Update error", zap.Any("Model", reflect.TypeOf(record)), zap.Error(result.Error), zap.Any("record", record))
	}
	return result.Error
}

// UpdatePartial Record
func UpdatePartial(DB *gorm.DB, table string, id any, record map[string]interface{}) error {
	// check fields and if inner map convert it to json
	for k, v := range record {
		switch v.(type) {
		case map[string]any, []any:
			j := types.JSON{}
			j.Marshal(v)
			record[k] = j
		}
	}
	result := DB.Table(table).Where("ID=?", id).Updates(record)
	if result.Error != nil {
		logging.Logger.Error("Update error", zap.Any("Model", reflect.TypeOf(record)), zap.Error(result.Error), zap.Any("record", record))
	}
	return result.Error
}

// Delete Record
func Delete(DB *gorm.DB, record interface{}) error {
	result := DB.Delete(record)
	if result.Error != nil {
		logging.Logger.Error("Delete error", zap.Any("Model", reflect.TypeOf(record)), zap.Error(result.Error), zap.Any("record", record))
	}
	return result.Error
}

// DeleteAll Record
func DeleteAll(DB *gorm.DB, record interface{}, ids []any) error {
	result := DB.Delete(record, ids)
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
