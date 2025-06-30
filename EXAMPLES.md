# Sincap-Common Complete Guide

Comprehensive guide for using sincap-common with sqlx for high-performance database operations, including GORM-like preload functionality.

## 🚀 Quick Start

```go
import (
    "github.com/filllabs/sincap-common/db"
    "github.com/filllabs/sincap-common/db/mysql"
    "github.com/filllabs/sincap-common/db/queryapi"
    "github.com/filllabs/sincap-common/middlewares/qapi"
)

// Configure database connections
configs := []db.Config{
    {
        Name: "default",
        Args: []string{"user:password@tcp(localhost:3306)/database?parseTime=true"},
    },
}

db.Configure(configs)
database := db.DB()
```

## 📝 Model Definition

### Optimized Model with Relationships (Recommended)

Define models with performance interfaces and GORM-like relationship tags:

```go
type User struct {
    ID       uint   `db:"ID"`
    Name     string `db:"Name"`
    Email    string `db:"Email"`
    Age      int    `db:"Age"`
    
    // One-to-One relationship
    ProfileID uint     `db:"ProfileID"`
    Profile   *Profile `db:"-" join:"one2one,table:Profile,foreign_key:UserID"`
    
    // One-to-Many relationship  
    Orders    []*Order `db:"-" join:"one2many,table:Order,foreign_key:UserID"`
    
    // Many-to-Many relationship
    Tags      []*Tag   `db:"-" join:"many2many,table:Tag,through:UserTag,pivot_local_key:UserID,pivot_foreign_key:TagID"`
    
    // Polymorphic relationship
    Comments  []*Comment `db:"-" join:"polymorphic,table:Comment,id:CommentableID,type:CommentableType,value:User"`
}

// Performance interfaces - eliminates reflection
func (User) TableName() string {
    return "User"
}

func (u User) GetID() interface{} {
    return u.ID
}

func (u *User) SetID(id interface{}) error {
    if idVal, ok := id.(uint64); ok {
        u.ID = uint(idVal)
        return nil
    }
    return fmt.Errorf("invalid ID type")
}

func (u User) GetFieldMap() map[string]interface{} {
    return map[string]interface{}{
        "ID":    u.ID,
        "Name":  u.Name,
        "Email": u.Email,
        "Age":   u.Age,
    }
}
```

### Simple Model

Models without interfaces still work perfectly:

```go
type Product struct {
    ID    uint    `db:"ID"`
    Name  string  `db:"Name"`
    Price float64 `db:"Price"`
}

func (Product) TableName() string {
    return "Product"
}
```

## 🔗 Relationship Configuration

### Three-Tier Approach for Maximum Flexibility

The system provides three ways to define relationships, with automatic fallback:

1. **Struct Tags** (Recommended) - GORM-like relationship definitions
2. **Custom JoinRegistry** - For complex relationships  
3. **Automatic Joins** - Simple naming convention fallback

### 1. Struct Tag-Based Relationships (Recommended)

Define relationships directly in your struct tags, similar to GORM:

#### One-to-One
```go
Profile *Profile `db:"-" join:"one2one,table:Profile,foreign_key:UserID"`
```
**Generated SQL:**
```sql
LEFT JOIN Profile ON User.ID = Profile.UserID
```

#### One-to-Many
```go
Orders []*Order `db:"-" join:"one2many,table:Order,foreign_key:UserID"`
```
**Generated SQL:**
```sql
LEFT JOIN Order ON User.ID = Order.UserID
```

#### Many-to-Many
```go
Tags []*Tag `db:"-" join:"many2many,table:Tag,through:UserTag,pivot_local_key:UserID,pivot_foreign_key:TagID"`
```
**Generated SQL:**
```sql
LEFT JOIN UserTag ON User.ID = UserTag.UserID
LEFT JOIN Tag ON UserTag.TagID = Tag.ID
```

#### Polymorphic
```go
Comments []*Comment `db:"-" join:"polymorphic,table:Comment,id:CommentableID,type:CommentableType,value:User"`
```
**Generated SQL:**
```sql
LEFT JOIN Comment ON User.ID = Comment.CommentableID
WHERE Comment.CommentableType = 'User'
```

