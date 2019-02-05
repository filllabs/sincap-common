package validator

import (
	"context"
	"net/http"
	"reflect"
	"regexp"
	"sincap-common/json"
	"sincap-common/resources"

	validator "gopkg.in/go-playground/validator.v9"
)

// Validate is a common default validator instance for all validation operations.
var Validate = validator.New()
var ibanRegexp = regexp.MustCompile("([A-Za-z]{2})([0-9]{24})")
var phoneRegexp = regexp.MustCompile("([+]?[0-9]{1,2})?([0-9]{3}[.-]?){2}[0-9]{4}")
var plateRegexp = regexp.MustCompile("([0-9]{2})([A-Z]{1,4})([0-9]{2,4})")

func init() {
	Validate.RegisterValidation("iban", iban)
	Validate.RegisterValidation("phone", phone)
	Validate.RegisterValidation("plate", plate)
}
func iban(fl validator.FieldLevel) bool {
	val := fl.Field().String()
	if len(val) == 0 {
		return true
	}
	return ibanRegexp.MatchString(val)
}
func phone(fl validator.FieldLevel) bool {
	return phoneRegexp.MatchString(fl.Field().String())
}
func plate(fl validator.FieldLevel) bool {
	return plateRegexp.MatchString(fl.Field().String())
}

// Context middleware is used to parse interfaces and validate them.
func Context(contextKey resources.ContextKey, in interface{}) func(next http.Handler) http.Handler {
	t := reflect.TypeOf(in)
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			record := reflect.New(t).Interface()
			if err := json.Decode(r.Body, record); err != nil {
				resources.ResponseErr(w, r, err, http.StatusBadRequest)
				return
			}
			if err := Validate.Struct(record); err != nil {
				resources.ResponseErr(w, r, err, http.StatusUnprocessableEntity)
				return
			}
			ctx := context.WithValue(r.Context(), contextKey, record)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
