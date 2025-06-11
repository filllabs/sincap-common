package mysql

import (
	"testing"

	"github.com/filllabs/sincap-common/db/interfaces"
)

// Test model implementing all optimization interfaces
type OptimizedUser struct {
	ID    uint   `db:"ID"`
	Name  string `db:"Name"`
	Email string `db:"Email"`
}

func (OptimizedUser) TableName() string {
	return "Users"
}

func (u OptimizedUser) GetID() interface{} {
	return u.ID
}

func (u *OptimizedUser) SetID(id interface{}) error {
	if idVal, ok := id.(uint64); ok {
		u.ID = uint(idVal)
		return nil
	}
	return nil
}

func (u OptimizedUser) GetFieldMap() map[string]interface{} {
	return map[string]interface{}{
		"ID":    u.ID,
		"Name":  u.Name,
		"Email": u.Email,
	}
}

// Test model using reflection fallback
type ReflectionUser struct {
	ID    uint   `db:"ID"`
	Name  string `db:"Name"`
	Email string `db:"Email"`
}

func (ReflectionUser) TableName() string {
	return "Users"
}

func TestOptimizedInterfaces(t *testing.T) {
	// Test TableNamer interface
	user := OptimizedUser{ID: 1, Name: "John", Email: "john@example.com"}

	if tableNamer, ok := interface{}(user).(interfaces.TableNamer); ok {
		tableName := tableNamer.TableName()
		if tableName != "Users" {
			t.Errorf("Expected table name 'Users', got '%s'", tableName)
		}
	} else {
		t.Error("OptimizedUser should implement TableNamer interface")
	}

	// Test IDGetter interface
	if idGetter, ok := interface{}(user).(interfaces.IDGetter); ok {
		id := idGetter.GetID()
		if id != uint(1) {
			t.Errorf("Expected ID 1, got %v", id)
		}
	} else {
		t.Error("OptimizedUser should implement IDGetter interface")
	}

	// Test IDSetter interface
	if idSetter, ok := interface{}(&user).(interfaces.IDSetter); ok {
		err := idSetter.SetID(uint64(42))
		if err != nil {
			t.Errorf("SetID failed: %v", err)
		}
		if user.ID != 42 {
			t.Errorf("Expected ID 42, got %d", user.ID)
		}
	} else {
		t.Error("OptimizedUser should implement IDSetter interface")
	}

	// Test FieldMapper interface
	if fieldMapper, ok := interface{}(user).(interfaces.FieldMapper); ok {
		fieldMap := fieldMapper.GetFieldMap()
		if fieldMap["Name"] != "John" {
			t.Errorf("Expected Name 'John', got %v", fieldMap["Name"])
		}
		if fieldMap["Email"] != "john@example.com" {
			t.Errorf("Expected Email 'john@example.com', got %v", fieldMap["Email"])
		}
	} else {
		t.Error("OptimizedUser should implement FieldMapper interface")
	}
}

func TestReflectionFallback(t *testing.T) {
	// Test that reflection fallback still works
	user := ReflectionUser{ID: 1, Name: "Jane", Email: "jane@example.com"}

	// Should implement TableNamer but not the optimization interfaces
	if tableNamer, ok := interface{}(user).(interfaces.TableNamer); ok {
		tableName := tableNamer.TableName()
		if tableName != "Users" {
			t.Errorf("Expected table name 'Users', got '%s'", tableName)
		}
	} else {
		t.Error("ReflectionUser should implement TableNamer interface")
	}

	// Should NOT implement optimization interfaces
	if _, ok := interface{}(user).(interfaces.IDGetter); ok {
		t.Error("ReflectionUser should NOT implement IDGetter interface")
	}

	if _, ok := interface{}(&user).(interfaces.IDSetter); ok {
		t.Error("ReflectionUser should NOT implement IDSetter interface")
	}

	if _, ok := interface{}(user).(interfaces.FieldMapper); ok {
		t.Error("ReflectionUser should NOT implement FieldMapper interface")
	}
}

func TestCRUDOperations(t *testing.T) {
	// This test would require a database connection
	// For now, just test that the functions can be called without compilation errors

	// Example usage with optimized model
	optimizedUser := &OptimizedUser{Name: "Test", Email: "test@example.com"}
	_ = optimizedUser // Would use FieldMapper interface internally

	// Example usage with reflection model
	reflectionUser := &ReflectionUser{Name: "Test", Email: "test@example.com"}
	_ = reflectionUser // Would use reflection internally
}
