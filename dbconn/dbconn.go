package dbconn

import (
	"gitlab.com/sincap/sincap-common/logging"

	"github.com/jinzhu/gorm"
	"go.uber.org/zap"
	"moul.io/zapgorm"

	// for Driver support
	_ "github.com/jinzhu/gorm/dialects/mysql"
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
	gorm.AddNamingStrategy(&gorm.NamingStrategy{
		DB: func(name string) string {
			return name
		},
		Table: func(name string) string {
			return name
		},
		Column: func(name string) string {
			return name
		},
	})

	for i := range dbConfs {
		conf := dbConfs[i]
		args := make([]interface{}, len(conf.Args))
		for i, v := range conf.Args {
			args[i] = v
		}
		conn, err := gorm.Open(conf.Dialog, args...)

		if err != nil {
			logging.Logger.Fatal("DB Could not open connection.", zap.String("name", conf.Name), zap.Error(err))
		}
		if conf.LogMode {
			conn.LogMode(conf.LogMode)
			conn.SetLogger(zapgorm.New(logging.Logger))
		} else {
			conn.SetLogger(zapgorm.New(logging.Logger))
		}
		conn.SingularTable(true)
		db[conf.Name] = conn.Set("gorm:association_autoupdate", false).Set("gorm:association_autocreate", false)

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
