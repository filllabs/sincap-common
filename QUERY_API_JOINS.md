# Query API with Explicit Joins

The Query API has been updated to support explicit join configurations instead of relying on GORM tags. This provides more control and flexibility for relationship queries while removing GORM dependencies.

## Overview

The new system uses a **JoinRegistry** to define how tables should be joined for relationship queries. Instead of parsing GORM tags, you explicitly configure joins for each relationship.

**Note**: This documentation assumes PascalCase naming convention for database schemas (e.g., `ID`, `Name`, `UserID`, `CityID`).

## Basic Usage

### Without Relationships (No Changes)

Simple queries without relationships work exactly the same:

```go
import "github.com/filllabs/sincap-common/db/queryapi"

// Simple filtering - no changes needed
query := &qapi.Query{
    Filter: []qapi.Filter{
        {Name: "Name", Value: "John", Operation: qapi.EQ},
        {Name: "Age", Value: "25", Operation: qapi.GT},
    },
    Limit: 10,
}

result, err := queryapi.GenerateSQL(query, User{})
```

### With Relationships (New Join System)

For relationship queries, you need to set up a join registry:

```go
import "github.com/filllabs/sincap-common/db/queryapi"

// 1. Create a join registry
joinRegistry := queryapi.NewJoinRegistry()

// 2. Register your relationships (using PascalCase naming)
joinRegistry.Register("Profile", queryapi.JoinConfig{
    Type:       queryapi.OneToOne,
    Table:      "Profiles",
    LocalKey:   "ID",       // Users.ID
    ForeignKey: "UserID",   // Profiles.UserID
    JoinType:   queryapi.LeftJoin,
})

joinRegistry.Register("Roles", queryapi.JoinConfig{
    Type:            queryapi.ManyToMany,
    Table:           "Roles",
    PivotTable:      "UserRoles",
    PivotLocalKey:   "UserID",   // UserRoles.UserID
    PivotForeignKey: "RoleID",   // UserRoles.RoleID
    JoinType:        queryapi.LeftJoin,
})

// 3. Use the registry in your query
query := &qapi.Query{
    Filter: []qapi.Filter{
        {Name: "Profile.Bio", Value: "developer", Operation: qapi.LK},
        {Name: "Roles.Name", Value: "admin", Operation: qapi.EQ},
    },
}

options := &queryapi.QueryOptions{
    JoinRegistry: joinRegistry,
}

result, err := queryapi.GenerateSQLWithOptions(query, User{}, options)
```

## Relationship Types

### 1. One-to-One Relationships

```go
joinRegistry.Register("Profile", queryapi.JoinConfig{
    Type:       queryapi.OneToOne,
    Table:      "Profiles",
    LocalKey:   "ID",        // Users.ID
    ForeignKey: "UserID",    // Profiles.UserID
    JoinType:   queryapi.LeftJoin,
})
```

**Generated SQL:**
```sql
SELECT * FROM Users 
LEFT JOIN Profiles ON Users.ID = Profiles.UserID 
WHERE Profiles.Bio LIKE ?
```

### 2. One-to-Many Relationships

```go
joinRegistry.Register("Posts", queryapi.JoinConfig{
    Type:       queryapi.OneToMany,
    Table:      "Posts",
    LocalKey:   "ID",        // Users.ID
    ForeignKey: "UserID",    // Posts.UserID
    JoinType:   queryapi.LeftJoin,
})
```

**Generated SQL:**
```sql
SELECT * FROM Users 
LEFT JOIN Posts ON Users.ID = Posts.UserID 
WHERE Posts.Title LIKE ?
```

### 3. Many-to-Many Relationships

```go
joinRegistry.Register("Roles", queryapi.JoinConfig{
    Type:            queryapi.ManyToMany,
    Table:           "Roles",
    PivotTable:      "UserRoles",
    PivotLocalKey:   "UserID",   // UserRoles.UserID
    PivotForeignKey: "RoleID",   // UserRoles.RoleID
    JoinType:        queryapi.LeftJoin,
})
```

**Generated SQL:**
```sql
SELECT * FROM Users 
LEFT JOIN UserRoles ON Users.ID = UserRoles.UserID 
LEFT JOIN Roles ON UserRoles.RoleID = Roles.ID 
WHERE Roles.Name = ?
```

### 4. Polymorphic Relationships

```go
joinRegistry.Register("Comments", queryapi.JoinConfig{
    Type:             queryapi.Polymorphic,
    Table:            "Comments",
    PolymorphicID:    "CommentableID",    // Comments.CommentableID
    PolymorphicType:  "CommentableType",  // Comments.CommentableType
    PolymorphicValue: "User",
    JoinType:         queryapi.LeftJoin,
})
```

**Generated SQL:**
```sql
SELECT * FROM Users 
LEFT JOIN Comments ON Users.ID = Comments.CommentableID 
WHERE Comments.Content LIKE ? AND Comments.CommentableType = 'User'
```

## Complete Example

