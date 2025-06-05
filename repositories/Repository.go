package repositories

import (
	"github.com/filllabs/sincap-common/middlewares/qapi"
	"gorm.io/gorm"
)

// Repository provides a generic CRUD interface for database operations.
type Repository interface {
	// List combines previous List and ListSmartSelect
	// If record is of type E, performs regular list, otherwise does smart select
	List(db *gorm.DB, record any, query *qapi.Query, lang ...string) (int, error)

	// Read combines previous Read and ReadSmartSelect
	// If record is of type *E, performs regular read, otherwise does smart select
	Read(db *gorm.DB, record any, id any, preloads ...string) error

	Create(db *gorm.DB, record any) error

	// Update handles both full and partial updates
	// For full updates: provide record of type *E and empty fields parameter
	// For partial updates: provide table name, id, and fields map
	Update(db *gorm.DB, record any, fieldParams ...map[string]any) error

	// Delete handles both single and bulk deletions
	// When ids is nil or empty, deletes the single record
	// When ids is provided, deletes all records with matching ids
	Delete(db *gorm.DB, record any, ids ...any) error
}
