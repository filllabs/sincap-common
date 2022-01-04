package repositories

import (
	"gitlab.com/sincap/sincap-common/resources/query"
)

type Repository[E any] interface {
	List(record E, query *query.Query, preloads ...string) (interface{}, int, error)
	Create(record *E) error
	Read(record *E, id uint, preloads ...string) error
	Update(record *E) error
	UpdatePartial(table string, id uint, record map[string]interface{}) error
	Delete(record *E) error
}
