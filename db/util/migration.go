package util

import (
	"reflect"

	"gitlab.com/sincap/sincap-common/logging"
	"gitlab.com/sincap/sincap-common/reflection"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// DropAll drops all tables at the database
func DropAll(DB *gorm.DB, models ...interface{}) {
	logging.Logger.Info("Dropping all tables", zap.Int("count", len(models)), zap.Array("names", &logging.InterfaceArrayMarshaller{Arr: models}))
	if err := DB.Migrator().DropTable(models...); err != nil {
		logging.Logger.Panic("Cannot drop tables", zap.Error(err))
	}
	DropRelationTables(DB, models...)
}

// DropRelationTables drops all relation tables at the database
func DropRelationTables(DB *gorm.DB, models ...interface{}) {
	var tables []string
	for _, model := range models {
		typ := reflection.ExtractRealTypeField(reflect.TypeOf(model))
		for i := 0; i < typ.NumField(); i++ {
			f := typ.Field(i)
			if name, isM2M := GetMany2Many(&f); isM2M {
				tables = append(tables, name)
			}
		}
	}
	logging.Logger.Info("Dropping all relation tables", zap.Int("count", len(tables)), zap.Strings("names", tables))
	for _, name := range tables {
		if err := DB.Migrator().DropTable(name); err != nil {
			logging.Logger.Panic("Cannot drop tables", zap.Error(err))
		}
	}
}
