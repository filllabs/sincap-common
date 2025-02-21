package services

import (
	"context"

	"github.com/filllabs/sincap-common/middlewares/qapi"
	"github.com/filllabs/sincap-common/repositories"
	"gorm.io/gorm"
)

// GormService implements Service interface using GORM
type GormService struct {
	dbCtxKey   string
	repository repositories.GormRepository
}

// NewGormService creates a new instance of GormService
func NewGormService(dbCtxKey string) GormService {
	return GormService{
		dbCtxKey:   dbCtxKey,
		repository: repositories.GormRepository{},
	}
}

// List retrieves a collection of records based on the query parameters
func (s *GormService) List(ctx context.Context, record any, query *qapi.Query) (int, error) {
	db := ctx.Value(s.dbCtxKey).(*gorm.DB)
	return s.repository.List(db, record, query)
}

// Read retrieves a single record by its ID
func (s *GormService) Read(ctx context.Context, record any, id any, preloads ...string) error {
	db := ctx.Value(s.dbCtxKey).(*gorm.DB)
	return s.repository.Read(db, record, id, preloads...)
}

// Create inserts a new record into the database
func (s *GormService) Create(ctx context.Context, record any) error {
	db := ctx.Value(s.dbCtxKey).(*gorm.DB)
	return s.repository.Create(db, record)
}

// Update modifies an existing record
func (s *GormService) Update(ctx context.Context, record any, fieldParams ...map[string]any) error {
	db := ctx.Value(s.dbCtxKey).(*gorm.DB)
	return s.repository.Update(db, record, fieldParams...)
}

// Delete removes one or more records from the database
func (s *GormService) Delete(ctx context.Context, record any, ids ...any) error {
	db := ctx.Value(s.dbCtxKey).(*gorm.DB)
	return s.repository.Delete(db, record, ids...)
}