#### Tag Parameters Reference

- `table`: Target table name (auto-inferred from field type if not specified)
- `local_key`: Local table key (defaults to "ID")
- `foreign_key`: Foreign table key (defaults to FieldName + "ID")
- `through` / `pivot_table`: Junction table for many-to-many
- `pivot_local_key`: Local key in junction table (defaults to "UserID")
- `pivot_foreign_key`: Foreign key in junction table (defaults to FieldName + "ID")
- `id` / `polymorphic_id`: Polymorphic ID field
- `type` / `polymorphic_type`: Polymorphic type field
- `value` / `polymorphic_value`: Polymorphic type value
- `join_type`: JOIN type (LEFT, INNER, RIGHT - defaults to LEFT)

### 2. Custom JoinRegistry (Advanced)

For complex relationships or when you need full control:

```go
func CreateUserJoinRegistry() *queryapi.JoinRegistry {
    registry := queryapi.NewJoinRegistry()
    
    // Complex One-to-One with custom table name
    registry.Register("Profile", queryapi.JoinConfig{
        Type:       queryapi.OneToOne,
        Table:      "UserProfile",      // Custom table name
        LocalKey:   "ID",
        ForeignKey: "UserID",
        JoinType:   queryapi.InnerJoin, // INNER JOIN instead of LEFT
    })
    
    // Many-to-Many with custom junction table
    registry.Register("Role", queryapi.JoinConfig{
        Type:            queryapi.ManyToMany,
        Table:           "Role",
        PivotTable:      "UserRoleMapping", // Custom junction table
        PivotLocalKey:   "UserID",
        PivotForeignKey: "RoleID",
        JoinType:        queryapi.LeftJoin,
    })
    
    return registry
}
```

### 3. Automatic Joins (Fallback)

When no struct tags or custom JoinRegistry is provided, the system uses naming conventions:

**Automatic join assumptions:**
- Relationship Type: Always `OneToMany`
- Table Name: Same as preload name (e.g., "Profile", "Order")
- Local Key: Always "ID"
- Foreign Key: PreloadName + "ID" (e.g., "ProfileID", "OrderID")
- Join Type: Always `LEFT JOIN`

### When to Use Each Approach

- **Struct Tags**: GORM-like simplicity, self-documenting code, recommended for most cases
- **Custom JoinRegistry**: Complex relationships, runtime configuration, override struct tags
- **Automatic Joins**: Simple conventions, prototyping, quick development

## 📖 CRUD Operations

### Create

```go
// Create a user (uses FieldMapper interface for optimization)
user := &User{
    Name:  "John Doe",
    Email: "john@example.com",
    Age:   30,
}

err := mysql.Create(database, user)
if err != nil {
    log.Printf("Create error: %v", err)
} else {
    fmt.Printf("Created user with ID: %d\n", user.ID)
}
```

### Read

```go
// Read a user by ID (uses TableName interface for optimization)
var user User
err := mysql.Read(database, &user, 1)
if err != nil {
    log.Printf("Read error: %v", err)
} else {
    fmt.Printf("Read user: %+v\n", user)
}

// Read with preloads (uses struct tags automatically)
var userWithRelations User
err := mysql.Read(database, &userWithRelations, 1, "Profile", "Orders")
```

### Update

```go
// Update full record (uses FieldMapper + IDGetter interfaces)
user.Age = 31
err := mysql.Update(database, &user)

// Partial update with field map (uses IDGetter interface)
err := mysql.Update(database, &User{ID: 1}, map[string]any{
    "Email": "john.doe@example.com",
    "Age":   32,
})
```

### Delete

```go
// Delete single record
err := mysql.Delete(database, &user)

// Bulk delete by IDs
err := mysql.DeleteAll(database, &User{}, 2, 3, 4)
```

## 🔍 Query Examples

### Basic Filtering

```go
var users []User
query := &qapi.Query{
    Filter: []qapi.Filter{
        {Name: "Age", Value: "25", Operation: qapi.GT},
        {Name: "Name", Value: "John", Operation: qapi.LK},
    },
    Sort:   []string{"Name ASC"},
    Limit:  10,
    Offset: 0,
}

count, err := mysql.List(database, &users, query)
```

