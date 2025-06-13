package translations

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
)

const DEFAULT_LANG_CODE = "en-US"

// Translations is a type alias for a map of string to string. It is used to store translations for specific fields in all languages.
type Translations struct {
	data map[string]string
}

// Get returns the translation for the given key.
func (t *Translations) Get(key string) string {
	return t.data[key]
}

// Set sets the translation for the given key.
func (t *Translations) Set(key, value string) {
	if t.data == nil {
		t.data = make(map[string]string)
	}
	t.data[key] = value
}

// MarshalJSON is a custom JSON marshaller for the Translations type.
func (t *Translations) MarshalJSON() ([]byte, error) {
	return json.Marshal(t.data)
}

// UnmarshalJSON is a custom JSON unmarshaller for the Translations type.
func (t *Translations) UnmarshalJSON(data []byte) error {
	if t.data == nil {
		t.data = make(map[string]string)
	}
	return json.Unmarshal(data, &t.data)
}

// Marshal is used for saving to the database.
func (t *Translations) Marshal() ([]byte, error) {
	return json.Marshal(t.data)
}

// Unmarshal is used for reading from the database.
func (t *Translations) Unmarshal(data []byte) error {
	if t.data == nil {
		t.data = make(map[string]string)
	}
	return json.Unmarshal(data, &t.data)
}

// Scan implements the sql.Scanner interface for reading from the database.
func (t *Translations) Scan(value any) error {
	if value == nil {
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("unsupported type for Translations")
	}

	// Handle empty values
	if len(bytes) == 0 {
		return nil
	}

	// Try to unmarshal as JSON first
	if t.data == nil {
		t.data = make(map[string]string)
	}

	// Check if it's valid JSON by trying to unmarshal
	err := json.Unmarshal(bytes, &t.data)
	if err != nil {
		// If JSON unmarshaling fails, treat it as a plain string
		// Store it with the DEFAULT_LANG_CODE as the key
		t.data = make(map[string]string)
		t.data[DEFAULT_LANG_CODE] = string(bytes)
		return nil
	}

	return nil
}

// Value implements the driver.Valuer interface for writing to the database.
func (t *Translations) Value() (driver.Value, error) {
	return t.Marshal()
}
