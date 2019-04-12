package dbutil

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	"gitlab.com/sincap/sincap-common/resources/query"

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
		where, values, err := filter2Sql(q.Filter, typ)
		if err != nil {
			db.AddError(err)
			return db
		}
		db = db.Where(where, values...)
	}
	db = db.Offset(q.Offset)
	db = db.Limit(q.Limit)

	if len(q.Fields) > 0 {
		db = db.Select(q.Fields)
	}
	return db
}

func filter2Sql(filters []query.Filter, typ reflect.Type) (string, []interface{}, error) {
	var where []string
	var values []interface{}
	var targetField *reflect.StructField

	// Convert all filters to a where condition with AND
	for _, filter := range filters {
		var condition []string
		// var innerFieldFound = false

		// Get field name with split (it will make inner field queries possible)
		fieldNames := strings.Split(filter.Name, ".")

		// If it has more than 1 field name it has inner fields (another table)
		if len(fieldNames) > 1 {
			if cond, f, err := generateQuery(fieldNames, 1, typ, filter); err == nil {
				condition = append(condition, cond)
				targetField = f
			} else {
				return "", values, err
			}

		} else {
			condition = getCondition(condition, filter.Name, filter.Value, filter.Operation)
			field, isFieldFound := typ.FieldByName(filter.Name)
			if !isFieldFound {
				return "", values, fmt.Errorf("Can't find field for %s", filter.Name)
			}
			targetField = &field

		}
		where = append(where, strings.Join(condition, " "))
		kind := extractRealType(targetField.Type).Kind()
		if filter.Operation == query.IN {
			inVals := strings.Split(filter.Value, "|")
			for i := 0; i < len(inVals); i++ {
				var err error
				values, err = convertValue(filter, typ, kind, values, inVals[i])
				if err != nil {
					return "", values, err
				}
			}
		} else {
			var err error
			values, err = convertValue(filter, typ, kind, values, filter.Value)
			if err != nil {
				return "", values, err
			}
		}

	}
	return strings.Join(where, " AND "), values, nil
}

func generateQuery(fieldNames []string, i int, structType reflect.Type, filter query.Filter) (string, *reflect.StructField, error) {

	var condition []string

	fieldName := fieldNames[i-1]
	innerFieldName := fieldNames[i]

	if structType.Kind() != reflect.Struct {
		return "", nil, fmt.Errorf("%s is not struct", structType.Name())
	}

	field, isFieldFound := structType.FieldByName(fieldName)
	if !isFieldFound {
		return "", nil, fmt.Errorf("Can't find struct: %s field: %s", structType.Name(), filter.Name)
	}
	ft := extractRealType(field.Type)
	if ft.Kind() != reflect.Struct && ft.Kind() != reflect.Slice {
		return "", nil, fmt.Errorf("%s is not struct field in %s", filter.Name, structType.Name())
	}

	if ft.Kind() == reflect.Slice {
		ft = extractRealType(ft.Elem())
	}
	innerField, isInnerFieldFound := extractRealType(ft).FieldByName(innerFieldName)
	if !isInnerFieldFound {
		return "", nil, fmt.Errorf("Can't find struct: %s inner field: %s", extractRealType(ft).Name(), filter.Name)
	}

	innerCond := ""
	var targetField *reflect.StructField
	var innerErr error
	// first dive into inner fields
	if i < len(fieldNames)-1 {
		innerCond, targetField, innerErr = generateQuery(fieldNames, i+1, extractRealType(ft), filter)
		if innerErr != nil {
			return "", targetField, innerErr
		}
	} else {
		targetField = &innerField
	}

	table := extractRealType(ft).Name()

	if prefix, isPoly := getPolymorphicPrefix(&field); isPoly {
		polyID := prefix + "ID"
		polyType := prefix + "Type"
		outerTable := structType.Name()
		condition = append(condition, "ID", "IN (", "SELECT", polyID, "FROM", table, "WHERE (")
		if len(innerCond) > 0 {
			condition = append(condition, innerCond)
		} else {
			condition = getCondition(condition, innerFieldName, filter.Value, filter.Operation)
		}
		condition = append(condition, "AND", polyID, "=", "`"+outerTable+"`.ID", "AND", polyType, "=", "'"+outerTable+"'", ")", ")")
	} else if m2mTable, isM2M := getMany2Many(&field); isM2M {
		srcRef := structType.Name() + "_ID"
		destRef := table + "_ID"
		condition = append(condition, "ID", "IN (", "SELECT", srcRef, "FROM", m2mTable, "WHERE (", destRef, "IN (", "SELECT ID FROM", table, "WHERE (")
		condition = getCondition(condition, innerFieldName, filter.Value, filter.Operation)
		condition = append(condition, ")", ")", ")", ")")
	} else {
		condition = append(condition, fieldName+"ID", "IN (", "SELECT ID FROM", table, "WHERE (")
		if len(innerCond) > 0 {
			condition = append(condition, innerCond)
		} else {
			condition = getCondition(condition, innerFieldName, filter.Value, filter.Operation)
		}
		condition = append(condition, ")", ")")
	}
	return strings.Join(condition, " "), targetField, nil
}

func getPolymorphicPrefix(f *reflect.StructField) (string, bool) {
	// get gorm tag
	if tag, ok := f.Tag.Lookup("gorm"); ok {
		props := strings.Split(tag, ";")
		// find polymorphic info
		for _, prop := range props {
			if strings.HasPrefix(prop, "polymorphic:") {
				return strings.TrimPrefix(prop, "polymorphic:"), true
			}
		}
	}
	return "", false
}
func getMany2Many(f *reflect.StructField) (string, bool) {
	// get gorm tag
	if tag, ok := f.Tag.Lookup("gorm"); ok {
		props := strings.Split(tag, ";")
		// find polymorphic info
		for _, prop := range props {
			if strings.HasPrefix(prop, "many2many:") {
				return strings.TrimPrefix(prop, "many2many:"), true
			}
		}
	}
	return "", false
}

func extractRealType(field reflect.Type) reflect.Type {
	if field.Kind() == reflect.Ptr {
		return field.Elem()
	}
	return field
}

func convertValue(filter query.Filter, typ reflect.Type, kind reflect.Kind, values []interface{}, value interface{}) ([]interface{}, error) {
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
			return nil, fmt.Errorf("QApi cannot parse date: %s for %s. Cause: %v", value.(string), filter.Name, err)
		}
		values = append(values, time.Unix(0, i*int64(time.Millisecond)))
	default:
		return nil, fmt.Errorf("Field type not supported for QApi %s : %s", typ.Name(), filter.Name)
	}
	return values, nil
}
