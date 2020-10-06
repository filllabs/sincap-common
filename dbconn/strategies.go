package dbconn

import "github.com/jinzhu/gorm"

// AsIsNamingStrategy creates a namingstrategy which returns desired names as is without any modification
// like case or plural conversion
func AsIsNamingStrategy() *gorm.NamingStrategy {
	return &gorm.NamingStrategy{
		DB: func(name string) string {
			return name
		},
		Table: func(name string) string {
			return name
		},
		Column: func(name string) string {
			return name
		},
	}
}
