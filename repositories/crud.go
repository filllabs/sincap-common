package repositories

import (
	"github.com/filllabs/sincap-common/middlewares/qapi"
)

type Repository[E any] interface {
	List(record E, query *qapi.Query, preloads ...string) (interface{}, int, error)
	ListSmartSelect(record any, query *qapi.Query, preloads ...string) (interface{}, int, error)
	Create(record *E) error
	Read(record *E, id any, preloads ...string) error
	ReadSmartSelect(record any, id any, preloads ...string) error
	Update(record *E) error
	UpdatePartial(table string, id any, record map[string]interface{}) error
	Delete(record *E) error
	DeleteAll(record *E, ids []any) error
}
