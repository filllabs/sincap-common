package dbutil

import (
	"reflect"

	"gitlab.com/sincap/sincap-common/dbconn"
	"gitlab.com/sincap/sincap-common/logging"
	"gitlab.com/sincap/sincap-common/resources/query"

	"github.com/fatih/structs"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// ListSmartSelect calls ListByQuery or ListAll according to the query parameter with smart select support
func ListSmartSelect(DB *gorm.DB, typ interface{}, query *query.Query, styp interface{}, preloads ...string) (interface{}, int, error) {
	if query == nil {
		return ListAllSmartSelect(DB, typ, styp, preloads)
	}
	return ListByQuery(DB, typ, styp, query, preloads)
}

// List calls ListByQuery or ListAll according to the query parameter
func List(DB *gorm.DB, typ interface{}, query *query.Query, preloads ...string) (interface{}, int, error) {
	if query == nil {
		return ListAllSmartSelect(DB, typ, typ, preloads)
	}
	return ListByQuery(DB, typ, typ, query, preloads)
}

// ListByQuery returns all records matches with the Query API
func ListByQuery(DB *gorm.DB, typ interface{}, styp interface{}, query *query.Query, preloads []string) (interface{}, int, error) {
	tableName := ""
	if tableNameFunc, useMethod := reflect.TypeOf(typ).MethodByName("TableName"); useMethod {
		a := reflect.ValueOf(typ)
		values := tableNameFunc.Func.Call([]reflect.Value{a})
		tableName = values[0].Interface().(string)
	} else {
		tableName = reflect.TypeOf(typ).Name()
	}
	slice := reflect.New(reflect.SliceOf(reflect.TypeOf(styp)))
	records := slice.Interface()

	// Get count
	var count int64 = -1
	db := GenerateDB(query, DB, typ).Table(tableName)

	eTyp := reflect.TypeOf(typ)
	cDB := db

	cDB = cDB.Count(&count)
	if cDB.Error != nil {
		return make([]interface{}, 0, 0), 0, cDB.Error
	}

	// Add Offset and limit than select
	db = db.Offset(query.Offset)
	db = db.Limit(query.Limit)
	db = addPreloads(eTyp, db, preloads)
	result := db.Find(records)
	if result.Error != nil {
		return make([]interface{}, 0, 0), 0, result.Error
	}
	recordArr := reflect.ValueOf(records).Elem()

	filteredList := make([]interface{}, 0, count)
	if len(query.Fields) == 0 {
		for i := 0; i < recordArr.Len(); i++ {
			// since the is no fileds user entity instead of map
			entity := recordArr.Index(i).Interface()
			filteredList = append(filteredList, entity)
		}
		return filteredList, int(count), result.Error
	}
	for i := 0; i < recordArr.Len(); i++ {
		entity := recordArr.Index(i).Interface()
		json := structs.Map(entity)
		filtered := make(map[string]interface{}, len(query.Fields))
		for i := range query.Fields {
			filtered[query.Fields[i]] = json[query.Fields[i]]
		}
		filteredList = append(filteredList, filtered)
	}
	return filteredList, int(count), result.Error
}

// ListAllSmartSelect returns all records
func ListAllSmartSelect(DB *gorm.DB, typ interface{}, styp interface{}, preloads []string) (interface{}, int, error) {
	eTyp := reflect.TypeOf(typ)
	tableName := eTyp.Name()
	slice := reflect.New(reflect.SliceOf(reflect.TypeOf(styp)))
	records := slice.Interface()
	result := addPreloads(eTyp, DB, preloads).Table(tableName).Find(records)
	recordArr := reflect.ValueOf(records).Elem()

	return recordArr.Interface(), recordArr.Len(), result.Error
}

// ListAll returns all records
func ListAll(DB *gorm.DB, typ interface{}, preloads []string) (interface{}, int, error) {
	return ListAllSmartSelect(DB, typ, typ, preloads)
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
func Read(DB *gorm.DB, record interface{}, id uint, preloads ...string) error {
	if len(preloads) > 0 {
		DB = addPreloads(reflect.TypeOf(record), DB, preloads)
	}
	result := DB.First(record, id)
	if result.Error != nil {
		logging.Logger.Error("Read error", zap.Any("Model", reflect.TypeOf(record)), zap.Error(result.Error), zap.Uint("id", id))
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
func UpdatePartial(DB *gorm.DB, table string, id uint, record map[string]interface{}) error {
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

// DB Returns default DB connection clone
func DB() *gorm.DB {
	return dbconn.GetDefault()
}

// Preload opens "auto_preload" for the given DB
func Preload(DB *gorm.DB) *gorm.DB {
	return DB.Set("gorm:auto_preload", true)
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
		fType, ok := typ.FieldByName(field)
		if !ok {
			db = db.Joins(field)
			continue
		}
		_, isM2m := GetMany2ManyTableName(&fType)
		if isM2m { // many2many does not support joins.
			db = db.Preload(field)
		} else {
			db = db.Joins(field)
		}
		// "JOIN emails ON emails.user_id = users.id AND emails.email = ?"
	}
	return db
}

// MultiCreate  creates multiple records consecutively. Stops on any error, returns the error.
// Don't forget to start TX and rollback on any error if you need transactions support.
func MultiCreate(DB *gorm.DB, records ...interface{}) error {
	for _, record := range records {
		if record == nil {
			continue
		}
		err := Create(DB, record)
		if err != nil {
			return err
		}
	}
	return nil
}

// MultiUpdate  updates multiple records consecutively. Stops on any error, returns the error.
// Don't forget to start TX and rollback on any error if you need transactions support.
func MultiUpdate(DB *gorm.DB, records ...interface{}) error {
	for _, record := range records {
		if record == nil {
			continue
		}
		err := Update(DB, record)
		if err != nil {
			return err
		}
	}
	return nil
}

// MultiDelete  deletes multiple records consecutively. Stops on any error, returns the error.
// Don't forget to start TX and rollback on any error if you need transactions support.
func MultiDelete(DB *gorm.DB, records ...interface{}) error {
	for _, record := range records {
		if isNil(record) {
			continue
		}
		err := Delete(DB, record)
		if err != nil {
			return err
		}
	}
	return nil
}

func isNil(record interface{}) bool {
	if record == nil {
		return true
	}

	var val = reflect.ValueOf(record)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
		if !val.IsValid() {
			return true
		}
	}
	if val.FieldByName("ID").Uint() == 0 {
		return true
	}
	return false
}
