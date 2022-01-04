// Package query helps to parse api query and
// includes Context for using it.
package query

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	"gitlab.com/sincap/sincap-common/resources/contexts"
	"gitlab.com/sincap/sincap-common/types"
)

// QueryContextKey is the key to access query object
var QueryContextKey = contexts.NewContextKey()

// Query holds parsed query params for the query
type Query struct {
	Q          string
	Fields     []string
	Preloads   []string
	Offset     int
	Limit      int
	Sort       []string
	Filter     []Filter
	TotalCount int
}

// Parse parses request query params and fills inside
func (query *Query) Parse(qParams map[string]string) error {
	isEmpty := true

	if q := qParams["_q"]; len(q) != 0 {
		query.Q = q
		isEmpty = false
	}

	if fieldParams := qParams["_fields"]; len(fieldParams) != 0 {
		fields := strings.Split(fieldParams, ",")
		query.Fields = fields
		isEmpty = false
	} else {
		query.Fields = make([]string, 0)
	}
	if preloadParams := qParams["_preloads"]; len(preloadParams) != 0 {
		preloads := strings.Split(preloadParams, ",")
		query.Preloads = preloads
		isEmpty = false
	} else {
		query.Preloads = make([]string, 0)
	}

	if offsetParam := qParams["_offset"]; len(offsetParam) != 0 {
		offset, errOffset := strconv.Atoi(offsetParam)
		if errOffset == nil {
			isEmpty = false
			query.Offset = offset
		} else {
			query.Offset = -1
		}
	} else {
		query.Offset = -1
	}

	if limitParam := qParams["_limit"]; len(limitParam) != 0 {
		limit, errLimit := strconv.Atoi(qParams["_limit"])
		if errLimit == nil {
			isEmpty = false
			query.Limit = limit
		} else {
			query.Limit = -1
		}
	} else {
		query.Limit = -1
	}

	if sortParam := qParams["_sort"]; len(sortParam) != 0 {
		isEmpty = false
		sorts := strings.Split(sortParam, ",")
		query.Sort = make([]string, len(sorts))
		for i, value := range sorts {
			sort := Sort{}
			sort.Parse(value)
			query.Sort[i] = sort.String()
		}
	} else {
		query.Sort = make([]string, 0)
	}
	if filterParam := qParams["_filter"]; len(filterParam) != 0 {
		isEmpty = false
		filters := strings.Split(filterParam, ",")
		query.Filter = make([]Filter, len(filters))
		for i, value := range filters {
			filter := Filter{}
			filter.Parse(value)
			query.Filter[i] = filter
		}
	} else {
		query.Filter = make([]Filter, 0)
	}
	if isEmpty {
		return errors.New("Query not found")
	}
	return nil
}

// Context parses the query params for the query
// Context parses the query params for the query
func Context(contextKey string) func(ctx *fiber.Ctx) error {
	return func(ctx *fiber.Ctx) error {
		query := Query{}
		params := map[string]string{}
		params["_q"] = ctx.Query("_q", "")
		params["_fields"] = ctx.Query("_fields", "")
		params["_preloads"] = ctx.Query("_preloads", "")
		params["_offset"] = ctx.Query("_offset", "")
		params["_limit"] = ctx.Query("_limit", "")
		params["_sort"] = ctx.Query("_sort", "")
		params["_filter"] = ctx.Query("_filter", "")

		if err := query.Parse(params); err == nil {
			ctx.Locals("queryapi", &query)
		}
		return ctx.Next()
	}
}

// ContextWithOwnerID and adds OwnerID filter
func ContextWithOwnerID(r *http.Request, ownerID uint) *Query {
	q, ok := r.Context().Value(QueryContextKey).(*Query)
	if !ok {
		q = &Query{Limit: -1, Offset: -1}
	}
	q.Filter = append(q.Filter, Filter{Name: "OwnerID", Operation: EQ, Value: types.FormatUint(ownerID)})
	return q
}
