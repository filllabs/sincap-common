// Package validator provides commonly needed valitator utilities and a validator instance with some implemented valitations (iban, phone, plate, etc.)
package validator

import (
	"regexp"

	validator "gopkg.in/go-playground/validator.v9"
)

// Validate is a common default validator instance for all validation operations.
var Validate = validator.New()

func init() {
	Validate.RegisterValidation("iban", iban)
	Validate.RegisterValidation("phone", phone)
	Validate.RegisterValidation("plate", plate)
}

var ibanRegexp = regexp.MustCompile("([A-Za-z]{2})([0-9]{24})")

func iban(fl validator.FieldLevel) bool {
	val := fl.Field().String()
	if len(val) == 0 {
		return true
	}
	return ibanRegexp.MatchString(val)
}

var phoneRegexp = regexp.MustCompile("[0-9]{7,20}")

func phone(fl validator.FieldLevel) bool {
	return phoneRegexp.MatchString(fl.Field().String())
}

var plateRegexp = regexp.MustCompile("([0-9]{2})([A-Z]{1,4})([0-9]{2,4})")

func plate(fl validator.FieldLevel) bool {
	return plateRegexp.MatchString(fl.Field().String())
}
