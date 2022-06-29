// Package config includes the standard functions for loading a configuration file.
// Also basic implementation of a standart web app is provides as structs
package config

import (
	"fmt"
	"io/ioutil"
	"log"
	"strings"

	"github.com/yosuke-furukawa/json5/encoding/json5"
	"gitlab.com/sincap/sincap-common/db"
	"gitlab.com/sincap/sincap-common/server"
	"gitlab.com/sincap/sincap-common/server/fileserver"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
)

type Config struct {
	Server server.Config `json:"server,omitempty" yaml:"server,omitempty" `
	Auth   Auth          `json:"auth" yaml:"auth"`

	FileServer []fileserver.Config `json:"fileServer,omitempty" yaml:"fileServer,omitempty"`
	DB         []db.Config         `json:"db,omitempty" 	yaml:"db,omitempty"`
	Log        zap.Config          `json:"log,omitempty" yaml:"log,omitempty"`
	Mail       Mail                `json:"mail,omitempty" yaml:"mail,omitempty"`
}

// Load loads the configuration file from the given path and fills the given config pointer
func Load(path string, config interface{}) error {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		log.Printf("Config: Can't read configuration file from %s", path)
		// USED AT TESTS: Check for relative path
		data, err = ioutil.ReadFile("../../" + path)
		if err != nil {
			return fmt.Errorf("Config: Can't read configuration file from %v", err)
		}
	}
	if strings.HasSuffix(path, ".json") || strings.HasSuffix(path, ".json5") {
		err = json5.Unmarshal(data, config)
	} else if strings.HasSuffix(path, ".yaml") || strings.HasSuffix(path, ".yml") {
		err = yaml.Unmarshal(data, config)
	} else {
		return fmt.Errorf("Config: Unknown file extension for %s", path)
	}
	log.Println("Config:", "Loaded ", path)
	return err
}
