// Package dbconn gives utility methods for creating and using gorm connections
package dbconn

import (
	"gitlab.com/sincap/sincap-common/dbconn/zapgorm"
	"gitlab.com/sincap/sincap-common/logging"

	"go.uber.org/zap"
	"gorm.io/gorm"

	// for Driver support
	"gorm.io/driver/mysql"
)

// DB connection for all DB operatins
var db = map[string]*gorm.DB{}

// DBConfig holds database configuration
type DBConfig struct {
	Name        string   `json:"name"`
	Dialog      string   `json:"dialog"`
	Args        []string `json:"args"`
	LogMode     bool     `json:"logMode"`
	AutoMigrate []string `json:"autoMigrate"`
}

// Configure DB connection
func Configure(dbConfs []DBConfig) {

	for i := range dbConfs {
		conf := dbConfs[i]
		args := make([]interface{}, len(conf.Args))
		for i, v := range conf.Args {
			args[i] = v
		}
		conn, err := gorm.Open(mysql.Open(conf.Args[0]), &gorm.Config{
			NamingStrategy:                           AsIsNamingStrategy(),
			Logger:                                   zapgorm.New(logging.Logger),
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

// GetDefault returns the DB connection named "default".
func GetDefault() *gorm.DB {
	return Get("default")
}

// Get returns the DB connection with the given name.
func Get(name string) *gorm.DB {
	conn, ok := db[name]
	if !ok {
		logging.Logger.Error("DB is nil. Returning default", zap.String("name", name))
		conn, ok := db["default"]
		if !ok {
			logging.Logger.Fatal("DB Some fatal problems occured", zap.String("name", name))
		}
		return conn
	}
	return conn
}

// Close tries to close db connection returns if any error occures.
func Close(db *gorm.DB) error {
	sdb, err := db.DB()
	if err == nil {
		err = sdb.Close()
	}
	return err
}
