package qapi

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFilterParse(t *testing.T) {
	cases := []struct {
		input       string
		filter      Filter
		errExpected bool
		err         error
	}{
		{input: "name=seray", filter: Filter{Name: "name", Operation: EQ, Value: "seray"}},
		{input: "name!=seray", filter: Filter{Name: "name", Operation: NEQ, Value: "seray"}},
		{input: "age<35", filter: Filter{Name: "age", Operation: LT, Value: "35"}},
		{input: "age<=35", filter: Filter{Name: "age", Operation: LTE, Value: "35"}},
		{input: "age>35", filter: Filter{Name: "age", Operation: GT, Value: "35"}},
		{input: "age>=35", filter: Filter{Name: "age", Operation: GTE, Value: "35"}},
		{input: "hobbies~=chess", filter: Filter{Name: "hobbies", Operation: LK, Value: "chess"}},
		{input: "hobbies|=chess", filter: Filter{Name: "hobbies", Operation: IN, Value: "chess"}},
		{input: "hobbies|=chess|go", filter: Filter{Name: "hobbies", Operation: IN, Value: "chess|go"}},
		{input: "hobbies*=chess*go", filter: Filter{Name: "hobbies", Operation: IN_ALT, Value: "chess*go"}},
		{input: "", err: ErrParamLength},
		{input: "asdsds", err: ErrInvalidOp},
		{input: "abc=", err: ErrMissingNameValue},
		{input: "=abc", err: ErrMissingNameValue},
	}

	for _, testCase := range cases {
		t.Run(testCase.input, func(t *testing.T) {
			filter := Filter{}
			err := filter.Parse(testCase.input)
			if testCase.err != nil {
				assert.Equal(t, err, testCase.err)
			} else {
				assert.NoError(t, err, testCase.input)
				assert.Equal(t, filter.Operation.String(), testCase.filter.Operation.String())
				assert.Equal(t, filter.Name, testCase.filter.Name)
				assert.Equal(t, filter.Value, testCase.filter.Value)
			}
		})
	}
}
