package repositories

import (
	"gitlab.com/sincap/sincap-common/middlewares/qapi"
)

type Repository[E any] interface {
	List(record E, query *qapi.Query, preloads ...string) (interface{}, int, error)
	ListSmartSelect(record any, query *qapi.Query, preloads ...string) (interface{}, int, error)
	Create(record *E) error
	Read(record *E, id uint, preloads ...string) error
	ReadSmartSelect(record any, id uint, preloads ...string) error
	Update(record *E) error
	UpdatePartial(table string, id uint, record map[string]interface{}) error
	Delete(record *E) error
}
