package queryapi

import (
	"fmt"
	"strings"
)

// JoinType represents the type of SQL join
type JoinType string

const (
	InnerJoin JoinType = "INNER JOIN"
	LeftJoin  JoinType = "LEFT JOIN"
	RightJoin JoinType = "RIGHT JOIN"
)

// RelationType represents the type of relationship
type RelationType string

const (
	OneToOne    RelationType = "one_to_one"
	OneToMany   RelationType = "one_to_many"
	ManyToMany  RelationType = "many_to_many"
	Polymorphic RelationType = "polymorphic"
)

// JoinConfig defines how to join tables for relationship queries
type JoinConfig struct {
	// Relationship type
	Type RelationType `json:"type"`

	// Target table to join
	Table string `json:"table"`

	// Join type (INNER, LEFT, RIGHT)
	JoinType JoinType `json:"join_type,omitempty"`

	// For direct relationships (one_to_one, one_to_many)
	LocalKey   string `json:"local_key,omitempty"`   // e.g., "id"
	ForeignKey string `json:"foreign_key,omitempty"` // e.g., "user_id"

	// For many_to_many relationships
	PivotTable      string `json:"pivot_table,omitempty"`       // e.g., "user_roles"
	PivotLocalKey   string `json:"pivot_local_key,omitempty"`   // e.g., "user_id"
	PivotForeignKey string `json:"pivot_foreign_key,omitempty"` // e.g., "role_id"

	// For polymorphic relationships
	PolymorphicType  string `json:"polymorphic_type,omitempty"`  // e.g., "commentable_type"
	PolymorphicID    string `json:"polymorphic_id,omitempty"`    // e.g., "commentable_id"
	PolymorphicValue string `json:"polymorphic_value,omitempty"` // e.g., "User"
}

// JoinRegistry holds join configurations for different field paths
type JoinRegistry struct {
	joins map[string]JoinConfig
}

// NewJoinRegistry creates a new join registry
func NewJoinRegistry() *JoinRegistry {
	return &JoinRegistry{
		joins: make(map[string]JoinConfig),
	}
}

// Register adds a join configuration for a field path
func (jr *JoinRegistry) Register(fieldPath string, config JoinConfig) {
	// Set default join type if not specified
	if config.JoinType == "" {
		config.JoinType = LeftJoin
	}
	jr.joins[fieldPath] = config
}

// Get retrieves a join configuration for a field path
func (jr *JoinRegistry) Get(fieldPath string) (JoinConfig, bool) {
	config, exists := jr.joins[fieldPath]
	return config, exists
}

// GenerateJoinSQL generates the JOIN clause and WHERE conditions for a relationship
func (jr *JoinRegistry) GenerateJoinSQL(fieldPath, baseTable string) (joinClause string, whereClause string, err error) {
	config, exists := jr.joins[fieldPath]
	if !exists {
		return "", "", fmt.Errorf("no join configuration found for field path: %s", fieldPath)
	}

	switch config.Type {
	case OneToOne, OneToMany:
		return jr.generateDirectJoin(baseTable, config)
	case ManyToMany:
		return jr.generateManyToManyJoin(baseTable, config)
	case Polymorphic:
		return jr.generatePolymorphicJoin(baseTable, config)
	default:
		return "", "", fmt.Errorf("unsupported relationship type: %s", config.Type)
	}
}

// generateDirectJoin creates JOIN for one-to-one and one-to-many relationships
func (jr *JoinRegistry) generateDirectJoin(baseTable string, config JoinConfig) (string, string, error) {
	if config.LocalKey == "" || config.ForeignKey == "" {
		return "", "", fmt.Errorf("local_key and foreign_key are required for direct relationships")
	}

	joinClause := fmt.Sprintf("%s %s ON %s.%s = %s.%s",
		config.JoinType,
		safeMySQLNaming(config.Table),
		safeMySQLNaming(baseTable),
		safeMySQLNaming(config.LocalKey),
		safeMySQLNaming(config.Table),
		safeMySQLNaming(config.ForeignKey))

	return joinClause, "", nil
}

// generateManyToManyJoin creates JOIN for many-to-many relationships
func (jr *JoinRegistry) generateManyToManyJoin(baseTable string, config JoinConfig) (string, string, error) {
	if config.PivotTable == "" || config.PivotLocalKey == "" || config.PivotForeignKey == "" {
		return "", "", fmt.Errorf("pivot_table, pivot_local_key, and pivot_foreign_key are required for many-to-many relationships")
	}

	// Generate two JOINs: base -> pivot -> target
	joinClauses := []string{
		fmt.Sprintf("%s %s ON %s.%s = %s.%s",
			config.JoinType,
			safeMySQLNaming(config.PivotTable),
			safeMySQLNaming(baseTable),
			safeMySQLNaming("ID"),
			safeMySQLNaming(config.PivotTable),
			safeMySQLNaming(config.PivotLocalKey)),
		fmt.Sprintf("%s %s ON %s.%s = %s.%s",
			config.JoinType,
			safeMySQLNaming(config.Table),
			safeMySQLNaming(config.PivotTable),
			safeMySQLNaming(config.PivotForeignKey),
			safeMySQLNaming(config.Table),
			safeMySQLNaming("ID")),
	}

	return strings.Join(joinClauses, " "), "", nil
}

// generatePolymorphicJoin creates JOIN for polymorphic relationships
func (jr *JoinRegistry) generatePolymorphicJoin(baseTable string, config JoinConfig) (string, string, error) {
	if config.PolymorphicID == "" || config.PolymorphicType == "" || config.PolymorphicValue == "" {
		return "", "", fmt.Errorf("polymorphic_id, polymorphic_type, and polymorphic_value are required for polymorphic relationships")
	}

	joinClause := fmt.Sprintf("%s %s ON %s.%s = %s.%s",
		config.JoinType,
		safeMySQLNaming(config.Table),
		safeMySQLNaming(baseTable),
		safeMySQLNaming("ID"),
		safeMySQLNaming(config.Table),
		safeMySQLNaming(config.PolymorphicID))

	whereClause := fmt.Sprintf("%s.%s = '%s'",
		safeMySQLNaming(config.Table),
		safeMySQLNaming(config.PolymorphicType),
		config.PolymorphicValue)

	return joinClause, whereClause, nil
}

// BuildJoinQuery builds a complete query with joins for relationship filtering
func (jr *JoinRegistry) BuildJoinQuery(baseQuery, baseTable string, fieldPaths []string) (string, []string, error) {
	var joinClauses []string
	var whereClauses []string

	for _, fieldPath := range fieldPaths {
		joinClause, whereClause, err := jr.GenerateJoinSQL(fieldPath, baseTable)
		if err != nil {
			return "", nil, err
		}

		if joinClause != "" {
			joinClauses = append(joinClauses, joinClause)
		}

		if whereClause != "" {
			whereClauses = append(whereClauses, whereClause)
		}
	}

	// Build the complete query
	query := baseQuery
	if len(joinClauses) > 0 {
		query += " " + strings.Join(joinClauses, " ")
	}

	return query, whereClauses, nil
}
