package crud

import (
	"reflect"

	"gorm.io/gorm"
)

// MultiCreate  creates multiple records consecutively. Stops on any error, returns the error.
// Don't forget to start TX and rollback on any error if you need transactions support.
func MultiCreate(DB *gorm.DB, records ...interface{}) error {
	for _, record := range records {
		if record == nil {
			continue
		}
		err := Create(DB, record)
		if err != nil {
			return err
		}
	}
	return nil
}

// MultiUpdate  updates multiple records consecutively. Stops on any error, returns the error.
// Don't forget to start TX and rollback on any error if you need transactions support.
func MultiUpdate(DB *gorm.DB, records ...interface{}) error {
	for _, record := range records {
		if record == nil {
			continue
		}
		err := Update(DB, record)
		if err != nil {
			return err
		}
	}
	return nil
}

// MultiDelete  deletes multiple records consecutively. Stops on any error, returns the error.
// Don't forget to start TX and rollback on any error if you need transactions support.
func MultiDelete(DB *gorm.DB, records ...interface{}) error {
	for _, record := range records {
		if isNotValid(record) {
			continue
		}
		err := Delete(DB, record)
		if err != nil {
			return err
		}
	}
	return nil
}

func isNotValid(record interface{}) bool {
	if record == nil {
		return true
	}

	var val = reflect.ValueOf(record)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
		if !val.IsValid() {
			return true
		}
	}
	if val.FieldByName("ID").Uint() == 0 {
		return true
	}
	return false
}