```go
package main

import (
    "log"
    "github.com/filllabs/sincap-common/db/queryapi"
    "github.com/filllabs/sincap-common/middlewares/qapi"
)

type User struct {
    ID   int    `json:"id" db:"ID"`
    Name string `json:"name" db:"Name"`
}

func (User) TableName() string {
    return "Users"
}

func main() {
    // Set up join registry
    joinRegistry := queryapi.NewJoinRegistry()
    
    // One-to-one: User -> Profile
    joinRegistry.Register("Profile", queryapi.JoinConfig{
        Type:       queryapi.OneToOne,
        Table:      "Profiles",
        LocalKey:   "ID",
        ForeignKey: "UserID",
    })
    
    // One-to-many: User -> Posts
    joinRegistry.Register("Posts", queryapi.JoinConfig{
        Type:       queryapi.OneToMany,
        Table:      "Posts",
        LocalKey:   "ID",
        ForeignKey: "UserID",
    })
    
    // Many-to-many: User <-> Roles
    joinRegistry.Register("Roles", queryapi.JoinConfig{
        Type:            queryapi.ManyToMany,
        Table:           "Roles",
        PivotTable:      "UserRoles",
        PivotLocalKey:   "UserID",
        PivotForeignKey: "RoleID",
    })
    
    // Complex query with relationships
    query := &qapi.Query{
        Filter: []qapi.Filter{
            {Name: "Name", Value: "John", Operation: qapi.LK},
            {Name: "Profile.Bio", Value: "developer", Operation: qapi.LK},
            {Name: "Posts.Status", Value: "published", Operation: qapi.EQ},
            {Name: "Roles.Name", Value: "admin", Operation: qapi.EQ},
        },
        Sort:   []string{"Name ASC", "Profile.CreatedAt DESC"},
        Limit:  20,
        Offset: 0,
    }
    
    options := &queryapi.QueryOptions{
        JoinRegistry: joinRegistry,
    }
    
    result, err := queryapi.GenerateSQLWithOptions(query, User{}, options)
    if err != nil {
        log.Fatal(err)
    }
    
    log.Printf("Query: %s", result.Query)
    log.Printf("Args: %v", result.Args)
    log.Printf("Count Query: %s", result.CountQuery)
}
```

## Migration from GORM Tags

### Before (with GORM tags):
```go
type User struct {
    ID      int     `json:"id"`
    Name    string  `json:"name"`
    Profile Profile `gorm:"foreignKey:UserID"`
    Roles   []Role  `gorm:"many2many:user_roles"`
}

// Query worked automatically based on tags
query := &qapi.Query{
    Filter: []qapi.Filter{
        {Name: "Profile.Bio", Value: "developer", Operation: qapi.LK},
    },
}
result, err := queryapi.GenerateSQL(query, User{})
```

### After (with explicit joins):
```go
type User struct {
    ID   int    `json:"id" db:"ID"`
    Name string `json:"name" db:"Name"`
    // No GORM tags needed
}

// Set up joins explicitly
joinRegistry := queryapi.NewJoinRegistry()
joinRegistry.Register("Profile", queryapi.JoinConfig{
    Type:       queryapi.OneToOne,
    Table:      "Profiles",
    LocalKey:   "ID",
    ForeignKey: "UserID",
})

// Use with options
query := &qapi.Query{
    Filter: []qapi.Filter{
        {Name: "Profile.Bio", Value: "developer", Operation: qapi.LK},
    },
}
options := &queryapi.QueryOptions{JoinRegistry: joinRegistry}
result, err := queryapi.GenerateSQLWithOptions(query, User{}, options)
```

## Backward Compatibility

The old `GenerateSQL` function still works for queries without relationships:

```go
// This still works for simple queries
result, err := queryapi.GenerateSQL(query, User{})

// But relationship queries will fall back to subqueries instead of joins
// (less efficient but still functional)
```

## Join Types

- `queryapi.InnerJoin` - Only returns records with matches in both tables
- `queryapi.LeftJoin` - Returns all records from left table, matched records from right
- `queryapi.RightJoin` - Returns all records from right table, matched records from left

Default is `LeftJoin` if not specified.

## Best Practices

1. **Set up join registry once** - Create it at application startup and reuse
2. **Use appropriate join types** - `LeftJoin` for optional relationships, `InnerJoin` for required
3. **Register only needed relationships** - Don't register joins you don't query
4. **Test your joins** - Verify the generated SQL matches your expectations
5. **Consider performance** - Joins are more efficient than subqueries for large datasets
6. **Use PascalCase consistently** - Match your database naming convention

## Performance Comparison

### With Joins (New System):
```sql
SELECT * FROM Users 
LEFT JOIN Profiles ON Users.ID = Profiles.UserID 
WHERE Profiles.Bio LIKE 'developer%'
```

### Without Joins (Fallback):
```sql
SELECT * FROM Users 
WHERE Users.ID IN (
    SELECT Profiles.UserID FROM Profiles 
    WHERE Profiles.Bio LIKE 'developer%'
)
```

The join-based approach is typically more efficient, especially for complex queries with multiple relationships.

## Error Handling

```go
result, err := queryapi.GenerateSQLWithOptions(query, User{}, options)
if err != nil {
    // Common errors:
    // - "no join configuration found for field path: Profile"
    // - "LocalKey and ForeignKey are required for direct relationships"
    // - "PivotTable, PivotLocalKey, and PivotForeignKey are required for many-to-many"
    log.Printf("Query generation failed: %v", err)
}
```

This new system provides much more control and flexibility while removing the dependency on GORM tags! 