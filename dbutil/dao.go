package dbutil

import (
	"reflect"
	"sincap-common/dbconn"
	"sincap-common/logging"
	"sincap-common/resources/query"

	"github.com/fatih/structs"
	"github.com/jinzhu/gorm"
	"go.uber.org/zap"
)

// List calls ListByQuery or ListAll according to the query parameter
func List(DB *gorm.DB, typ interface{}, query *query.Query, preloads ...string) (interface{}, int, error) {
	if query == nil {
		return ListAll(DB, typ, preloads)
	}
	return ListByQuery(DB, typ, query, preloads)
}

// ListByQuery returns all records matches with the Query API
func ListByQuery(DB *gorm.DB, typ interface{}, query *query.Query, preloads []string) (interface{}, int, error) {
	tableName := ""
	if tableNameFunc, useMethod := reflect.TypeOf(typ).MethodByName("TableName"); useMethod {
		a := reflect.ValueOf(typ)
		values := tableNameFunc.Func.Call([]reflect.Value{a})
		tableName = values[0].Interface().(string)
	} else {
		tableName = reflect.TypeOf(typ).Name()
	}
	slice := reflect.New(reflect.SliceOf(reflect.TypeOf(typ)))
	records := slice.Interface()

	//TODO: Don't call GenerateDB 2 times for count
	db := GenerateDB(query, DB, typ).Table(tableName)
	db = addPreloads(db, preloads)
	result := db.Find(records)
	if result.Error != nil {
		return nil, 0, result.Error
	}
	recordArr := reflect.ValueOf(records).Elem()
	query.Limit = -1
	query.Offset = -1

	count := -1
	GenerateDB(query, DB, typ).Table(tableName).Count(&count)

	var filteredList []interface{}
	if len(query.Fields) == 0 {
		for i := 0; i < recordArr.Len(); i++ {
			entity := recordArr.Index(i).Interface()
			json := structs.Map(entity)
			filteredList = append(filteredList, json)
		}
		return filteredList, count, result.Error
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
	return filteredList, count, result.Error
}

// ListAll returns all records
func ListAll(DB *gorm.DB, typ interface{}, preloads []string) (interface{}, int, error) {
	tableName := reflect.TypeOf(typ).Name()
	slice := reflect.New(reflect.SliceOf(reflect.TypeOf(typ)))
	records := slice.Interface()
	result := addPreloads(DB, preloads).Table(tableName).Find(records)
	recordArr := reflect.ValueOf(records).Elem()

	return recordArr.Interface(), recordArr.Len(), result.Error
}

// Create Record
func Create(DB *gorm.DB, record interface{}) error {
	result := DB.Create(record)
	if result.Error != nil {
		logging.Logger.Error("Create error", zap.Any("Model", reflect.TypeOf(record)), zap.Error(result.Error), zap.Any("record", record))
	}
	return result.Error
}

// Read Record
func Read(DB *gorm.DB, record interface{}, id uint) error {
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

// Delete Record
func Delete(DB *gorm.DB, record interface{}) error {
	result := DB.Delete(record)
	if result.Error != nil {
		logging.Logger.Error("Delete error", zap.Any("Model", reflect.TypeOf(record)), zap.Error(result.Error), zap.Any("record", record))
	}
	return result.Error
}

// DB Returns default DB connection
func DB() *gorm.DB {
	return dbconn.GetDefault()
}

// Preload opens "auto_preload" for the given DB
func Preload(DB *gorm.DB) *gorm.DB {
	return DB.Set("gorm:auto_preload", true)
}

// Associations opens "save_associations" for the given DB
func Associations(DB *gorm.DB) *gorm.DB {
	return DB.Set("gorm:association_autoupdate", true).Set("gorm:association_autocreate", true)
}

// AddPreloads helps you to add preloads to the given DB
func AddPreloads(db *gorm.DB, preloads ...string) *gorm.DB {
	return addPreloads(db, preloads)
}
func addPreloads(db *gorm.DB, preloads []string) *gorm.DB {
	for _, field := range preloads {
		db = db.Preload(field)
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
