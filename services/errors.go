package services

import (
	"errors"

	"github.com/gofiber/fiber/v2/utils"
	"gorm.io/gorm"
)

type Error struct {
	Code    int     `json:"code"`
	Message *string `json:"message,omitempty"`
	Inner   error   `json:"inner,omitempty"`
}

func (e Error) Error() string {
	if e.Message != nil {
		return *e.Message
	}
	if e.Inner != nil {
		return e.Inner.Error()
	}
	return ""
}
func (e Error) Is(target error) bool {
	_, ok := target.(*Error)
	return ok
}

// NewError creates a new Error instance with an optional message
func NewError(code int, errs ...interface{}) *Error {
	e := &Error{
		Code: code,
	}
	if len(errs) > 0 {
		if msg, ok := errs[0].(string); ok {
			e.Message = &msg
		} else if err, ok := errs[0].(error); ok {
			e.Inner = err
		}
	} else {
		msg := utils.StatusMessage(code)
		e.Message = &msg
	}
	return e
}

func ConvertErr(err error) error {
	if err == nil {
		return nil
	}
	code := 500
	msg := err.Error()
	if errors.Is(err, gorm.ErrRecordNotFound) {
		code = 404
	} else {
		code = 500
	}
	e := &Error{
		Code:    code,
		Inner:   err,
		Message: &msg,
	}
	return e
}
