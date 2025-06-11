package queryapi

import (
	"testing"

	"github.com/filllabs/sincap-common/middlewares/qapi"
)

func TestJoinRegistry_OneToOne(t *testing.T) {
	registry := NewJoinRegistry()
	registry.Register("Profile", JoinConfig{
		Type:       OneToOne,
		Table:      "Profiles",
		LocalKey:   "ID",
		ForeignKey: "UserID",
		JoinType:   LeftJoin,
	})

	joinClause, whereClause, err := registry.GenerateJoinSQL("Profile", "Users")
	if err != nil {
		t.Errorf("GenerateJoinSQL() error = %v", err)
		return
	}

	expectedJoin := "LEFT JOIN `Profiles` ON `Users`.`ID` = `Profiles`.`UserID`"
	if joinClause != expectedJoin {
		t.Errorf("GenerateJoinSQL() joinClause = %v, want %v", joinClause, expectedJoin)
	}

	if whereClause != "" {
		t.Errorf("GenerateJoinSQL() whereClause = %v, want empty", whereClause)
	}
}

func TestJoinRegistry_ManyToMany(t *testing.T) {
	registry := NewJoinRegistry()
	registry.Register("Roles", JoinConfig{
		Type:            ManyToMany,
		Table:           "Roles",
		PivotTable:      "UserRoles",
		PivotLocalKey:   "UserID",
		PivotForeignKey: "RoleID",
		JoinType:        LeftJoin,
	})

	joinClause, whereClause, err := registry.GenerateJoinSQL("Roles", "Users")
	if err != nil {
		t.Errorf("GenerateJoinSQL() error = %v", err)
		return
	}

	expectedJoin := "LEFT JOIN `UserRoles` ON `Users`.`ID` = `UserRoles`.`UserID` LEFT JOIN `Roles` ON `UserRoles`.`RoleID` = `Roles`.`ID`"
	if joinClause != expectedJoin {
		t.Errorf("GenerateJoinSQL() joinClause = %v, want %v", joinClause, expectedJoin)
	}

	if whereClause != "" {
		t.Errorf("GenerateJoinSQL() whereClause = %v, want empty", whereClause)
	}
}

func TestJoinRegistry_Polymorphic(t *testing.T) {
	registry := NewJoinRegistry()
	registry.Register("Comments", JoinConfig{
		Type:             Polymorphic,
		Table:            "Comments",
		PolymorphicID:    "CommentableID",
		PolymorphicType:  "CommentableType",
		PolymorphicValue: "User",
		JoinType:         LeftJoin,
	})

	joinClause, whereClause, err := registry.GenerateJoinSQL("Comments", "Users")
	if err != nil {
		t.Errorf("GenerateJoinSQL() error = %v", err)
		return
	}

	expectedJoin := "LEFT JOIN `Comments` ON `Users`.`ID` = `Comments`.`CommentableID`"
	expectedWhere := "`Comments`.`CommentableType` = 'User'"

	if joinClause != expectedJoin {
		t.Errorf("GenerateJoinSQL() joinClause = %v, want %v", joinClause, expectedJoin)
	}

	if whereClause != expectedWhere {
		t.Errorf("GenerateJoinSQL() whereClause = %v, want %v", whereClause, expectedWhere)
	}
}

func TestJoinRegistry_NotFound(t *testing.T) {
	registry := NewJoinRegistry()

	_, _, err := registry.GenerateJoinSQL("NonExistent", "Users")
	if err == nil {
		t.Error("GenerateJoinSQL() expected error for non-existent field path")
	}

	expectedError := "no join configuration found for field path: NonExistent"
	if err.Error() != expectedError {
		t.Errorf("GenerateJoinSQL() error = %v, want %v", err.Error(), expectedError)
	}
}

func TestJoinRegistry_DefaultJoinType(t *testing.T) {
	registry := NewJoinRegistry()
	config := JoinConfig{
		Type:       OneToOne,
		Table:      "Profiles",
		LocalKey:   "ID",
		ForeignKey: "UserID",
		// JoinType not specified - should default to LeftJoin
	}
	registry.Register("Profile", config)

	// Verify the default was set
	storedConfig, exists := registry.Get("Profile")
	if !exists {
		t.Error("Expected config to exist")
	}

	if storedConfig.JoinType != LeftJoin {
		t.Errorf("Expected default JoinType to be LeftJoin, got %v", storedConfig.JoinType)
	}
}

func TestGenerateDBWithJoins(t *testing.T) {
	// Set up join registry
	registry := NewJoinRegistry()
	registry.Register("InnerF", JoinConfig{
		Type:       OneToOne,
		Table:      "Inner1",
		LocalKey:   "ID",
		ForeignKey: "SampleID",
	})

	// Create a query with relationship filter using actual fields from the test structs
	query := &qapi.Query{
		Filter: []qapi.Filter{
			{Name: "Name", Value: "John", Operation: qapi.EQ},
			{Name: "InnerF.Name", Value: "developer", Operation: qapi.LK}, // Use actual field from Inner1
		},
		Limit: 10,
	}

	options := &QueryOptions{
		JoinRegistry: registry,
	}

	result, err := GenerateDBWithOptions(query, Sample{}, options)
	if err != nil {
		t.Errorf("GenerateDBWithOptions() error = %v", err)
		return
	}

	// The query should contain a JOIN
	if result.Query == "" {
		t.Error("Expected non-empty query")
	}

	// Should have parameters for both filters
	if len(result.Args) < 2 {
		t.Errorf("Expected at least 2 arguments, got %d", len(result.Args))
	}

	t.Logf("Generated Query: %s", result.Query)
	t.Logf("Generated Args: %v", result.Args)
}

func TestGenerateDB_BackwardCompatibility(t *testing.T) {
	// Test that the old GenerateDB function still works for simple queries
	query := &qapi.Query{
		Filter: []qapi.Filter{
			{Name: "Name", Value: "John", Operation: qapi.EQ},
		},
		Limit: 10,
	}

	result, err := GenerateDB(query, Sample{})
	if err != nil {
		t.Errorf("GenerateDB() error = %v", err)
		return
	}

	if result.Query == "" {
		t.Error("Expected non-empty query")
	}

	// Should have 2 arguments: one for the filter value and one for the limit
	if len(result.Args) != 2 {
		t.Errorf("Expected 2 arguments (filter + limit), got %d", len(result.Args))
	}

	// Verify the arguments
	if result.Args[0] != "John" {
		t.Errorf("Expected first argument to be 'John', got %v", result.Args[0])
	}
	if result.Args[1] != 10 {
		t.Errorf("Expected second argument to be 10, got %v", result.Args[1])
	}

	t.Logf("Backward Compatible Query: %s", result.Query)
	t.Logf("Backward Compatible Args: %v", result.Args)
}
