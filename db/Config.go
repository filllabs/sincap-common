package db

import (
	"testing"

	mocket "github.com/selvatico/go-mocket"
	"gitlab.com/sincap/sincap-common/db/zapgorm"
	"gitlab.com/sincap/sincap-common/logging"
	"go.uber.org/zap"
	"gorm.io/driver/mysql"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
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

// ConfigureTestDB returns new mock db connection for test and override db instance with mock db connection.
func ConfigureTestDB(t *testing.T) (*gorm.DB, *mocket.MockCatcher) {
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
	return mockDB, mocket.Catcher
}

// ConfigureTestDB returns new mock db connection for test and override db instance with mock db connection.
func ConfigureMockDB(name string) *gorm.DB {
	mockDB, err := gorm.Open(sqlite.Open("file:"+name+"?mode=memory&cache=shared"),
		&gorm.Config{
			NamingStrategy: schema.NamingStrategy{
				SingularTable: true,
			},
		},
	)
	if err != nil {
		logging.Logger.Fatal("Cant create  mock sqlite db", zap.Error(err))
	}

	db["default"] = mockDB
	return mockDB
}
