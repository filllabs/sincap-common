package query

import (
	"context"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi"
	"github.com/stretchr/testify/assert"
)

func TestQuery(t *testing.T) {
	api := Query{}
	r := httptest.NewRequest("POST", "http://localhost:3000/api/menus?_q=nissan&_fields=manufacturer,model,id,color&_offset=10&_limit=5&_sort=-manufacturer,+model&_filter=name=seray,active!=true,order|=1|2", strings.NewReader("Read will return these bytes"))

	rctx := chi.NewRouteContext()

	r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))

	err := api.Parse(r)
	assert.NoError(t, err, "Parser error")

	assert.Equal(t, "nissan", api.Q, "Q test failed.")
	assert.Equal(t, []string{"manufacturer", "model", "id", "color"}, api.Fields, "Fields test failed.")
	assert.Equal(t, 10, api.Offset, "Offset test failed.")
	assert.Equal(t, 5, api.Limit, "Limit test failed.")
	assert.Equal(t, []string{"manufacturer desc", "model asc"},
		api.Sort, "Sort test failed.")
	assert.Equal(t, []Filter{
		{Name: "name", Operation: EQ, Value: "seray"},
		{Name: "active", Operation: NEQ, Value: "true"},
		{Name: "order", Operation: IN, Value: "1|2"},
	}, api.Filter, "Filter test failed.")
}
func TestQueryJustLimit(t *testing.T) {
	api := Query{}
	r := httptest.NewRequest("POST", "http://localhost:3000/api/menus?_offset=10&_limit=5", strings.NewReader("Read will return these bytes"))

	rctx := chi.NewRouteContext()

	r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))

	err := api.Parse(r)
	assert.NoError(t, err, "Parser error")

	assert.Equal(t, "", api.Q, "Q test failed.")
	assert.Equal(t, []string{}, api.Fields, "Fields test failed.")
	assert.Equal(t, 10, api.Offset, "Offset test failed.")
	assert.Equal(t, 5, api.Limit, "Limit test failed.")
	assert.Equal(t, []string{},
		api.Sort, "Sort test failed.")
	assert.Equal(t, []Filter{}, api.Filter, "Filter test failed.")
}
