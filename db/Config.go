package db

import (
	"gitlab.com/sincap/sincap-common/db/zapgorm"
	"gitlab.com/sincap/sincap-common/logging"
	"go.uber.org/zap"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// Config holds database configuration
type Config struct {
	Name        string   `json:"name"`
	Dialog      string   `json:"dialog"`
	Args        []string `json:"args"`
	LogMode     bool     `json:"logMode"`
	AutoMigrate []string `json:"autoMigrate"`
}

// Configure DB connection
func Configure(dbConfs []Config) {

	for i := range dbConfs {
		conf := dbConfs[i]
		args := make([]interface{}, len(conf.Args))
		for i, v := range conf.Args {
			args[i] = v
		}
		conn, err := gorm.Open(mysql.Open(conf.Args[0]), &gorm.Config{
			NamingStrategy:                           AsIsNamingStrategy(),
			Logger:                                   zapgorm.New(logging.Logger, conf.LogMode),
			DisableForeignKeyConstraintWhenMigrating: true,
			SkipDefaultTransaction:                   true,
		})

		if err != nil {
			logging.Logger.Fatal("DB Could not open connection.", zap.String("name", conf.Name), zap.Error(err))
		}
		DB := conn.Session(&gorm.Session{FullSaveAssociations: true})
		db[conf.Name] = DB
		logging.Logger.Info("DB initialized", zap.String("name", conf.Name))
	}
}