### Query with Relationships (Using Struct Tags)

```go
// Simple - struct tags handle everything automatically
query := &qapi.Query{
    Preloads: []string{"Profile", "Orders", "Tags"},
    Filter: []qapi.Filter{
        {Name: "Profile.Bio", Value: "developer", Operation: qapi.LK},
        {Name: "Orders.Status", Value: "active", Operation: qapi.EQ},
        {Name: "Orders.Total", Value: "100", Operation: qapi.GT},
    },
    Limit: 10,
}

var users []User
count, err := mysql.List(database, &users, query)
```

### Query with Custom JoinRegistry

```go
// Override struct tags with custom registry
query := &qapi.Query{
    Preloads:     []string{"Profile", "Role"},
    JoinRegistry: CreateUserJoinRegistry(), // Custom configuration
    Filter: []qapi.Filter{
        {Name: "Role.Name", Value: "admin", Operation: qapi.EQ},
    },
    Limit: 10,
}

var users []User
count, err := mysql.List(database, &users, query)
```

### Advanced Query with Multiple Filters

```go
query := &qapi.Query{
    Filter: []qapi.Filter{
        {Name: "Age", Value: "18", Operation: qapi.GTE},
        {Name: "Age", Value: "65", Operation: qapi.LT},
        {Name: "Email", Value: "gmail.com", Operation: qapi.LK},
        {Name: "Status", Value: "active", Operation: qapi.EQ},
    },
    Sort:   []string{"Name ASC", "Age DESC"},
    Limit:  20,
    Offset: 0,
    Q:      []string{"developer"}, // Full-text search
}

var users []User
count, err := mysql.List(database, &users, query)
```

## 🔧 Advanced Examples

### Custom SQL Generation

```go
// Generate SQL without executing
query := &qapi.Query{
    Filter: []qapi.Filter{
        {Name: "Status", Value: "active", Operation: qapi.EQ},
    },
    Sort: []string{"CreatedAt DESC"},
}

result, err := queryapi.GenerateSQL(query, User{})
if err == nil {
    fmt.Printf("SQL: %s\n", result.Query)
    fmt.Printf("Args: %v\n", result.Args)
    
    // Use with sqlx directly
    var users []User
    err = database.Select(&users, result.Query, result.Args...)
}
```

### Transactions

```go
// Begin transaction
tx, err := database.Beginx()
if err != nil {
    return err
}
defer tx.Rollback()

// Use tx instead of database for operations
user := &User{Name: "Jane", Email: "jane@example.com"}
err = mysql.Create(tx, user)
if err != nil {
    return err
}

// Update in same transaction
err = mysql.Update(tx, &User{ID: user.ID}, map[string]any{
    "Status": "verified",
})
if err != nil {
    return err
}

// Commit transaction
return tx.Commit()
```

### Complex Raw SQL Queries

```go
// For very complex queries, use sqlx directly
var results []struct {
    Username   string  `db:"Username"`
    OrderCount int     `db:"OrderCount"`
    TotalSpent float64 `db:"TotalSpent"`
}

query := `
    SELECT 
        u.Name as Username,
        COUNT(o.ID) as OrderCount,
        SUM(o.Total) as TotalSpent
    FROM User u
    LEFT JOIN Order o ON u.ID = o.UserID
    WHERE u.CreatedAt > ?
    GROUP BY u.ID
    HAVING OrderCount > 0
    ORDER BY TotalSpent DESC
    LIMIT 10
`

err := database.Select(&results, query, time.Now().AddDate(0, -1, 0))
```

## 🔄 Migration from GORM

### Before (GORM)
```go
type User struct {
    ID         uint        `gorm:"primary_key"`
    Profile    Profile     `gorm:"foreignKey:UserID"`
    Orders     []Order     `gorm:"foreignKey:UserID"`
    Tags       []*Tag      `gorm:"many2many:user_tags;"`
}

// Query
db.Preload("Profile").Preload("Orders").Find(&users)
db.Joins("Profile").Where("Profile.Bio LIKE ?", "%developer%").Find(&users)
```

