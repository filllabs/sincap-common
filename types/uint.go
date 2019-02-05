package types

import "strconv"

// ParseUint parses the given string to uint (not uint64)
func ParseUint(val string) (uint, error) {
	num, err := strconv.ParseUint(val, 10, 32)
	if err != nil {
		return 0, err
	}
	return uint(num), nil
}

// FormatUint formats the given uint to string
func FormatUint(val uint) string {
	str := strconv.Itoa(int(val))
	return str
}
