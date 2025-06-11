package util

// SafeMySQLNaming wraps column/table names with backticks for MySQL
func SafeMySQLNaming(data string) string {
	return "`" + data + "`"
}

// Column creates a fully qualified column reference for MySQL
func Column(tableName string, columnName string) string {
	return SafeMySQLNaming(tableName) + "." + SafeMySQLNaming(columnName)
}
