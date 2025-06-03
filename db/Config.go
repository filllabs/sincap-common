package db

import (
	"testing"

	"github.com/filllabs/sincap-common/logging"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	mocket "github.com/selvatico/go-mocket"
	"go.uber.org/zap"
)

// Config holds database configuration
type Config struct {
	Name        string   `json:"name" yaml:"name"`
	Dialog      string   `json:"dialog" yaml:"dialog"`
	Args        []string `json:"args" yaml:"args"`
	LogMode     bool     `json:"logMode" yaml:"logMode"`
	AutoMigrate []string `json:"autoMigrate" yaml:"autoMigrate"`
}

// Configure DB connection
func Configure(dbConfs []Config) {

	for i := range dbConfs {
		conf := dbConfs[i]
		args := make([]interface{}, len(conf.Args))
		for i, v := range conf.Args {
			args[i] = v
		}
		conn, err := sqlx.Connect("mysql", conf.Args[0])

		if err != nil {
			logging.Logger.Fatal("DB Could not open connection.", zap.String("name", conf.Name), zap.Error(err))
		}

		// Set connection pool settings
		conn.SetMaxOpenConns(25)
		conn.SetMaxIdleConns(25)

		db[conf.Name] = conn
		logging.Logger.Info("DB initialized", zap.String("name", conf.Name))
	}
}

// ConfigureTestDB returns new mock db connection for test and override db instance with mock db connection.
func ConfigureTestDB(t *testing.T) (*sqlx.DB, *mocket.MockCatcher) {
	mocket.Catcher.Register()
	mocket.Catcher.Logging = true
	// GORM
	conn, err := sqlx.Connect(mocket.DriverName, "connection_string") // Can be any connection string
	if err != nil {
		t.Fatalf("Failed to connect to mock DB: %v", err)
	}

	catcher := mocket.Catcher.Reset()
	db["default"] = conn
	return conn, catcher
}

// ConfigureMockDB configures a mock database connection for testing
func ConfigureMockDB(name string) *sqlx.DB {
	mocket.Catcher.Register()
	mocket.Catcher.Logging = true

	conn, err := sqlx.Connect(mocket.DriverName, "connection_string")
	if err != nil {
		logging.Logger.Fatal("Failed to connect to mock DB", zap.String("name", name), zap.Error(err))
	}

	db[name] = conn
	return conn
}
