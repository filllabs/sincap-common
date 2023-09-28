package qapi

import (
	"errors"
	"strings"
)

// Operation defines the type of the filter
type Operation int

// ErrInvalidOp is a default operator error
var ErrInvalidOp = errors.New("Invalid operator")

// ErrParamLength is a default filter error
var ErrParamLength = errors.New("Filter param can't be shorter than 3")

// ErrMissingNameValue is a default filter without any value
var ErrMissingNameValue = errors.New("Filter name or value can't be empty")

const (
	// EQ =
	EQ Operation = iota + 1
	// NEQ !=
	NEQ
	// LT <
	LT
	// LTE <=
	LTE
	// GT >
	GT
	// GTE >=
	GTE
	// LK ~=
	LK
	// IN |= (values must be separated with | )
	IN
	// IN_ALT *= (values must be separated with * )
	IN_ALT
)

func (op Operation) String() string {
	names := [...]string{
		"Unknown",
		"EQ",
		"NEQ",
		"LT",
		"LTE",
		"GT",
		"GTE",
		"LK",
		"IN",
		"IN_ALT",
	}
	return names[op]
}

// Filter holds necessary info for a filter
type Filter struct {
	Name      string
	Operation Operation
	Value     string
}

// Parse parses and fills the filter
func (filter *Filter) Parse(param string) error {
	param = strings.TrimSpace(param)
	if len(param) < 3 {
		return ErrParamLength
	}
	op := ""
outer:
	for _, ch := range param {
		switch ch {
		case '=', '!', '<', '>', '~', '|', '*':
			op = op + string(ch)
			if len(op) == 2 {
				break outer
			}
		default:
			if len(op) == 1 {
				break outer
			}
		}
	}

	switch op {
	case "=":
		filter.Operation = EQ
	case "!=":
		filter.Operation = NEQ
	case "<":
		filter.Operation = LT
	case "<=":
		filter.Operation = LTE
	case ">":
		filter.Operation = GT
	case ">=":
		filter.Operation = GTE
	case "~=":
		filter.Operation = LK
	case "|=":
		filter.Operation = IN
	case "*=":
		filter.Operation = IN_ALT
	default:
		return ErrInvalidOp
	}
	nameValue := strings.Split(param, op)
	filter.Name = strings.TrimSpace(nameValue[0])
	filter.Value = strings.TrimSpace(nameValue[1])
	if len(filter.Name) == 0 || len(filter.Value) == 0 {
		return ErrMissingNameValue
	}
	return nil
}
