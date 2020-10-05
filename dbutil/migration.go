package dbutil

import (
	"reflect"

	"github.com/jinzhu/gorm"
	"gitlab.com/sincap/sincap-common/logging"
	"gitlab.com/sincap/sincap-common/reflection"
	"go.uber.org/zap"
)

// DropAll drops all tables at the database
func DropAll(DB *gorm.DB, models ...interface{}) {
	logging.Logger.Info("Dropping all tables")
	if db := DB.DropTableIfExists(models...); db.Error != nil {
		logging.Logger.Panic("Cannot drop tables", zap.Error(db.Error))
	}
	DropRelationTables(DB, models...)
}

// DropRelationTables drops all relation tables at the database
func DropRelationTables(DB *gorm.DB, models ...interface{}) {
	logging.Logger.Info("Dropping all relation tables")
	for _, model := range models {
		typ := reflection.ExtractRealType(reflect.TypeOf(model))
		for i := 0; i < typ.NumField(); i++ {
			f := typ.Field(i)
			if name, isM2M := GetMany2ManyTableName(&f); isM2M {
				if db := DB.DropTableIfExists(name); db.Error != nil {
					logging.Logger.Panic("Cannot drop table", zap.String("name", name), zap.Error(db.Error))
				}
				logging.Logger.Debug("Deleting Relation Table", zap.String("name", name))
			}
		}
	}
}
