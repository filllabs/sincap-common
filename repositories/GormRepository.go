package repositories

import (
	"github.com/filllabs/sincap-common/db/crud"
	"github.com/filllabs/sincap-common/middlewares/qapi"
	"gorm.io/gorm"
)

type GormRepository[E any] struct {
	DB *gorm.DB
}

func NewGormRepository[E any](db *gorm.DB) GormRepository[E] {
	return GormRepository[E]{DB: db}
}
func (rep *GormRepository[E]) List(record E, query *qapi.Query, preloads ...string) (interface{}, int, error) {
	return crud.List(rep.DB, record, query, preloads...)
}
func (rep *GormRepository[E]) ListSmartSelect(record any, query *qapi.Query, preloads ...string) (interface{}, int, error) {
	return crud.List(rep.DB, record, query, preloads...)
}
func (rep *GormRepository[E]) Create(record *E) error {
	return crud.Create(rep.DB, record)
}
func (rep *GormRepository[E]) Read(record *E, id uint, preloads ...string) error {
	return crud.Read(rep.DB, record, id, preloads...)
}
func (rep *GormRepository[E]) ReadSmartSelect(record any, id uint, preloads ...string) error {
	return crud.Read(rep.DB, record, id, preloads...)
}
func (rep *GormRepository[E]) Update(record *E) error {
	return crud.Update(rep.DB, record)
}
func (rep *GormRepository[E]) UpdatePartial(table string, id uint, record map[string]interface{}) error {
	return crud.UpdatePartial(rep.DB, table, id, record)
}
func (rep *GormRepository[E]) Delete(record *E) error {
	return crud.Delete(rep.DB, record)
}

type TxAble interface {
	Begin(opts ...interface{}) interface{}
	Commit() interface{}
	Rollback() interface{}
}
