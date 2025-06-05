// Package services provides an interface for CRUD operations and an implementation for GORM
package services

import (
	"context"

	"github.com/filllabs/sincap-common/middlewares/qapi"
)

// Service defines the standard CRUD operations for database interactions
type Service[E any] interface {
	// List retrieves a collection of records based on the query parameters
	// Returns the total count of records and any error encountered
	List(ctx context.Context, record *[]E, query *qapi.Query, lang ...string) (int, error)

	// Read retrieves a single record by its ID
	// Accepts optional preload parameters for eager loading related data
	Read(ctx context.Context, record *E, id any, preloads ...string) error

	// Create inserts a new record into the database
	Create(ctx context.Context, record *E) error

	// Update modifies an existing record
	// Optional fieldParams can be provided for partial updates
	Update(ctx context.Context, record *E, fieldParams ...map[string]any) error

	// Delete removes one or more records from the database
	// When ids are provided, performs bulk deletion
	Delete(ctx context.Context, record *E, ids ...any) error
}

// HasList checks if the service implements List
type HasList[E any] interface {
	List(ctx context.Context, record *[]E, query *qapi.Query, lang ...string) (int, error)
}

// HasRead checks if the service implements Read
type HasRead[E any] interface {
	Read(ctx context.Context, record *E, id any, preloads ...string) error
}

// HasCreate checks if the service implements Create
type HasCreate[E any] interface {
	Create(ctx context.Context, record *E) error
}

// HasUpdate checks if the service implements Update
type HasUpdate[E any] interface {
	Update(ctx context.Context, record *E, fieldParams ...map[string]any) error
}

// HasDelete checks if the service implements Delete
type HasDelete[E any] interface {
	Delete(ctx context.Context, record *E, ids ...any) error
}
