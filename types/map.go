package types

// MapValuesAsStrings converts all values to string and returns as a slice
func MapValuesAsStrings(data *map[string]interface{}) []string {
	var sArr []string
	for _, val := range *data {
		if s, err := ToString(val); err == nil {
			sArr = append(sArr, s)
		}
	}
	return sArr
}

// MapKeysAsStrings collects all keys and returns as a slice
func MapKeysAsStrings(data *map[string]interface{}) []string {
	keys := make([]string, 0, len(*data))
	for k := range *data {
		keys = append(keys, k)
	}
	return keys
}
