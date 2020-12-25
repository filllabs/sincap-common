package dbutil

import (
	"errors"
	"fmt"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"gitlab.com/sincap/sincap-common/dbconn"
)

func getMockDB() (*gorm.DB, error) {
	gorm.AddNamingStrategy(dbconn.AsIsNamingStrategy())

	db, mock, err := sqlmock.New()
	if err != nil {
		return nil, fmt.Errorf("an error '%s' was not expected when opening a stub database connection", err)
	}
	if err != nil {
		return nil, fmt.Errorf("Failed to open mock sql db, got error: %v", err)
	}

	if db == nil {
		return nil, errors.New("mock is null")
	}

	if mock == nil {
		return nil, errors.New("sqlmock is null")
	}

	gormDB, err := gorm.Open("mysql", db)
	if err != nil {
		return gormDB, err
	}
	gormDB.SingularTable(true)
	return gormDB, nil
}

func TestDropAll(t *testing.T) {
	// TODO: Attention!!! there is a problem with mock db fix it
	// db, err := getMockDB()
	// if err != nil {
	// 	t.Error(err)
	// 	return
	// }
	// db = db.AutoMigrate(Sample{}, SampleM2M{})
	// if db.Error != nil {
	// 	t.Error(db.Error)
	// 	return
	// }
	// defer db.Close()
	// DropAll(db, Sample{}, SampleM2M{})
}
