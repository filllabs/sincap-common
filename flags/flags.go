// Package flags  helps to parse flags.
package flags

import (
	"errors"
	"flag"
)

// Parse takes a slice of strings () and parses command line flags
func Parse(fArr ...string) (*flag.FlagSet, error) {
	fs := flag.NewFlagSet("", flag.PanicOnError)
	length := len(fArr)
	if length%3 != 0 {
		return nil, errors.New("input array must be dividable by 3 (name,value,usage)")
	}

	for i := 0; i < len(fArr); i += 3 {
		fs.String(fArr[i], fArr[i+1], fArr[i+2])
	}
	return fs, nil
}

// Defaults are the information needed for config and commanf flags.
var Defaults = []string{"config", "config.json", "Location of the config file.", "command", "server", "Command for the executable."}
