package dbutil

import (
	"bytes"
	"database/sql/driver"
	"encoding/json"
	"errors"
)

// JSON is a type for using json type at databases
type JSON []byte

// Value returns a string of json bytes
func (j JSON) Value() (driver.Value, error) {
	if j.IsNull() {
		return nil, nil
	}
	return string(j), nil
}

// Scan helps to read data and store it
func (j *JSON) Scan(value interface{}) error {
	if value == nil {
		*j = nil
		return nil
	}
	s, ok := value.([]byte)
	if !ok {
		return errors.New("Invalid Scan Source")
	}
	*j = append((*j)[0:0], s...)
	return nil
}

// MarshalJSON returns the value as []byte
func (j JSON) MarshalJSON() ([]byte, error) {
	if j == nil {
		return []byte("null"), nil
	}
	return j, nil
}

// Marshal json.Marshals the data into the given interface
func (j *JSON) Marshal(v interface{}) error {
	data, err := json.Marshal(v)
	if err != nil {
		return err
	}
	*j = append((*j)[0:0], data...)
	return nil
}

// UnmarshalJSON writes the given data in
func (j *JSON) UnmarshalJSON(data []byte) error {
	if j == nil {
		return errors.New("null point exception")
	}
	*j = append((*j)[0:0], data...)
	return nil
}

// Unmarshal json.Unmarshal the data from the given interface
func (j *JSON) Unmarshal(v interface{}) error {
	if err := json.Unmarshal([]byte(*j), v); err != nil {
		return err
	}
	return nil
}

// IsNull checks if the value is null
func (j JSON) IsNull() bool {
	return len(j) == 0 || string(j) == "null"
}

// Equals checks if the value is equal to the given parameter
func (j JSON) Equals(j1 JSON) bool {
	return bytes.Equal([]byte(j), []byte(j1))
}
