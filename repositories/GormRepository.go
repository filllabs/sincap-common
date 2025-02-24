package repositories

import (
	"github.com/filllabs/sincap-common/db/mysql"
	"github.com/filllabs/sincap-common/middlewares/qapi"
	"gorm.io/gorm"
)

// GormRepository is a repository struct for GORM operations
type GormRepository struct {
}

// List retrieves a list of records based on the given query and preloads
func (rep *GormRepository) List(db *gorm.DB, records any, query *qapi.Query) (int, error) {
	return mysql.List(db, records, query)
}

// Read retrieves a single record by its ID with optional preloads
func (rep *GormRepository) Read(db *gorm.DB, record any, id any, preloads ...string) error {
	return mysql.Read(db, record, id, preloads...)
}

// Create inserts a new record into the database
func (rep *GormRepository) Create(db *gorm.DB, record any) error {
	return mysql.Create(db, record)
}

// Update handles both full and partial updates
func (rep *GormRepository) Update(db *gorm.DB, record any, fieldParams ...map[string]any) error {
	return mysql.Update(db, record, fieldParams...)
}

// Delete handles both single and bulk deletions
func (rep *GormRepository) Delete(db *gorm.DB, record any, ids ...any) error {
	if len(ids) == 0 {
		return mysql.Delete(db, record)
	}
	return mysql.DeleteAll(db, record, ids...)
}
