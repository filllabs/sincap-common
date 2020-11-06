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

// Marshal json.Marshals the data from the given interface
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

// UnmarshalArray writes the data to a array
func (j *JSON) UnmarshalArray() ([]interface{}, error) {
	if j == nil {
		return nil, errors.New("null point exception")
	}
	var arr []interface{}
	if err := json.Unmarshal([]byte(*j), &arr); err != nil {
		return nil, err
	}
	return arr, nil
}

// UnmarshalMap writes the data to a map
func (j *JSON) UnmarshalMap() (map[string]interface{}, error) {
	if j == nil {
		return nil, errors.New("null point exception")
	}
	var m map[string]interface{}
	if err := json.Unmarshal([]byte(*j), &m); err != nil {
		return nil, err
	}
	return m, nil
}

// Unmarshal json.Unmarshal the data from the given interface
func (j *JSON) Unmarshal(v interface{}) error {
	if j == nil {
		return errors.New("null point exception")
	}
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