### After (sqlx with struct tags)
```go
type User struct {
    ID       uint     `db:"ID"`
    Profile  *Profile `db:"-" join:"one2one,foreign_key:UserID"`
    Orders   []*Order `db:"-" join:"one2many,foreign_key:UserID"`
    Tags     []*Tag   `db:"-" join:"many2many,through:UserTag"`
}

// Query - much cleaner!
query := &qapi.Query{
    Preloads: []string{"Profile", "Orders"},
    Filter: []qapi.Filter{
        {Name: "Profile.Bio", Value: "developer", Operation: qapi.LK},
    },
}
count, err := mysql.List(DB, &users, query)
```

## 💡 Best Practices

### 1. Always Implement Performance Interfaces

```go
// For frequently used models, implement all interfaces
func (Model) TableName() string { ... }
func (m Model) GetID() interface{} { ... }
func (m *Model) SetID(id interface{}) error { ... }
func (m Model) GetFieldMap() map[string]interface{} { ... }
```

### 2. Use Struct Tags for Most Cases

```go
// Recommended approach - self-documenting and GORM-like
type User struct {
    ID      uint     `db:"ID"`
    Profile *Profile `db:"-" join:"one2one,foreign_key:UserID"`
    Orders  []*Order `db:"-" join:"one2many,foreign_key:UserID"`
}
```

### 3. Choose the Right Query Method

```go
// For simple queries without relationships
mysql.List(db, &users, query)

// For queries with struct tag relationships (recommended)
query.Preloads = []string{"Profile", "Orders"}
mysql.List(db, &users, query)

// For complex custom relationships
query.JoinRegistry = customRegistry
mysql.List(db, &users, query)

// For very complex custom queries
db.Select(&results, customSQL, args...)
```

### 4. Handle Errors Properly

```go
import "database/sql"

err := mysql.Read(db, &user, id)
if errors.Is(err, sql.ErrNoRows) {
    // Handle not found
    return nil, fmt.Errorf("user not found")
} else if err != nil {
    // Handle other errors
    return nil, fmt.Errorf("database error: %w", err)
}
```

### 5. Use Transactions for Related Operations

```go
tx, err := db.Beginx()
if err != nil {
    return err
}
defer tx.Rollback()

// Multiple related operations
mysql.Create(tx, &user)
mysql.Create(tx, &profile)

return tx.Commit()
```

### 6. Set Up Join Registry Once (When Needed)

```go
// Create at application startup and reuse for complex cases
var globalJoinRegistry = setupJoinRegistry()

func setupJoinRegistry() *queryapi.JoinRegistry {
    registry := queryapi.NewJoinRegistry()
    
    registry.Register("Profile", queryapi.JoinConfig{
        Type:       queryapi.OneToOne,
        Table:      "Profile",
        LocalKey:   "ID",
        ForeignKey: "UserID",
    })
    
    return registry
}
```

### 7. Naming Conventions

Use singular, PascalCase for table and column names:
```go
type User struct {
    ID        uint   `db:"ID"`
    ProfileID uint   `db:"ProfileID"`
}
```

### 8. Always Use db:"-" for Relationship Fields

```go
// Correct - excludes from direct database mapping
Profile *Profile `db:"-" join:"one2one,foreign_key:UserID"`

// Incorrect - would try to map as database column
Profile *Profile `join:"one2one,foreign_key:UserID"`
```

## 🎯 Key Features

- **GORM-like Preloads**: Familiar syntax with `Preloads: []string{"Profile", "Orders"}`
- **Struct Tag Relationships**: Self-documenting relationship definitions
- **Three-Tier Fallback**: Automatic fallback from custom → struct tags → automatic
- **High Performance**: Optimized interfaces eliminate reflection overhead
- **Type Safety**: Compile-time relationship validation
- **Translation Support**: Seamless integration with translation system
- **Clean API**: Simple function signatures with optional parameters

## 🔗 Related Documentation

- **[README.md](./README.md)** - Quick start and API overview
- **[MIGRATION.md](./MIGRATION.md)** - Complete GORM to sqlx migration guide