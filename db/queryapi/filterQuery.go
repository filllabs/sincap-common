package queryapi

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/filllabs/sincap-common/db/util"
	"github.com/filllabs/sincap-common/middlewares/qapi"
	"github.com/filllabs/sincap-common/reflection"
)

// filter2SqlWithJoins generates SQL with support for joins
func filter2SqlWithJoins(filters []qapi.Filter, typ reflect.Type, tableName string, options *QueryOptions) (string, []interface{}, []string, error) {
	var where []string
	var values []interface{}
	var relationshipPaths []string
	var targetField *reflect.StructField

	// Convert all filters to a where condition with AND
	for _, filter := range filters {
		var condition []string

		// Get field name with split (it will make inner field queries possible)
		fieldNames := strings.Split(filter.Name, ".")

		// If it has more than 1 field name it has inner fields (another table)
		if len(fieldNames) > 1 {
			// If it is json than handle different
			field, isFieldFound := typ.FieldByName(fieldNames[0])
			if !isFieldFound {
				return "", nil, nil, fmt.Errorf("Can't find struct: %s field: %s", tableName, filter.Name)
			}

			dp := reflection.DepointerField(field.Type)
			if dp == jsonType {
				// concat new value
				condition = getCondition(condition, tableName+"."+fieldNames[0]+"->"+"'$."+fieldNames[1]+"'", filter.Value, qapi.LK)
				values = append(values, filter.Value)
				targetField = &field
			} else if cond, f, relPath, err := generateFilterQueryWithJoins(fieldNames, 1, typ, tableName, filter, options); err == nil {
				condition = append(condition, cond)
				targetField = f
				if relPath != "" {
					relationshipPaths = append(relationshipPaths, relPath)
				}
			} else {
				return "", values, relationshipPaths, err
			}

		} else {
			condition = getCondition(condition, safeMySQLNaming(tableName)+"."+safeMySQLNaming(filter.Name), filter.Value, filter.Operation)
			field, isFieldFound := typ.FieldByName(filter.Name)
			if !isFieldFound {
				return "", values, relationshipPaths, fmt.Errorf("Can't find field for %s", filter.Name)
			}
			targetField = &field

		}
		where = append(where, strings.Join(condition, " "))
		kind := reflection.ExtractRealTypeField(targetField.Type).Kind()
		switch filter.Operation {
		case qapi.IN:
			inVals := strings.Split(filter.Value, "|")
			for i := 0; i < len(inVals); i++ {
				var err error
				values, err = util.ConvertValue(filter, typ, kind, values, inVals[i])
				if err != nil {
					return "", values, relationshipPaths, err
				}
			}
		case qapi.IN_ALT:
			inVals := strings.Split(filter.Value, "*")
			for i := 0; i < len(inVals); i++ {
				var err error
				values, err = util.ConvertValue(filter, typ, kind, values, inVals[i])
				if err != nil {
					return "", values, relationshipPaths, err
				}
			}
		default:
			var err error
			values, err = util.ConvertValue(filter, typ, kind, values, filter.Value)
			if err != nil {
				return "", values, relationshipPaths, err
			}
		}

	}
	return strings.Join(where, " AND "), values, relationshipPaths, nil
}

// filter2Sql generates SQL without join support (backward compatibility)
func filter2Sql(filters []qapi.Filter, typ reflect.Type, tableName string) (string, []interface{}, error) {
	where, values, _, err := filter2SqlWithJoins(filters, typ, tableName, nil)
	return where, values, err
}

func generateFilterQueryWithJoins(fieldNames []string, i int, structType reflect.Type, tableName string, filter qapi.Filter, options *QueryOptions) (string, *reflect.StructField, string, error) {
	var condition []string

	fieldName := fieldNames[i-1]
	innerFieldName := fieldNames[i]

	if structType.Kind() != reflect.Struct {
		return "", nil, "", fmt.Errorf("%s is not struct", structType.Name())
	}
	field, isFieldFound := structType.FieldByName(fieldName)
	if !isFieldFound {
		return "", nil, "", fmt.Errorf("Can't find struct: %s field: %s", structType.Name(), filter.Name)
	}
	ft := reflection.ExtractRealTypeField(field.Type)

	if ft.Kind() != reflect.Struct && ft.Kind() != reflect.Slice {
		return "", nil, "", fmt.Errorf("%s is not struct field in %s", filter.Name, structType.Name())
	}
	//TODO:check  maybe noo need previous extect handles it
	if ft.Kind() == reflect.Slice {
		ft = reflection.ExtractRealTypeField(ft.Elem())
	}
	innerField, isInnerFieldFound := reflection.ExtractRealTypeField(ft).FieldByName(innerFieldName)
	if !isInnerFieldFound {
		return "", nil, "", fmt.Errorf("Can't find struct: %s inner field: %s", reflection.ExtractRealTypeField(ft).Name(), filter.Name)
	}
	innerCond := ""
	var targetField *reflect.StructField
	var innerErr error
	var relationshipPath string

	// first dive into inner fields
	if i < len(fieldNames)-1 {
		innerCond, targetField, relationshipPath, innerErr = generateFilterQueryWithJoins(fieldNames, i+1, ft, reflection.ExtractRealTypeField(field.Type).Name(), filter, options)
		if innerErr != nil {
			return "", targetField, relationshipPath, innerErr
		}
	} else {
		targetField = &innerField
	}

	table := reflection.ExtractRealTypeField(ft).Name()

	// Try to use join registry if available
	if options != nil && options.JoinRegistry != nil {
		fieldPath := strings.Join(fieldNames[:i], ".")
		if config, exists := options.JoinRegistry.Get(fieldPath); exists {
			// Use the configured table name and generate appropriate condition
			if len(innerCond) > 0 {
				condition = append(condition, innerCond)
			} else {
				condition = getCondition(condition, column(config.Table, innerFieldName), filter.Value, filter.Operation)
			}
			return strings.Join(condition, " "), targetField, fieldPath, nil
		}
	}

	// Fallback to subquery approach (no joins)
	condition = append(condition, column(tableName, fieldName+"ID"), "IN (", "SELECT "+column(table, "ID"), " FROM", safeMySQLNaming(table), "WHERE (")
	if len(innerCond) > 0 {
		condition = append(condition, innerCond)
	} else {
		condition = getCondition(condition, column(table, innerFieldName), filter.Value, filter.Operation)
	}
	condition = append(condition, ")", ")")

	return strings.Join(condition, " "), targetField, "", nil
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
