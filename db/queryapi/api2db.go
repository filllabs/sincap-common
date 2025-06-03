package queryapi

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/filllabs/sincap-common/db/types"
	"github.com/jmoiron/sqlx"

	"github.com/filllabs/sincap-common/middlewares/qapi"
	"github.com/filllabs/sincap-common/reflection"
)

var jsonType = reflect.TypeOf(types.JSON{})

// QueryResult holds the generated SQL query and parameters
type QueryResult struct {
	Query      string
	Args       []interface{}
	CountQuery string
	CountArgs  []interface{}
}

// GenerateSQL generates SQL query and parameters from the given api Query
func GenerateSQL(q *qapi.Query, entity interface{}) (*QueryResult, error) {
	typ, tableName := GetTableName(entity)

	var whereClauses []string
	var args []interface{}
	var orderClauses []string
	var selectFields string = "*"

	// Handle field selection
	if len(q.Fields) > 0 {
		selectFields = strings.Join(q.Fields, ", ")
	}

	// Handle sorting
	if len(q.Sort) > 0 {
		for _, s := range q.Sort {
			values := strings.Split(s, " ")
			fieldNames := strings.Split(values[0], ".")
			field, isFieldFound := typ.FieldByName(fieldNames[0])
			if isFieldFound {
				dp := reflection.DepointerField(field.Type)
				if dp == jsonType {
					orderClause := fmt.Sprintf("CAST(%s.%s->'$.%s' AS CHAR) %s",
						safeMySQLNaming(tableName),
						safeMySQLNaming(fieldNames[0]),
						fieldNames[1],
						values[1])
					orderClauses = append(orderClauses, orderClause)
				} else {
					orderClause := fmt.Sprintf("%s %s",
						safeMySQLNaming(strings.Join(fieldNames, "__")),
						values[1])
					orderClauses = append(orderClauses, orderClause)
				}
			}
		}
	}

	// Handle filters
	if len(q.Filter) > 0 {
		where, values, err := filter2Sql(q.Filter, typ, tableName)
		if err != nil {
			return nil, err
		}
		if where != "" {
			whereClauses = append(whereClauses, where)
			args = append(args, values...)
		}
	}

	// Handle Q search
	if len(q.Q) > 0 {
		where, values, err := q2Sql(q.Q, typ, tableName)
		if err != nil {
			return nil, err
		}
		if where != "" {
			whereClauses = append(whereClauses, fmt.Sprintf("(%s)", where))
			args = append(args, values...)
		}
	}

	// Build main query
	query := fmt.Sprintf("SELECT %s FROM %s", selectFields, safeMySQLNaming(tableName))

	if len(whereClauses) > 0 {
		query += " WHERE " + strings.Join(whereClauses, " AND ")
	}

	if len(orderClauses) > 0 {
		query += " ORDER BY " + strings.Join(orderClauses, ", ")
	}

	// Build count query for pagination
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM %s", safeMySQLNaming(tableName))
	countArgs := make([]interface{}, len(args))
	copy(countArgs, args)

	if len(whereClauses) > 0 {
		countQuery += " WHERE " + strings.Join(whereClauses, " AND ")
	}

	// Add pagination to main query
	if q.Limit > 0 {
		query += " LIMIT ?"
		args = append(args, q.Limit)
	}

	if q.Offset > 0 {
		query += " OFFSET ?"
		args = append(args, q.Offset)
	}

	return &QueryResult{
		Query:      query,
		Args:       args,
		CountQuery: countQuery,
		CountArgs:  countArgs,
	}, nil
}

// GenerateDB generates a valid db query from the given api Query (deprecated - use GenerateSQL)
func GenerateDB(q *qapi.Query, db *sqlx.DB, entity interface{}) (*sqlx.DB, error) {
	// This function is kept for backward compatibility but should be replaced with GenerateSQL
	return db, fmt.Errorf("GenerateDB is deprecated, use GenerateSQL instead")
}

/*
GetTableName reads the table name of the given interface{}
*/
func GetTableName(e any) (reflect.Type, string) {
	typ := reflection.ExtractRealTypeField(reflect.TypeOf(e))
	if m, hasName := typ.MethodByName("TableName"); hasName {
		res := m.Func.Call([]reflect.Value{reflect.ValueOf(e)})
		return typ, res[0].String()
	}
	return typ, typ.Name()
}

func isNull(value interface{}) bool {
	return value == "NULL" || value == "null" || value == "nil"
}

func safeMySQLNaming(data string) string {
	return "`" + data + "`"
}

func column(tableName string, columnName string) string {
	return safeMySQLNaming(tableName) + "." + safeMySQLNaming(columnName)
}
