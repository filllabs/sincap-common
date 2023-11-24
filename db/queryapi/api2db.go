package queryapi

import (
	"reflect"
	"strings"
	"time"

	"github.com/filllabs/sincap-common/db/types"
	"gorm.io/gorm"

	"github.com/filllabs/sincap-common/middlewares/qapi"
	"github.com/filllabs/sincap-common/reflection"
)

var timeKind = reflect.TypeOf(time.Time{}).Kind()
var jsonType = reflect.TypeOf(types.JSON{})

// GenerateDB generates a valid db query from the given api Query
func GenerateDB(q *qapi.Query, db *gorm.DB, entity interface{}) (*gorm.DB, error) {
	typ, tableName := GetTableName(entity)

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
					c := "CAST(" + tableName + "." + fieldNames[0] + "->" + "'$." + fieldNames[1] + "'" + "AS CHAR) " + values[1]
					sortFields = append(sortFields, c)
				} else {
					sortFields = append(sortFields, safeMySQLNaming(strings.Join(fieldNames, "__"))+" "+values[1])
				}
			}

		}
		db = db.Order(strings.Join(sortFields, ", "))
	}
	if len(q.Filter) > 0 {
		where, values, err := filter2Sql(q.Filter, typ, tableName)
		if err != nil {
			return db, err
		}
		db = db.Where(where, values...)
	}

	if len(q.Fields) > 0 {
		db = db.Select(q.Fields)
	}

	if len(q.Q) > 0 {
		where, values, err := q2Sql(q.Q, typ, tableName)
		if err != nil {
			return db, err
		}
		db = db.Where(where, values...)
	}
	return db, nil
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
