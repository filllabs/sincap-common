package translations

import (
	"time"

	"github.com/patrickmn/go-cache"
	"gorm.io/gorm"
)

var langCodesCACHE = cache.New(5*time.Minute, 10*time.Minute)

func ListCodes(db *gorm.DB) ([]string, error) {
	value, found := langCodesCACHE.Get("langCodes")
	if found {
		return value.([]string), nil
	}
	var codes []string
	err := db.Table("Language").Pluck("code", &codes).Error
	if err != nil {
		// Cache empty result on error (e.g., table doesn't exist)
		// to avoid repeated database queries on every request
		langCodesCACHE.Set("langCodes", codes, 5*time.Minute)
		return nil, err
	}
	langCodesCACHE.Set("langCodes", codes, 5*time.Minute)
	return codes, nil
}
