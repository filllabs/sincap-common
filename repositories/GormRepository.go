package repositories

import (
	"github.com/filllabs/sincap-common/db/crud"
	"github.com/filllabs/sincap-common/middlewares/qapi"
	"gorm.io/gorm"
)

// GormRepository is a repository struct for GORM operations
type GormRepository struct {
}

// List retrieves a list of records based on the given query and preloads
func (rep *GormRepository) List(db *gorm.DB, records []any, query *qapi.Query) ([]any, int, error) {
	return crud.List(db, records, query)
}

// Read retrieves a single record by its ID with optional preloads
func (rep *GormRepository) Read(db *gorm.DB, record any, id any, preloads ...string) error {
	return crud.Read(db, record, id, preloads...)
}

// Create inserts a new record into the database
func (rep *GormRepository) Create(db *gorm.DB, record any) error {
	return crud.Create(db, record)
}

// Update handles both full and partial updates
func (rep *GormRepository) Update(db *gorm.DB, record any, id any, fields map[string]any) error {
	if len(fields) == 0 {
		return crud.Update(db, record)
	}
	// For partial updates, record parameter contains the table name as string
	tableName, ok := record.(string)
	if !ok {
		return crud.Update(db, record)
	}
	return crud.UpdatePartial(db, tableName, id, fields)
}

// Delete handles both single and bulk deletions
func (rep *GormRepository) Delete(db *gorm.DB, record any, ids ...any) error {
	if len(ids) == 0 {
		return crud.Delete(db, record)
	}
	return crud.DeleteAll(db, record, ids)
}
