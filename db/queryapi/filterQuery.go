package queryapi

import (
	"fmt"
	"reflect"
	"strings"

	"gitlab.com/sincap/sincap-common/db/util"
	"gitlab.com/sincap/sincap-common/middlewares/qapi"
	"gitlab.com/sincap/sincap-common/reflection"
)

func filter2Sql(filters []qapi.Filter, typ reflect.Type, tableName string) (string, []interface{}, error) {
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
				return "", nil, fmt.Errorf("Can't find struct: %s field: %s", tableName, filter.Name)
			}

			dp := reflection.DepointerField(field.Type)
			if dp == jsonType {
				// concat new value
				condition = getCondition(condition, tableName+"."+fieldNames[0]+"->"+"'$."+fieldNames[1]+"'", filter.Value, qapi.LK)
				values = append(values, filter.Value)
				targetField = &field
			} else if cond, f, err := generateFilterQuery(fieldNames, 1, typ, tableName, filter); err == nil {
				condition = append(condition, cond)
				targetField = f
			} else {
				return "", values, err
			}

		} else {
			condition = getCondition(condition, safeMySQLNaming(tableName)+"."+safeMySQLNaming(filter.Name), filter.Value, filter.Operation)
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

func generateFilterQuery(fieldNames []string, i int, structType reflect.Type, tableName string, filter qapi.Filter) (string, *reflect.StructField, error) {

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
		// ftType, ftTableName := getTableName(entity)
		innerCond, targetField, innerErr = generateFilterQuery(fieldNames, i+1, ft, reflection.ExtractRealTypeField(field.Type).Name(), filter)
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
		outerTable := tableName
		condition = append(condition, column(outerTable, "ID"), "IN (", "SELECT", column(table, polyID), "FROM", safeMySQLNaming(table), "WHERE (")
		if len(innerCond) > 0 {
			condition = append(condition, innerCond)
		} else {
			condition = getCondition(condition, column(table, innerFieldName), filter.Value, filter.Operation)
		}
		condition = append(condition, "AND", column(table, polyID), "=", column(outerTable, "ID"), "AND", column(table, polyType), "=", "'"+outerTable+"'", ")", ")")
	} else if m2mTable, isM2M := util.GetMany2Many(&field); isM2M {
		srcRef := tableName + "ID"
		destRef := table + "ID"
		condition = append(condition, safeMySQLNaming(tableName)+".ID", "IN (", "SELECT", safeMySQLNaming(srcRef), "FROM", safeMySQLNaming(m2mTable), "WHERE (", safeMySQLNaming(destRef), "IN (", "SELECT ID FROM", safeMySQLNaming(table), "WHERE (")
		condition = getCondition(condition, safeMySQLNaming(innerFieldName), filter.Value, filter.Operation)
		condition = append(condition, ")", ")", ")", ")")
	} else {
		condition = append(condition, column(tableName, fieldName+"ID"), "IN (", "SELECT "+column(table, "ID"), " FROM", safeMySQLNaming(table), "WHERE (")
		if len(innerCond) > 0 {
			condition = append(condition, innerCond)
		} else {
			condition = getCondition(condition, column(table, innerFieldName), filter.Value, filter.Operation)
		}
		condition = append(condition, ")", ")")
	}
	return strings.Join(condition, " "), targetField, nil
}
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
