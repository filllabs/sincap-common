package queryapi

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/filllabs/sincap-common/db/interfaces"
	"github.com/filllabs/sincap-common/db/types"
	"github.com/filllabs/sincap-common/db/util"
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

// QueryOptions holds additional options for query generation
type QueryOptions struct {
	JoinRegistry *JoinRegistry // Optional join registry for relationship queries
}

// GenerateDB generates SQL query and parameters from the given api Query
func GenerateDB(q *qapi.Query, entity interface{}) (*QueryResult, error) {
	return GenerateDBWithOptions(q, entity, nil)
}

// GenerateDBWithOptions generates SQL query with additional options like join registry
func GenerateDBWithOptions(q *qapi.Query, entity interface{}, options *QueryOptions) (*QueryResult, error) {
	typ, tableName := GetTableName(entity)

	var whereClauses []string
	var args []interface{}
	var orderClauses []string
	var selectFields string = "*"
	var joinClauses []string
	var joinWhereClauses []string

	// Handle field selection
	if len(q.Fields) > 0 {
		selectFields = strings.Join(q.Fields, ", ")
	}

	// Handle sorting
	if len(q.Sort) > 0 {
		for _, s := range q.Sort {
			// Check if this is already a complex sort expression (contains CASE, JSON functions, etc.)
			if strings.Contains(strings.ToUpper(s), "CASE") ||
				strings.Contains(strings.ToUpper(s), "JSON_") ||
				strings.Contains(s, "(") {
				// This is already a formatted sort expression, use it as-is
				orderClauses = append(orderClauses, s)
				continue
			}

			var direction string
			var fieldName string

			// Parse the sort clause to extract field name and direction
			parts := strings.Fields(s)
			if len(parts) >= 2 {
				// Traditional format: "fieldName direction"
				fieldName = parts[0]
				if strings.ToUpper(parts[1]) == "DESC" {
					direction = "DESC"
				} else {
					direction = "ASC"
				}
			} else if strings.HasPrefix(s, "+") {
				// Prefix format: "+fieldName"
				direction = "ASC"
				fieldName = strings.TrimPrefix(s, "+")
			} else if strings.HasPrefix(s, "-") {
				// Prefix format: "-fieldName"
				direction = "DESC"
				fieldName = strings.TrimPrefix(s, "-")
			} else {
				// Default to ASC if no direction specified
				fieldName = s
				direction = "ASC"
			}

			// Simple approach: just use field name with direction
			orderClause := fmt.Sprintf("%s %s", fieldName, direction)
			orderClauses = append(orderClauses, orderClause)
		}
	}

	// Collect relationship field paths for joins
	var relationshipPaths []string

	// Handle filters
	if len(q.Filter) > 0 {
		where, values, relPaths, err := filter2SqlWithJoins(q.Filter, typ, tableName, options)
		if err != nil {
			return nil, err
		}
		if where != "" {
			whereClauses = append(whereClauses, where)
			args = append(args, values...)
		}
		relationshipPaths = append(relationshipPaths, relPaths...)
	}

	// Handle Q search
	if len(q.Q) > 0 {
		where, values, relPaths, err := q2SqlWithJoins(q.Q, typ, tableName, options)
		if err != nil {
			return nil, err
		}
		if where != "" {
			whereClauses = append(whereClauses, fmt.Sprintf("(%s)", where))
			args = append(args, values...)
		}
		relationshipPaths = append(relationshipPaths, relPaths...)
	}

	// Generate joins if we have a join registry and relationship paths
	if options != nil && options.JoinRegistry != nil && len(relationshipPaths) > 0 {
		baseQuery := fmt.Sprintf("SELECT %s FROM %s", selectFields, util.SafeMySQLNaming(tableName))
		query, joinWheres, err := options.JoinRegistry.BuildJoinQuery(baseQuery, tableName, relationshipPaths)
		if err != nil {
			return nil, err
		}

		// Extract join clauses from the built query
		if strings.Contains(query, "JOIN") {
			parts := strings.Split(query, util.SafeMySQLNaming(tableName))
			if len(parts) > 1 {
				joinPart := strings.TrimSpace(parts[1])
				if joinPart != "" {
					joinClauses = append(joinClauses, joinPart)
				}
			}
		}

		joinWhereClauses = append(joinWhereClauses, joinWheres...)
	}

	// Build main query
	query := fmt.Sprintf("SELECT %s FROM %s", selectFields, util.SafeMySQLNaming(tableName))

	// Add joins
	if len(joinClauses) > 0 {
		query += " " + strings.Join(joinClauses, " ")
	}

	// Combine all where clauses
	allWhereClauses := append(whereClauses, joinWhereClauses...)
	if len(allWhereClauses) > 0 {
		query += " WHERE " + strings.Join(allWhereClauses, " AND ")
	}

	if len(orderClauses) > 0 {
		query += " ORDER BY " + strings.Join(orderClauses, ", ")
	}

	// Build count query for pagination
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM %s", util.SafeMySQLNaming(tableName))
	countArgs := make([]interface{}, len(args))
	copy(countArgs, args)

	// Add joins to count query too
	if len(joinClauses) > 0 {
		countQuery += " " + strings.Join(joinClauses, " ")
	}

	if len(allWhereClauses) > 0 {
		countQuery += " WHERE " + strings.Join(allWhereClauses, " AND ")
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

// GetTableName reads the table name of the given interface{} with reduced reflection
func GetTableName(e any) (reflect.Type, string) {
	// Try interface-based approach first (no reflection)
	if tableNamer, ok := e.(interfaces.TableNamer); ok {
		typ := reflection.ExtractRealTypeField(reflect.TypeOf(e))
		return typ, tableNamer.TableName()
	}

	// Fallback to reflection
	return getTableNameWithReflection(e)
}

// getTableNameWithReflection is the original reflection-based implementation
func getTableNameWithReflection(e any) (reflect.Type, string) {
	typ := reflection.ExtractRealTypeField(reflect.TypeOf(e))
	if m, hasName := typ.MethodByName("TableName"); hasName {
		res := m.Func.Call([]reflect.Value{reflect.ValueOf(e)})
		return typ, res[0].String()
	}
	return typ, typ.Name()
}

// GetTableNameOptimized gets table name without reflection when possible
func GetTableNameOptimized(e any) string {
	// Try interface-based approach first (no reflection)
	if tableNamer, ok := e.(interfaces.TableNamer); ok {
		return tableNamer.TableName()
	}

	// Fallback to reflection
	_, tableName := getTableNameWithReflection(e)
	return tableName
}

func isNull(value interface{}) bool {
	return value == "NULL" || value == "null" || value == "nil"
}
