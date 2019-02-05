// Package flags initializes all flags and parses them.
// Also a single point to control & use all flags.
package flags

import "flag"

func init() {
	Config = flag.String("config", "config.json", "Location of the config file.")
	Command = flag.String("command", "server", "Command for the executable.")
	flag.Parse()

}

// Config is the configuration path to read config.json
var Config *string

// Command is the command name to run server/init etc.
var Command *string
