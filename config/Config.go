// Package config includes the standard functions for loading a configuration file.
// Also basic implementation of a standart web app is provides as structs
package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"

	"gitlab.com/sincap/sincap-common/db"
	"gitlab.com/sincap/sincap-common/server"
	"gitlab.com/sincap/sincap-common/server/fileserver"
	"go.uber.org/zap"
)

// Config is the configuration of the application
type Config struct {
	Server     server.Config       `json:"server"`
	FileServer []fileserver.Config `json:"fileServer"`
	DB         []db.Config         `json:"db"`
	Auth       Auth                `json:"auth"`
	Log        zap.Config          `json:"log"`
	Mail       Mail                `json:"mail"`
	Recaptcha  Recaptcha           `json:"recaptcha"`
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
	err = json.Unmarshal(data, config)
	log.Println("Config:", "Loaded ", path)
	return err
}
