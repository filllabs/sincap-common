package services

import (
	"context"

	"github.com/filllabs/sincap-common/middlewares/qapi"
	"github.com/filllabs/sincap-common/repositories"
	"github.com/jmoiron/sqlx"
)

// SqlxService implements Service interface using sqlx
type SqlxService struct {
	dbCtxKey   string
	repository repositories.SqlxRepository
}

// NewSqlxService creates a new instance of SqlxService
func NewSqlxService(dbCtxKey string) SqlxService {
	return SqlxService{
		dbCtxKey:   dbCtxKey,
		repository: repositories.SqlxRepository{},
	}
}

// List retrieves a collection of records based on the query parameters
func (s *SqlxService) List(ctx context.Context, record any, query *qapi.Query, lang ...string) (int, error) {
	db := ctx.Value(s.dbCtxKey).(*sqlx.DB)
	return s.repository.List(db, record, query, lang...)
}

// Read retrieves a single record by its ID
func (s *SqlxService) Read(ctx context.Context, record any, id any, preloads ...string) error {
	db := ctx.Value(s.dbCtxKey).(*sqlx.DB)
	return s.repository.Read(db, record, id, preloads...)
}

// Create inserts a new record into the database
func (s *SqlxService) Create(ctx context.Context, record any) error {
	db := ctx.Value(s.dbCtxKey).(*sqlx.DB)
	return s.repository.Create(db, record)
}

// Update modifies an existing record
func (s *SqlxService) Update(ctx context.Context, record any, fieldParams ...map[string]any) error {
	db := ctx.Value(s.dbCtxKey).(*sqlx.DB)
	return s.repository.Update(db, record, fieldParams...)
}

// Delete removes one or more records from the database
func (s *SqlxService) Delete(ctx context.Context, record any, ids ...any) error {
	db := ctx.Value(s.dbCtxKey).(*sqlx.DB)
	return s.repository.Delete(db, record, ids...)
}
