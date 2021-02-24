package resource_utils

import (
	"net/http"

	"github.com/go-chi/chi"
	"gitlab.com/sincap/sincap-common/dbutil"
	"gitlab.com/sincap/sincap-common/json"
	"gitlab.com/sincap/sincap-common/resources"
	"gitlab.com/sincap/sincap-common/resources/query"
	"gitlab.com/sincap/sincap-common/types"
)

// NewList creates a new list function.
func NewList(in interface{}, preloads ...string) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		query, _ := r.Context().Value(query.QueryContextKey).(*query.Query)
		records, count, err := dbutil.List(dbutil.DB(), in, query, preloads...)
		if err != nil {
			resources.Response500(w, r, err)
			return
		}
		resources.ResponseData(w, r, records, count)
	}
}

// NewCreate creates a new list function.
func NewCreate(in interface{}, bodyCtxKey resources.ContextKey) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		body := r.Context().Value(bodyCtxKey)
		if err := dbutil.Create(dbutil.DB(), body); err != nil {
			resources.Response500(w, r, err)
			return
		}
		resources.ResponseData(w, r, body)
	}
}

// NewRead creates a new list function.
func NewRead(in interface{}, pathParamCxtKey resources.ContextKey) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		record := r.Context().Value(pathParamCxtKey)
		resources.ResponseData(w, r, record)
	}
}

// NewUpdate creates a new update function.
func NewUpdate(in interface{}, bodyCtxKey resources.ContextKey) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		body := r.Context().Value(bodyCtxKey)
		if err := dbutil.Update(dbutil.DB(), body); err != nil {
			resources.Response500(w, r, err)
			return
		}
		resources.ResponseData(w, r, body)
	}
}

// NewUpdatePartial creates a new partial update function.
func NewUpdatePartial(table string) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		idParam := chi.URLParam(r, "id")
		id, err := types.ParseUint(idParam)
		if err != nil {
			resources.Response404(w, r)
			return
		}
		values := make(map[string]interface{})

		if err := json.Decode(r.Body, &values); err != nil {
			resources.ResponseErr(w, r, err, http.StatusBadRequest)
			return
		}
		if err := dbutil.UpdatePartial(dbutil.DB(), table, id, values); err != nil {
			resources.Response500(w, r, err)
			return
		}
		resources.Response204(w, r)
	}
}

// NewDelete creates a new delete function.
func NewDelete(in interface{}, pathParamCxtKey resources.ContextKey) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		record := r.Context().Value(pathParamCxtKey)
		if err := dbutil.Delete(dbutil.DB(), record); err != nil {
			resources.Response500(w, r, err)
			return
		}
		resources.Response204(w, r)
	}
}
