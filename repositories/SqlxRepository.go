package repositories

import (
	"github.com/filllabs/sincap-common/db/mysql"
	"github.com/filllabs/sincap-common/middlewares/qapi"
	"github.com/jmoiron/sqlx"
)

// SqlxRepository is a repository struct for sqlx operations
type SqlxRepository struct {
}

// List retrieves a list of records based on the given query and preloads
func (rep *SqlxRepository) List(db *sqlx.DB, records any, query *qapi.Query, lang ...string) (int, error) {
	return mysql.List(db, records, query, lang...)
}

// Read retrieves a single record by its ID with optional preloads
func (rep *SqlxRepository) Read(db *sqlx.DB, record any, id any, preloads ...string) error {
	return mysql.Read(db, record, id, preloads...)
}

// Create inserts a new record into the database
func (rep *SqlxRepository) Create(db *sqlx.DB, record any) error {
	return mysql.Create(db, record)
}

// Update handles both full and partial updates
func (rep *SqlxRepository) Update(db *sqlx.DB, record any, fieldParams ...map[string]any) error {
	return mysql.Update(db, record, fieldParams...)
}

// Delete handles both single and bulk deletions
func (rep *SqlxRepository) Delete(db *sqlx.DB, record any, ids ...any) error {
	if len(ids) == 0 {
		return mysql.Delete(db, record)
	}
	return mysql.DeleteAll(db, record, ids...)
}
