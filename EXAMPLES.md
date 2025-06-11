# Sincap-Common Examples

Practical code examples for using sincap-common with sqlx for high-performance database operations.

## 🚀 Setup

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

## 📝 Model Examples

### Optimized Model (Recommended)

Implement performance interfaces for maximum efficiency:

```go
type User struct {
    ID    uint   `db:"ID"`
    Name  string `db:"Name"`
    Email string `db:"Email"`
    Age   int    `db:"Age"`
}

// TableNamer - eliminates reflection for table name
func (User) TableName() string {
    return "User"
}

// IDGetter - eliminates reflection for ID access
func (u User) GetID() interface{} {
    return u.ID
}

// IDSetter - eliminates reflection for ID setting
func (u *User) SetID(id interface{}) error {
    if idVal, ok := id.(uint64); ok {
        u.ID = uint(idVal)
        return nil
    }
    return fmt.Errorf("invalid ID type")
}

// FieldMapper - eliminates reflection for CRUD operations
func (u User) GetFieldMap() map[string]interface{} {
    return map[string]interface{}{
        "ID":    u.ID,
        "Name":  u.Name,
        "Email": u.Email,
        "Age":   u.Age,
    }
}
```

### Standard Model

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

## 📖 CRUD Examples

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
```

### Update (Full Record)

```go
// Update full record (uses FieldMapper + IDGetter interfaces)
user.Age = 31
err := mysql.Update(database, &user)
if err != nil {
    log.Printf("Update error: %v", err)
} else {
    fmt.Println("User updated successfully")
}
```

### Update (Partial)

```go
// Partial update with field map (uses IDGetter interface)
err := mysql.Update(database, &User{ID: 1}, map[string]any{
    "Email": "john.doe@example.com",
    "Age":   32,
})
if err != nil {
    log.Printf("Partial update error: %v", err)
} else {
    fmt.Println("User updated with field map")
}
```

### Delete

```go
// Delete single record (uses TableName + IDGetter interfaces)
err := mysql.Delete(database, &user)
if err != nil {
    log.Printf("Delete error: %v", err)
} else {
    fmt.Println("User deleted successfully")
}
```

### Bulk Delete

```go
// Delete multiple records by IDs (uses TableName interface)
err := mysql.DeleteAll(database, &User{}, 2, 3, 4)
if err != nil {
    log.Printf("Bulk delete error: %v", err)
} else {
    fmt.Println("Users deleted successfully")
}
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
if err != nil {
    log.Printf("List error: %v", err)
} else {
    fmt.Printf("Found %d users\n", count)
}
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

### Query with Relationships (Joins)

```go
// Set up join registry for relationship queries
joinRegistry := queryapi.NewJoinRegistry()
joinRegistry.Register("Profile", queryapi.JoinConfig{
    Type:       queryapi.OneToOne,
    Table:      "Profile",
    LocalKey:   "ID",
    ForeignKey: "UserID",
})

// Query with relationship filters
complexQuery := &qapi.Query{
    Filter: []qapi.Filter{
        {Name: "Name", Value: "John", Operation: qapi.LK},
        {Name: "Profile.Bio", Value: "developer", Operation: qapi.LK},
    },
    Sort:  []string{"Name ASC"},
    Limit: 20,
}

options := &queryapi.QueryOptions{
    JoinRegistry: joinRegistry,
}

result, err := queryapi.GenerateDBWithOptions(complexQuery, User{}, options)
if err != nil {
    log.Printf("GenerateDBWithOptions error: %v", err)
} else {
    fmt.Printf("Generated query: %s\n", result.Query)
    fmt.Printf("Query args: %v\n", result.Args)
}
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

## 🔗 Relationship Examples

### One-to-One Relationship

```go
joinRegistry := queryapi.NewJoinRegistry()
joinRegistry.Register("Profile", queryapi.JoinConfig{
    Type:       queryapi.OneToOne,
    Table:      "Profile",
    LocalKey:   "ID",        // User.ID
    ForeignKey: "UserID",    // Profile.UserID
})

query := &qapi.Query{
    Filter: []qapi.Filter{
        {Name: "Profile.Bio", Value: "developer", Operation: qapi.LK},
    },
}

options := &queryapi.QueryOptions{JoinRegistry: joinRegistry}
result, err := queryapi.GenerateDBWithOptions(query, User{}, options)
```

### Many-to-Many Relationship

```go
joinRegistry := queryapi.NewJoinRegistry()
joinRegistry.Register("Corporate", queryapi.JoinConfig{
    Type:            queryapi.ManyToMany,
    Table:           "Corporate",
    PivotTable:      "UserCorporate",
    PivotLocalKey:   "UserID",       // UserCorporate.UserID
    PivotForeignKey: "CorporateID",  // UserCorporate.CorporateID
})

query := &qapi.Query{
    Filter: []qapi.Filter{
        {Name: "Corporate.Name", Value: "TechCorp", Operation: qapi.EQ},
    },
}

options := &queryapi.QueryOptions{JoinRegistry: joinRegistry}
result, err := queryapi.GenerateDBWithOptions(query, User{}, options)
```

### Polymorphic Relationship

```go
joinRegistry := queryapi.NewJoinRegistry()
joinRegistry.Register("Comment", queryapi.JoinConfig{
    Type:             queryapi.Polymorphic,
    Table:            "Comment",
    PolymorphicID:    "CommentableID",
    PolymorphicType:  "CommentableType",
    PolymorphicValue: "User",
})

query := &qapi.Query{
    Filter: []qapi.Filter{
        {Name: "Comment.Content", Value: "great", Operation: qapi.LK},
    },
}

options := &queryapi.QueryOptions{JoinRegistry: joinRegistry}
result, err := queryapi.GenerateDBWithOptions(query, User{}, options)
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

### 2. Choose the Right Query Method

```go
// For simple queries without relationships
mysql.List(db, &users, query)

// For complex queries with joins
queryapi.GenerateDBWithOptions(query, User{}, options)

// For very complex custom queries
db.Select(&results, customSQL, args...)
```

### 3. Handle Errors Properly

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

### 4. Use Transactions for Related Operations

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

### 5. Set Up Join Registry Once

```go
// Create at application startup and reuse
var globalJoinRegistry = setupJoinRegistry()

func setupJoinRegistry() *queryapi.JoinRegistry {
    registry := queryapi.NewJoinRegistry()
    
    registry.Register("Profile", queryapi.JoinConfig{
        Type:       queryapi.OneToOne,
        Table:      "Profile",
        LocalKey:   "ID",
        ForeignKey: "UserID",
    })
    
    registry.Register("Corporate", queryapi.JoinConfig{
        Type:            queryapi.ManyToMany,
        Table:           "Corporate",
        PivotTable:      "UserCorporate",
        PivotLocalKey:   "UserID",
        PivotForeignKey: "CorporateID",
    })
    
    return registry
}
```

## 🔗 Related Documentation

- **[README.md](./README.md)** - Quick start and API overview
- **[MIGRATION.md](./MIGRATION.md)** - Complete GORM to sqlx migration guide