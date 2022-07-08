package queryapi

import (
	"fmt"
	"reflect"
	"strings"
	"time"

	"gitlab.com/sincap/sincap-common/db/types"
	"gitlab.com/sincap/sincap-common/db/util"
	"gitlab.com/sincap/sincap-common/logging"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"gitlab.com/sincap/sincap-common/middlewares/qapi"
	"gitlab.com/sincap/sincap-common/reflection"
)

var timeKind = reflect.TypeOf(time.Time{}).Kind()
var jsonType = reflect.TypeOf(types.JSON{})

func getCondition(condition []string, field string, value interface{}, operation qapi.Operation) []string {
	condition = append(condition, field)
	switch operation {
	case qapi.EQ:
		if isNull(value) {
			condition = append(condition, "IS", "NULL")
		} else {
			condition = append(condition, "=", "?")
		}
	case qapi.NEQ:
		if isNull(value) {
			condition = append(condition, "IS NOT", "NULL")
		} else {
			condition = append(condition, "<>", "?")
		}
	case qapi.GT:
		condition = append(condition, ">", "?")
	case qapi.GTE:
		condition = append(condition, ">=", "?")
	case qapi.LT:
		condition = append(condition, "<", "?")
	case qapi.LTE:
		condition = append(condition, "<=", "?")
	case qapi.LK:
		condition = append(condition, "LIKE", "?")
	case qapi.IN:
		params := strings.Split(value.(string), "|")
		paramSection := strings.Repeat("?,", len(params))
		condition = append(condition, "IN", "("+paramSection[0:len(paramSection)-1]+")")
		value = params
	case qapi.IN_ALT:
		params := strings.Split(value.(string), "*")
		paramSection := strings.Repeat("?,", len(params))
		condition = append(condition, "IN", "("+paramSection[0:len(paramSection)-1]+")")
		value = params
	}
	return condition
}

// GenerateDB generates a valid db query from the given api Query
func GenerateDB(q *qapi.Query, db *gorm.DB, entity interface{}) *gorm.DB {
	typ := reflect.TypeOf(entity)

	//TODO: checkfieldnames with model
	if len(q.Sort) > 0 {
		var sortFields []string
		for _, s := range q.Sort {
			values := strings.Split(s, " ")
			fieldNames := strings.Split(values[0], ".")
			field, isFieldFound := typ.FieldByName(fieldNames[0])
			if isFieldFound {
				dp := reflection.DepointerField(field.Type)
				if dp == jsonType {
					c := "CAST(" + typ.Name() + "." + fieldNames[0] + "->" + "'$." + fieldNames[1] + "'" + "AS CHAR) " + values[1]
					sortFields = append(sortFields, c)
				} else {
					sortFields = append(sortFields, s)
				}
			}

		}
		db = db.Order(strings.Join(sortFields, ", "))
	}
	if len(q.Filter) > 0 {
		where, values, err := filter2Sql(q.Filter, typ)
		if err != nil {
			db.AddError(err)
			return db
		}
		db = db.Where(where, values...)
	}

	if len(q.Fields) > 0 {
		db = db.Select(q.Fields)
	}

	if len(q.Q) > 0 {
		where, values, err := q2Sql(q.Q, typ)
		if err != nil {
			db.AddError(err)
			return db
		}
		db = db.Where(where, values...)
	}
	return db
}

func q2Sql(q string, typ reflect.Type) (string, []interface{}, error) {

	// Convert q to  where condition with OR for all fields with tag
	where, values, err := generateQQuery(typ, q)
	if err != nil {
		logging.Logger.Warn("Can't create query from q", zap.Error(err))
	}
	return strings.Join(where, " OR "), values, nil
}

func generateQQuery(structType reflect.Type, q string) ([]string, []interface{}, error) {
	var where []string
	var values []interface{}
	for i := 0; i < structType.NumField(); i++ {
		field := structType.Field(i)
		fieldTyp := reflection.ExtractRealTypeField(field.Type)
		tag, hasTag := getQapiQPrefix(&field)
		if hasTag {
			if fieldTyp.Kind() == reflect.Struct {
				w, v, err := generateQQuery(fieldTyp, q)
				var cond []string
				if err != nil {
					logging.Logger.Warn("Can't create query from q", zap.Error(err))
				}
				table := reflection.ExtractRealTypeField(fieldTyp).Name()
				if prefix, isPoly := util.GetPolymorphic(&field); isPoly {
					polyID := prefix + "ID"
					cond = append(cond, structType.Name()+".ID", "IN (", "SELECT", polyID, "FROM", table, "WHERE (")
					cond = append(cond, strings.Join(w, " OR "))
					cond = append(cond, ") )")
					where = append(where, strings.Join(cond, " "))
				} else if m2mTable, isM2M := util.GetMany2Many(&field); isM2M {
					srcRef := structType.Name() + "ID"
					destRef := table + "ID"
					cond = append(cond, structType.Name()+".ID", "IN (", "SELECT", srcRef, "FROM", m2mTable, "WHERE (", destRef, "IN (", "SELECT ID FROM", table, "WHERE (")
					cond = append(cond, strings.Join(w, " OR "))
					cond = append(cond, ")", ")", ")", ")")
					where = append(where, strings.Join(cond, " "))
				} else {
					cond = append(cond, field.Name+"ID", "IN (", "SELECT ID FROM", table, "WHERE (")
					cond = append(cond, strings.Join(w, " OR "))
					cond = append(cond, ")", ")")
					where = append(where, strings.Join(cond, " "))
				}
				values = append(values, v...)
			} else {
				where = append(where, field.Name+" LIKE ?")
				values = append(values, strings.Replace(tag, "*", q, 1))
			}
		}
	}
	return where, values, nil
}

