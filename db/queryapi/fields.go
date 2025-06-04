package queryapi

import (
	"reflect"
	"strings"
	"sync"
)

var qapiFields sync.Map = sync.Map{}

type pair struct {
	Field     reflect.StructField
	Typ       reflect.Type
	Tag       string
	TableName string
}

func getQapiFields(structType reflect.Type) *[]pair {
	fields, cached := qapiFields.Load(structType.Name())
	if !cached {
		// load and parse tags
		taggedFields := []pair{}
		for i := 0; i < structType.NumField(); i++ {
			field := structType.Field(i)
			tag, hasTag := getQapiQPrefix(&field)
			fieldTyp := field.Type
			val := reflect.New(fieldTyp)
			realType, tableName := GetTableName(val.Interface())
			if hasTag {
				taggedFields = append(taggedFields, pair{
					Field:     field,
					Typ:       realType,
					Tag:       tag,
					TableName: tableName,
				})
			}
		}
		qapiFields.Store(structType.Name(), &taggedFields)
		return &taggedFields
	}
	return fields.(*[]pair)
}
func getQapiQPrefix(f *reflect.StructField) (string, bool) {
	if tag, ok := f.Tag.Lookup("qapi"); ok {
		props := strings.Split(tag, ";")
		// find qaip q info
		for _, prop := range props {
			if strings.HasPrefix(prop, "q:") {
				return strings.TrimPrefix(prop, "q:"), true
			}
		}
	}
	return "", false
}
