package queryapi

import (
	"reflect"
	"strings"

	"gitlab.com/sincap/sincap-common/db/util"
	"gitlab.com/sincap/sincap-common/logging"
	"go.uber.org/zap"
)

func q2Sql(q string, typ reflect.Type, tableName string) (string, []interface{}, error) {

	// Convert q to  where condition with OR for all fields with tag
	where, values, err := generateQQuery(typ, tableName, q)
	if err != nil {
		logging.Logger.Warn("Can't create query from q", zap.Error(err))
	}
	return strings.Join(where, " OR "), values, nil
}

func generateQQuery(structType reflect.Type, tableName string, q string) ([]string, []interface{}, error) {
	var where []string
	var values []interface{}
	taggedFields := getQapiFields(structType)
	for _, field := range *taggedFields {
		if field.Typ.Kind() != reflect.Struct {
			where = append(where, column(tableName, field.Field.Name)+" LIKE ?")
			values = append(values, strings.Replace(field.Tag, "*", q, 1))
			continue
		}
		// if its is struct generate query recursively
		w, v, err := generateQQuery(field.Typ, field.TableName, q)
		var cond []string
		if err != nil {
			logging.Logger.Warn("Can't create query from q", zap.Error(err))
			continue
		}

		if prefix, isPoly := util.GetPolymorphic(&field.Field); isPoly {
			polyID := prefix + "ID"

			cond = append(cond, column(tableName, "ID"), "IN (", "SELECT", column(field.TableName, polyID), "FROM", safeMySQLNaming(field.TableName), "WHERE (")
			cond = append(cond, strings.Join(w, " OR "))
			cond = append(cond, ") )")
			where = append(where, strings.Join(cond, " "))
		} else if m2mTable, isM2M := util.GetMany2Many(&field.Field); isM2M {
			srcRef := column(m2mTable, tableName+"ID")
			destRef := column(m2mTable, field.TableName+"ID")
			cond = append(cond, column(tableName, "ID"), "IN (", "SELECT", srcRef, "FROM", safeMySQLNaming(m2mTable), "WHERE (", destRef, "IN (", "SELECT ", column(field.TableName, "ID"), " FROM", safeMySQLNaming(field.TableName), "WHERE (")
			cond = append(cond, strings.Join(w, " OR "))
			cond = append(cond, ")", ")", ")", ")")
			where = append(where, strings.Join(cond, " "))
		} else {
			cond = append(cond, column(tableName, field.Field.Name+"ID"), "IN (", "SELECT ", column(field.TableName, "ID"), " FROM", safeMySQLNaming(field.TableName), "WHERE (")
			cond = append(cond, strings.Join(w, " OR "))
			cond = append(cond, ")", ")")
			where = append(where, strings.Join(cond, " "))
		}
		values = append(values, v...)

	}
	return where, values, nil
}