func filter2Sql(filters []qapi.Filter, typ reflect.Type) (string, []interface{}, error) {
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
			// If it is json than handle different
			field, isFieldFound := typ.FieldByName(fieldNames[0])
			if !isFieldFound {
				return "", nil, fmt.Errorf("Can't find struct: %s field: %s", typ.Name(), filter.Name)
			}

			dp := reflection.DepointerField(field.Type)
			if dp == jsonType {
				// concat new value
				condition = getCondition(condition, typ.Name()+"."+fieldNames[0]+"->"+"'$."+fieldNames[1]+"'", filter.Value, qapi.LK)
				values = append(values, filter.Value)
				targetField = &field
			} else if cond, f, err := generateFilterQuery(fieldNames, 1, typ, filter); err == nil {
				condition = append(condition, cond)
				targetField = f
			} else {
				return "", values, err
			}

		} else {
			condition = getCondition(condition, "`"+typ.Name()+"`"+"."+"`"+filter.Name+"`", filter.Value, filter.Operation)
			field, isFieldFound := typ.FieldByName(filter.Name)
			if !isFieldFound {
				return "", values, fmt.Errorf("Can't find field for %s", filter.Name)
			}
			targetField = &field

		}
		where = append(where, strings.Join(condition, " "))
		kind := reflection.ExtractRealTypeField(targetField.Type).Kind()
		if filter.Operation == qapi.IN {
			inVals := strings.Split(filter.Value, "|")
			for i := 0; i < len(inVals); i++ {
				var err error
				values, err = util.ConvertValue(filter, typ, kind, values, inVals[i])
				if err != nil {
					return "", values, err
				}
			}
		} else if filter.Operation == qapi.IN_ALT {
			inVals := strings.Split(filter.Value, "*")
			for i := 0; i < len(inVals); i++ {
				var err error
				values, err = util.ConvertValue(filter, typ, kind, values, inVals[i])
				if err != nil {
					return "", values, err
				}
			}
		} else {
			var err error
			values, err = util.ConvertValue(filter, typ, kind, values, filter.Value)
			if err != nil {
				return "", values, err
			}
		}

	}
	return strings.Join(where, " AND "), values, nil
}

func generateFilterQuery(fieldNames []string, i int, structType reflect.Type, filter qapi.Filter) (string, *reflect.StructField, error) {

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
	ft := reflection.ExtractRealTypeField(field.Type)

	logging.Logger.Info("TYPES", zap.Any("org", field.Type), zap.Any("orgKind", field.Type.Kind()), zap.Any("real", ft), zap.Any("realKind", ft.Kind()))

	if ft.Kind() != reflect.Struct && ft.Kind() != reflect.Slice {
		return "", nil, fmt.Errorf("%s is not struct field in %s", filter.Name, structType.Name())
	}
	//TODO:check  maybe noo need previous extect handles it
	if ft.Kind() == reflect.Slice {
		ft = reflection.ExtractRealTypeField(ft.Elem())
	}
	innerField, isInnerFieldFound := reflection.ExtractRealTypeField(ft).FieldByName(innerFieldName)
	if !isInnerFieldFound {
		return "", nil, fmt.Errorf("Can't find struct: %s inner field: %s", reflection.ExtractRealTypeField(ft).Name(), filter.Name)
	}
	innerCond := ""
	var targetField *reflect.StructField
	var innerErr error
	// first dive into inner fields
	if i < len(fieldNames)-1 {
		innerCond, targetField, innerErr = generateFilterQuery(fieldNames, i+1, reflection.ExtractRealTypeField(ft), filter)
		if innerErr != nil {
			return "", targetField, innerErr
		}
	} else {
		targetField = &innerField
	}
	table := reflection.ExtractRealTypeField(ft).Name()
	if prefix, isPoly := util.GetPolymorphic(&field); isPoly {
		polyID := prefix + "ID"
		polyType := prefix + "Type"
		outerTable := structType.Name()
		condition = append(condition, structType.Name()+".ID", "IN (", "SELECT", polyID, "FROM", table, "WHERE (")
		if len(innerCond) > 0 {
			condition = append(condition, innerCond)
		} else {
			condition = getCondition(condition, innerFieldName, filter.Value, filter.Operation)
		}
		condition = append(condition, "AND", polyID, "=", "`"+outerTable+"`.ID", "AND", polyType, "=", "'"+outerTable+"'", ")", ")")
	} else if m2mTable, isM2M := util.GetMany2Many(&field); isM2M {
		srcRef := structType.Name() + "ID"
		destRef := table + "ID"
		condition = append(condition, structType.Name()+".ID", "IN (", "SELECT", srcRef, "FROM", m2mTable, "WHERE (", destRef, "IN (", "SELECT ID FROM", table, "WHERE (")
		condition = getCondition(condition, innerFieldName, filter.Value, filter.Operation)
		condition = append(condition, ")", ")", ")", ")")
	} else {
		condition = append(condition, structType.Name()+"."+fieldName+"ID", "IN (", "SELECT ID FROM", table, "WHERE (")
		if len(innerCond) > 0 {
			condition = append(condition, innerCond)
		} else {
			condition = getCondition(condition, innerFieldName, filter.Value, filter.Operation)
		}
		condition = append(condition, ")", ")")
	}
	return strings.Join(condition, " "), targetField, nil
}

func getQapiQPrefix(f *reflect.StructField) (string, bool) {
	// get gorm tag
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

func isNull(value interface{}) bool {
	return value == "NULL" || value == "null" || value == "nil"
}
