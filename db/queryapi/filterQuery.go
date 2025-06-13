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
			} else if cond, f, relPath, err := generateFilterQuery(fieldNames, 1, typ, tableName, filter, options); err == nil {
				condition = append(condition, cond)
				targetField = f
				if relPath != "" {
					relationshipPaths = append(relationshipPaths, relPath)
				}
			} else {
				return "", values, relationshipPaths, err
			}

		} else {
			condition = getCondition(condition, util.SafeMySQLNaming(tableName)+"."+util.SafeMySQLNaming(filter.Name), filter.Value, filter.Operation)
			field, isFieldFound := typ.FieldByName(filter.Name)
			if !isFieldFound {
				return "", values, relationshipPaths, fmt.Errorf("Can't find field for %s", filter.Name)
			}
			targetField = &field

		}
		where = append(where, strings.Join(condition, " "))
		fieldType := reflection.ExtractRealTypeField(targetField.Type)
		kind := fieldType.Kind()

		// Use proper value conversion based on field type
		switch filter.Operation {
		case qapi.IN:
			inVals := strings.Split(filter.Value, "|")
			for i := 0; i < len(inVals); i++ {
				var err error
				values, err = util.ConvertValue(filter, fieldType, kind, values, inVals[i])
				if err != nil {
					return "", values, relationshipPaths, err
				}
			}
		case qapi.IN_ALT:
			inVals := strings.Split(filter.Value, "*")
			for i := 0; i < len(inVals); i++ {
				var err error
				values, err = util.ConvertValue(filter, fieldType, kind, values, inVals[i])
				if err != nil {
					return "", values, relationshipPaths, err
				}
			}
		default:
			var err error
			values, err = util.ConvertValue(filter, fieldType, kind, values, filter.Value)
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

func generateFilterQuery(fieldNames []string, i int, typ reflect.Type, tableName string, filter qapi.Filter, options *QueryOptions) (string, *reflect.StructField, string, error) {
	var condition []string
	var targetField *reflect.StructField

	fieldName := fieldNames[i-1]
	innerFieldName := fieldNames[i]

	field, isFieldFound := typ.FieldByName(fieldName)
	if !isFieldFound {
		return "", nil, "", fmt.Errorf("Can't find field for %s", fieldName)
	}

	// Get the type of the field
	fieldType := reflection.DepointerField(field.Type)
	if fieldType.Kind() == reflect.Slice {
		fieldType = reflection.DepointerField(fieldType.Elem())
	}

	// Get table name for the related model
	val := reflect.New(fieldType)
	_, table := GetTableName(val.Interface())

	var innerCond string
	if i+1 < len(fieldNames) {
		// Recursive call for nested relationships
		if cond, f, _, err := generateFilterQuery(fieldNames, i+1, fieldType, table, filter, options); err == nil {
			innerCond = cond
			targetField = f
		} else {
			return "", targetField, "", err
		}
	} else {
		// This is the final field, get its type from the related model
		if innerField, found := fieldType.FieldByName(innerFieldName); found {
			targetField = &innerField
		} else {
			return "", nil, "", fmt.Errorf("Can't find field %s in %s", innerFieldName, fieldType.Name())
		}
	}

	// Check if we have join configuration for this relationship
	if options != nil && options.JoinRegistry != nil {
		fieldPath := strings.Join(fieldNames[:i], ".")
		if config, exists := options.JoinRegistry.Get(fieldPath); exists {
			// Use the configured table name and generate appropriate condition
			if len(innerCond) > 0 {
				condition = append(condition, innerCond)
			} else {
				condition = getCondition(condition, util.Column(config.Table, innerFieldName), filter.Value, filter.Operation)
			}
			return strings.Join(condition, " "), targetField, fieldPath, nil
		}
	}

	// Fallback to subquery approach (no joins)
	condition = append(condition, util.Column(tableName, fieldName+"ID"), "IN (", "SELECT "+util.Column(table, "ID"), " FROM", util.SafeMySQLNaming(table), "WHERE (")
	if len(innerCond) > 0 {
		condition = append(condition, innerCond)
	} else {
		condition = getCondition(condition, util.Column(table, innerFieldName), filter.Value, filter.Operation)
	}
	condition = append(condition, ")", ")")

	return strings.Join(condition, " "), targetField, "", nil
}

func getCondition(condition []string, field string, value interface{}, operation qapi.Operation) []string {
	switch operation {
	case qapi.EQ:
		if isNull(value) {
			condition = append(condition, field, "IS NULL")
		} else {
			condition = append(condition, field, "=", "?")
		}
	case qapi.NEQ:
		if isNull(value) {
			condition = append(condition, field, "IS NOT NULL")
		} else {
			condition = append(condition, field, "!=", "?")
		}
	case qapi.GT:
		condition = append(condition, field, ">", "?")
	case qapi.GTE:
		condition = append(condition, field, ">=", "?")
	case qapi.LT:
		condition = append(condition, field, "<", "?")
	case qapi.LTE:
		condition = append(condition, field, "<=", "?")
	case qapi.LK:
		condition = append(condition, field, "LIKE", "?")
	case qapi.IN:
		params := strings.Split(value.(string), "|")
		paramSection := strings.Repeat("?,", len(params))
		condition = append(condition, field, "IN", "("+paramSection[0:len(paramSection)-1]+")")
	case qapi.IN_ALT:
		params := strings.Split(value.(string), "*")
		paramSection := strings.Repeat("?,", len(params))
		condition = append(condition, field, "IN", "("+paramSection[0:len(paramSection)-1]+")")
	default:
		condition = append(condition, field, "=", "?")
	}
	return condition
}
