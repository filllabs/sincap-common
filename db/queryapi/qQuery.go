package queryapi

import (
	"reflect"
	"strings"

	"github.com/filllabs/sincap-common/db/util"
	"github.com/filllabs/sincap-common/logging"
	"go.uber.org/zap"
)

// q2SqlWithJoins generates SQL with support for joins
func q2SqlWithJoins(q string, typ reflect.Type, tableName string, options *QueryOptions) (string, []interface{}, []string, error) {
	// Convert q to  where condition with OR for all fields with tag
	where, values, relationshipPaths, err := generateQQueryWithJoins(typ, tableName, q, options)
	if err != nil {
		logging.Logger.Warn("Can't create query from q", zap.Error(err))
	}
	return strings.Join(where, " OR "), values, relationshipPaths, nil
}

// q2Sql generates SQL without join support (backward compatibility)
func q2Sql(q string, typ reflect.Type, tableName string) (string, []interface{}, error) {
	where, values, _, err := q2SqlWithJoins(q, typ, tableName, nil)
	return where, values, err
}

func generateQQueryWithJoins(structType reflect.Type, tableName string, q string, options *QueryOptions) ([]string, []interface{}, []string, error) {
	var where []string
	var values []interface{}
	var relationshipPaths []string

	taggedFields := getQapiFields(structType)
	for _, field := range *taggedFields {
		if field.Typ.Kind() != reflect.Struct {
			where = append(where, util.Column(tableName, field.Field.Name)+" LIKE ?")
			values = append(values, strings.Replace(field.Tag, "*", q, 1))
			continue
		}

		// if its is struct generate query recursively
		w, v, relPaths, err := generateQQueryWithJoins(field.Typ, field.TableName, q, options)
		var cond []string
		if err != nil {
			logging.Logger.Warn("Can't create query from q", zap.Error(err))
			continue
		}

		// Try to use join registry if available
		if options != nil && options.JoinRegistry != nil {
			fieldPath := field.Field.Name
			if _, exists := options.JoinRegistry.Get(fieldPath); exists {
				// Use join-based approach
				if len(w) > 0 {
					cond = append(cond, strings.Join(w, " OR "))
					relationshipPaths = append(relationshipPaths, fieldPath)
				}
				where = append(where, strings.Join(cond, " "))
				values = append(values, v...)
				relationshipPaths = append(relationshipPaths, relPaths...)
				continue
			}
		}

		// Fallback to subquery approach (no joins)
		cond = append(cond, util.Column(tableName, field.Field.Name+"ID"), "IN (", "SELECT ", util.Column(field.TableName, "ID"), " FROM", util.SafeMySQLNaming(field.TableName), "WHERE (")
		cond = append(cond, strings.Join(w, " OR "))
		cond = append(cond, ")", ")")
		where = append(where, strings.Join(cond, " "))
		values = append(values, v...)
		relationshipPaths = append(relationshipPaths, relPaths...)
	}
	return where, values, relationshipPaths, nil
}
