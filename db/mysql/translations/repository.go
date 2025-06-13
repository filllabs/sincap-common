package translations

import (
	"time"

	"github.com/jmoiron/sqlx"
)

// Simple in-memory cache for language codes
var langCodesCache []string
var langCodesCacheTime time.Time
var cacheDuration = 5 * time.Minute

// ListCodes retrieves all language codes from the database with simple caching
func ListCodes(db *sqlx.DB) ([]string, error) {
	// Check if cache is still valid
	if time.Since(langCodesCacheTime) < cacheDuration && len(langCodesCache) > 0 {
		return langCodesCache, nil
	}

	// Query the database for language codes
	var codes []string
	err := db.Select(&codes, "SELECT code FROM Language")
	if err != nil {
		return nil, err
	}

	// Update cache
	langCodesCache = codes
	langCodesCacheTime = time.Now()

	return codes, nil
}
