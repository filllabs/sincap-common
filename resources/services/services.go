// Package services provides a set of servic generator functions in order to prevent code repetation
package services

import (
	"net/http"

	"github.com/go-chi/chi"
	"gitlab.com/sincap/sincap-common/db"
	"gitlab.com/sincap/sincap-common/db/crud"
	"gitlab.com/sincap/sincap-common/json"
	"gitlab.com/sincap/sincap-common/resources"
	"gitlab.com/sincap/sincap-common/resources/contexts"
	"gitlab.com/sincap/sincap-common/resources/query"
	"gitlab.com/sincap/sincap-common/resources/responses"
	"gitlab.com/sincap/sincap-common/types"
	"gitlab.com/sincap/sincap-common/validator"
)

// NewCRUDRouter creates a predefined router for simple crud scenario
func NewCRUDRouter(res resources.Resource, typ interface{}) func(r chi.Router) {
	return func(r chi.Router) {
		pathParamCtx := contexts.PathParamID(res.PathParamCxtKey, typ)
		bodyCtx := validator.Context(res.BodyCtxKey, typ)

		r.With(query.Context).Get("/", NewList(typ))
		r.With(bodyCtx).Post("/", NewCreate(typ, res.BodyCtxKey))
		r.Route("/{id}", func(r chi.Router) {
			r.Use(pathParamCtx)
			r.Get("/", NewRead(typ, res.PathParamCxtKey))
			r.With(bodyCtx).Put("/", NewUpdate(typ, res.BodyCtxKey))
			r.Delete("/", NewDelete(typ, res.PathParamCxtKey))
		})
	}
}

// NewCRPUDRouter creates a predefined router for simple crud scenario with partial update
func NewCRPUDRouter(res resources.Resource, typ interface{}, table string) func(r chi.Router) {
	return func(r chi.Router) {
		pathParamCtx := contexts.PathParamID(res.PathParamCxtKey, typ)
		bodyCtx := validator.Context(res.BodyCtxKey, typ)

		r.With(query.Context).Get("/", NewList(typ))
		r.With(bodyCtx).Post("/", NewCreate(typ, res.BodyCtxKey))
		r.Route("/{id}", func(r chi.Router) {
			r.Use(pathParamCtx)
			r.Get("/", NewRead(typ, res.PathParamCxtKey))
			r.Patch("/", NewUpdatePartial(table))
			r.Delete("/", NewDelete(typ, res.PathParamCxtKey))
		})
	}
}

// NewList creates a new list function.
func NewList(in interface{}, preloads ...string) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		query, _ := r.Context().Value(query.QueryContextKey).(*query.Query)
		records, count, err := crud.List(db.DB(), in, query, preloads...)
		if err != nil {
			responses.Status500(w, r, err)
			return
		}
		responses.Data(w, r, records, count)
	}
}

// NewCreate creates a new list function.
func NewCreate(in interface{}, bodyCtxKey contexts.ContextKey) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		body := r.Context().Value(bodyCtxKey)
		if err := crud.Create(db.DB(), body); err != nil {
			responses.Status500(w, r, err)
			return
		}
		responses.Data(w, r, body)
	}
}

// NewRead creates a new list function.
func NewRead(in interface{}, pathParamCxtKey contexts.ContextKey) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		record := r.Context().Value(pathParamCxtKey)
		responses.Data(w, r, record)
	}
}

// NewUpdate creates a new update function.
func NewUpdate(in interface{}, bodyCtxKey contexts.ContextKey) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		body := r.Context().Value(bodyCtxKey)
		if err := crud.Update(db.DB(), body); err != nil {
			responses.Status500(w, r, err)
			return
		}
		responses.Data(w, r, body)
	}
}

// NewUpdatePartial creates a new partial update function.
func NewUpdatePartial(table string) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		idParam := chi.URLParam(r, "id")
		id, err := types.ParseUint(idParam)
		if err != nil {
			responses.Status404(w, r)
			return
		}
		values := make(map[string]interface{})

		if err := json.Decode(r.Body, &values); err != nil {
			responses.Err(w, r, err, http.StatusBadRequest)
			return
		}
		if err := crud.UpdatePartial(db.DB(), table, id, values); err != nil {
			responses.Status500(w, r, err)
			return
		}
		responses.Status204(w, r)
	}
}

// NewDelete creates a new delete function.
func NewDelete(in interface{}, pathParamCxtKey contexts.ContextKey) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		record := r.Context().Value(pathParamCxtKey)
		if err := crud.Delete(db.DB(), record); err != nil {
			responses.Status500(w, r, err)
			return
		}
		responses.Status204(w, r)
	}
}
