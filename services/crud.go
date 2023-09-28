package services

import (
	"context"

	"github.com/filllabs/sincap-common/middlewares/qapi"
	"github.com/filllabs/sincap-common/repositories"
)

type Service[E any] interface {
	List(ctx context.Context, query *qapi.Query, preloads ...string) (interface{}, int, error)
	ListSmartSelect(ctx context.Context, e any, query *qapi.Query, preloads ...string) (interface{}, int, error)
	Create(ctx context.Context, data *E) error
	Read(ctx context.Context, id uint) (*E, error)
	ReadSmartSelect(ctx context.Context, e any, id uint) (any, error)
	Update(ctx context.Context, table string, id uint, data map[string]interface{}) error
	Delete(ctx context.Context, id uint) (*E, error)
}

type CrudService[E any] struct {
	Repository repositories.Repository[E]
}

func NewService[E any](r repositories.Repository[E]) Service[E] {
	return &CrudService[E]{r}
}

func (ser *CrudService[E]) List(ctx context.Context, query *qapi.Query, preloads ...string) (interface{}, int, error) {
	e := new(E)
	list, count, err := ser.Repository.List(*e, query, preloads...)
	return list, count, ConvertErr(err)
}

func (ser *CrudService[E]) ListSmartSelect(ctx context.Context, e any, query *qapi.Query, preloads ...string) (interface{}, int, error) {
	list, count, err := ser.Repository.ListSmartSelect(e, query, preloads...)
	return list, count, ConvertErr(err)
}

func (ser *CrudService[E]) Create(ctx context.Context, u *E) error {
	return ser.Repository.Create(u)
}

func (ser *CrudService[E]) Read(ctx context.Context, uid uint) (*E, error) {
	e := new(E)
	if err := ser.Repository.Read(e, uid); err != nil {
		return nil, ConvertErr(err)
	}
	return e, nil
}
func (ser *CrudService[E]) ReadSmartSelect(ctx context.Context, e any, uid uint) (any, error) {
	if err := ser.Repository.ReadSmartSelect(e, uid); err != nil {
		return nil, ConvertErr(err)
	}
	return e, nil
}
func (ser *CrudService[E]) Update(ctx context.Context, table string, uid uint, data map[string]interface{}) error {
	if err := ser.Repository.UpdatePartial(table, uid, data); err != nil {
		return ConvertErr(err)
	}
	return nil
}
func (ser *CrudService[E]) Delete(ctx context.Context, uid uint) (*E, error) {
	e := new(E)
	if err := ser.Repository.Read(e, uid); err != nil {
		return nil, ConvertErr(err)
	}
	if err := ser.Repository.Delete(e); err != nil {
		return nil, ConvertErr(err)
	}
	return e, nil
}
