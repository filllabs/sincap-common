package dbutil

import (
	"errors"
	"fmt"
	"reflect"
	"gitlab.com/sincap/sincap-common/resources/query"
	"strconv"
	"strings"
	"time"

	"github.com/jinzhu/gorm"
)

var timeKind = reflect.TypeOf(time.Time{}).Kind()

func getCondition(condition []string, field string, value interface{}, operation query.Operation) []string {
	condition = append(condition, field)
	switch operation {
	case query.EQ:
		condition = append(condition, "=", "?")
	case query.NEQ:
		condition = append(condition, "<>", "?")
	case query.GT:
		condition = append(condition, ">", "?")
	case query.GTE:
		condition = append(condition, ">=", "?")
	case query.LT:
		condition = append(condition, "<", "?")
	case query.LTE:
		condition = append(condition, "<=", "?")
	case query.LK:
		condition = append(condition, "LIKE", "?")
	case query.IN:
		params := strings.Split(value.(string), "|")
		paramSection := strings.Repeat("?,", len(params))
		condition = append(condition, "IN", "("+paramSection[0:len(paramSection)-1]+")")
		value = params
	}
	return condition
}

// GenerateDB generates a valid db query from the given api Query
func GenerateDB(q *query.Query, db *gorm.DB, entity interface{}) *gorm.DB {
	typ := reflect.TypeOf(entity)

	//TODO: checkfieldnames with model
	if len(q.Sort) > 0 {
		db = db.Order(strings.Join(q.Sort, ", "))
	}
	if len(q.Filter) > 0 {
		var where []string
		var values []interface{}
		for i := range q.Filter {
			filter := q.Filter[i]
			var value interface{} = filter.Value
			var condition []string
			var fieldNames []string
			isStruct := strings.ContainsRune(filter.Name, '.')
			if isStruct {
				table := typ.Name()
				fieldNames = strings.Split(filter.Name, ".")
				field := fieldNames[1]
				f, ok := typ.FieldByName(fieldNames[0])
				innerTableType := f.Type.Name()
				if f.Type.Kind() == reflect.Ptr {
					innerTableType = f.Type.Elem().Name()
				}
				innerTable := "`" + innerTableType + "`"
				isPoly := ok && strings.Contains(f.Tag.Get("gorm"), "polymorphic:Holder")
				if isPoly {
					condition = append(condition, "ID", "= (", "SELECT HolderID FROM", innerTable, "WHERE (")
					condition = getCondition(condition, field, value, filter.Operation)
					condition = append(condition, "AND HolderID =", "`"+table+"`.ID", "AND HolderType =", "'"+table+"'", ")", ")")
				} else {
					condition = append(condition, fieldNames[0]+"ID", "IN (", "SELECT ID FROM", innerTable, "WHERE (")
					condition = getCondition(condition, field, value, filter.Operation)
					condition = append(condition, ")", ")")
				}
			} else {
				condition = getCondition(condition, filter.Name, value, filter.Operation)
			}

			var fieldType reflect.StructField
			var ok = false
			if isStruct {
				//Navigate names with range and find internal fieldType
				var structField reflect.StructField
				if len(fieldNames) > 0 {
					structField, ok = typ.FieldByName(fieldNames[0])
					len := len(fieldNames)
					for i := 1; i < len; i++ {
						if structField.Type.Kind() == reflect.Struct {
							structField, ok = structField.Type.FieldByName(fieldNames[i])
							if !ok {
								break
							}
						} else if structField.Type.Kind() == reflect.Ptr {
							if structField.Type.Elem().Kind() == reflect.Struct {
								structField, ok = structField.Type.Elem().FieldByName(fieldNames[i])
								if !ok {
									break
								}
							}
						}
					}
					fieldType = structField
				}
			} else {
				fieldType, ok = typ.FieldByName(filter.Name)
			}
			if !ok {
				db.AddError(errors.New("Can't find field for " + filter.Name))
				return db
			}
			where = append(where, strings.Join(condition, " "))
			kind := fieldType.Type.Kind()
			if kind == reflect.Ptr {
				kind = fieldType.Type.Elem().Kind()
			}
			if filter.Operation == query.IN {
				inVals := strings.Split(value.(string), "|")
				for i := 0; i < len(inVals); i++ {
					var err error
					values, err = convertValue(db, filter, typ, kind, values, inVals[i])
					if err != nil {
						return db
					}
				}
			} else {
				var err error
				values, err = convertValue(db, filter, typ, kind, values, value)
				if err != nil {
					return db
				}
			}

		}
		db = db.Where(strings.Join(where, " AND "), values...)
	}
	db = db.Offset(q.Offset)
	db = db.Limit(q.Limit)

	if len(q.Fields) > 0 {
		db = db.Select(q.Fields)
	}
	return db
}

func convertValue(db *gorm.DB, filter query.Filter, typ reflect.Type, kind reflect.Kind, values []interface{}, value interface{}) ([]interface{}, error) {
	switch kind {
	case reflect.String:
		values = append(values, value)
	case reflect.Int,
		reflect.Int8,
		reflect.Int16,
		reflect.Int32,
		reflect.Int64:
		if i, e := strconv.Atoi(value.(string)); e == nil {
			values = append(values, i)
		}
	case reflect.Uint,
		reflect.Uint8,
		reflect.Uint16,
		reflect.Uint32,
		reflect.Uint64:
		if i, e := strconv.ParseUint(value.(string), 10, 64); e == nil {
			values = append(values, i)
		}
	case reflect.Float32,
		reflect.Float64:
		if i, e := strconv.ParseFloat(value.(string), 64); e == nil {
			values = append(values, i)
		}
	case reflect.Bool:
		values = append(values, value.(string) == "true")
	case timeKind:
		i, err := strconv.ParseInt(value.(string), 10, 64)
		if err != nil {
			db.AddError(fmt.Errorf("QApi cannot parse date: %s for %s", value.(string), filter.Name))
			db.AddError(err)
			return nil, err
		}
		values = append(values, time.Unix(0, i*int64(time.Millisecond)))
	default:
		err := fmt.Errorf("Field type not supported for QApi %s : %s", typ.Name(), filter.Name)
		db.AddError(err)
		return nil, err
	}
	return values, nil
}
