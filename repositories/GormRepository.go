package repositories

import (
	"github.com/filllabs/sincap-common/db/crud"
	"github.com/filllabs/sincap-common/middlewares/qapi"
	"gorm.io/gorm"
)

// GormRepository is a generic repository struct for GORM operations
type GormRepository[E any] struct {
	DB *gorm.DB
}

// NewGormRepository creates a new instance of GormRepository
func NewGormRepository[E any](db *gorm.DB) GormRepository[E] {
	return GormRepository[E]{DB: db}
}

// List retrieves a list of records based on the given query and preloads
func (rep *GormRepository[E]) List(record E, query *qapi.Query, preloads ...string) (interface{}, int, error) {
	return ListWithDB(rep.DB, record, query, preloads...)
}

// ListSmartSelect retrieves a list of records with smart select functionality
func (rep *GormRepository[E]) ListSmartSelect(record any, query *qapi.Query, preloads ...string) (interface{}, int, error) {
	return ListSmartSelectWithDB(rep.DB, record, query, preloads...)
}

// Create inserts a new record into the database
func (rep *GormRepository[E]) Create(record *E) error {
	return CreateWithDB(rep.DB, record)
}

// Read retrieves a single record by its ID with optional preloads
func (rep *GormRepository[E]) Read(record *E, id any, preloads ...string) error {
	return ReadWithDB(rep.DB, record, id, preloads...)
}

// ReadSmartSelect retrieves a single record with smart select functionality
func (rep *GormRepository[E]) ReadSmartSelect(record any, id any, preloads ...string) error {
	return ReadSmartSelectWithDB(rep.DB, record, id, preloads...)
}

// Update modifies an existing record in the database
func (rep *GormRepository[E]) Update(record *E) error {
	return UpdateWithDB(rep.DB, record)
}

// UpdatePartial performs a partial update on a record using a map of fields
func (rep *GormRepository[E]) UpdatePartial(table string, id any, record map[string]interface{}) error {
	return UpdatePartialWithDB(rep.DB, table, id, record)
}

// Delete removes a record from the database
func (rep *GormRepository[E]) Delete(record *E) error {
	return DeleteWithDB(rep.DB, record)
}

// DeleteAll removes multiple records from the database based on their IDs
func (rep *GormRepository[E]) DeleteAll(record *E, ids []any) error {
	return DeleteAllWithDB(rep.DB, record, ids)
}

// ListWithDB retrieves a list of records based on the given query and preloads
func ListWithDB[E any](db *gorm.DB, record E, query *qapi.Query, preloads ...string) (interface{}, int, error) {
	return crud.List(db, record, query, preloads...)
}

// ListSmartSelectWithDB retrieves a list of records with smart select functionality
func ListSmartSelectWithDB(db *gorm.DB, record any, query *qapi.Query, preloads ...string) (interface{}, int, error) {
	return crud.List(db, record, query, preloads...)
}

// CreateWithDB inserts a new record into the database
func CreateWithDB[E any](db *gorm.DB, record *E) error {
	return crud.Create(db, record)
}

// ReadWithDB retrieves a single record by its ID with optional preloads
func ReadWithDB[E any](db *gorm.DB, record *E, id any, preloads ...string) error {
	return crud.Read(db, record, id, preloads...)
}

// ReadSmartSelectWithDB retrieves a single record with smart select functionality
func ReadSmartSelectWithDB(db *gorm.DB, record any, id any, preloads ...string) error {
	return crud.Read(db, record, id, preloads...)
}

// UpdateWithDB modifies an existing record in the database
func UpdateWithDB[E any](db *gorm.DB, record *E) error {
	return crud.Update(db, record)
}

// UpdatePartialWithDB performs a partial update on a record using a map of fields
func UpdatePartialWithDB(db *gorm.DB, table string, id any, record map[string]interface{}) error {
	return crud.UpdatePartial(db, table, id, record)
}

// DeleteWithDB removes a record from the database
func DeleteWithDB[E any](db *gorm.DB, record *E) error {
	return crud.Delete(db, record)
}

// DeleteAllWithDB removes multiple records from the database based on their IDs
func DeleteAllWithDB[E any](db *gorm.DB, record *E, ids []any) error {
	return crud.DeleteAll(db, record, ids)
}
