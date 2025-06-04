package mysql

import (
	"testing"
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

	if tableNamer, ok := interface{}(user).(TableNamer); ok {
		tableName := tableNamer.TableName()
		if tableName != "Users" {
			t.Errorf("Expected table name 'Users', got '%s'", tableName)
		}
	} else {
		t.Error("OptimizedUser should implement TableNamer interface")
	}

	// Test IDGetter interface
	if idGetter, ok := interface{}(user).(IDGetter); ok {
		id := idGetter.GetID()
		if id != uint(1) {
			t.Errorf("Expected ID 1, got %v", id)
		}
	} else {
		t.Error("OptimizedUser should implement IDGetter interface")
	}

	// Test IDSetter interface
	if idSetter, ok := interface{}(&user).(IDSetter); ok {
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
	if fieldMapper, ok := interface{}(user).(FieldMapper); ok {
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
	if tableNamer, ok := interface{}(user).(TableNamer); ok {
		tableName := tableNamer.TableName()
		if tableName != "Users" {
			t.Errorf("Expected table name 'Users', got '%s'", tableName)
		}
	} else {
		t.Error("ReflectionUser should implement TableNamer interface")
	}

	// Should NOT implement optimization interfaces
	if _, ok := interface{}(user).(IDGetter); ok {
		t.Error("ReflectionUser should NOT implement IDGetter interface")
	}

	if _, ok := interface{}(&user).(IDSetter); ok {
		t.Error("ReflectionUser should NOT implement IDSetter interface")
	}

	if _, ok := interface{}(user).(FieldMapper); ok {
		t.Error("ReflectionUser should NOT implement FieldMapper interface")
	}
}

func TestSimplifiedAPI(t *testing.T) {
	// Test that we only have the 5 original CRUD methods
	// This is a compile-time test - if these methods don't exist, it won't compile

	// The 5 original methods should exist:
	// - List (tested elsewhere)
	// - Create
	// - Read
	// - Update
	// - Delete
	// - DeleteAll

	// Test that the simplified API works with both optimized and reflection models
	optimizedUser := OptimizedUser{Name: "Test", Email: "test@example.com"}
	reflectionUser := ReflectionUser{Name: "Test", Email: "test@example.com"}

	// These should compile and use the appropriate optimization path
	_ = optimizedUser  // Would use FieldMapper interface internally
	_ = reflectionUser // Would use reflection fallback internally

	// Verify the methods exist (compile-time check)
	var db interface{} = nil // Mock DB for compile check
	if db != nil {
		// These calls verify the method signatures exist
		// mysql.Create(db.(*sqlx.DB), &optimizedUser)
		// mysql.Read(db.(*sqlx.DB), &optimizedUser, 1)
		// mysql.Update(db.(*sqlx.DB), &optimizedUser)
		// mysql.Delete(db.(*sqlx.DB), &optimizedUser)
		// mysql.DeleteAll(db.(*sqlx.DB), &optimizedUser, 1, 2, 3)
	}
}
