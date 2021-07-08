// Package db gives utility methods for creating and using gorm connections. It supports multiple connections with different names.
// One connection is required with the name "default" it is primary connection and can be accessed via DB(), GetDefault() or Get("default").
// Others are accessible via Get("{name}"). A connection with name {name} must be defined at config file.
package db

import (
	mocket "github.com/selvatico/go-mocket"
	"gitlab.com/sincap/sincap-common/logging"
	"gorm.io/driver/mysql"
	"testing"

	"go.uber.org/zap"
	"gorm.io/gorm"
	// for Driver support
)

// DB connection for all DB operatins
var db = map[string]*gorm.DB{}

// DB Returns default DB connection clone
func DB() *gorm.DB {
	return GetDefault()
}

// GetDefault returns the DB connection named "default".
func GetDefault() *gorm.DB {
	return Get("default")
}

func SetupMockDBForTest(t *testing.T) *gorm.DB {
	var err error
	mocket.Catcher.Reset()
	mocket.Catcher.Register()
	mocket.Catcher.Logging = true

	dialect := mysql.New(mysql.Config{
		DSN:                       "mockdb",
		DriverName:                mocket.DriverName,
		SkipInitializeWithVersion: true,
	})

	mockDB, err := gorm.Open(dialect, new(gorm.Config))
	if err != nil {
		t.Fatalf("failed to open mock database connection: %s", err)
	}
	db["default"] = mockDB
	return mockDB
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

// CloseAll tries to close all db connections
func CloseAll() {
	for name, con := range db {
		sdb, err := con.DB()
		if err != nil {
			logging.Logger.Named("DB").Error("Can't get sql DB connection", zap.String("name", name), zap.Error(err))
		}
		err = sdb.Close()
		if err != nil {
			logging.Logger.Named("DB").Error("Can't close DB connection", zap.String("name", name), zap.Error(err))
		}
	}
	logging.Logger.Named("DB").Info("connections closed")
}
