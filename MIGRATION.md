# Sincap-Common: Complete GORM to sqlx Migration Guide

## 🎯 Overview

This guide covers the complete migration from GORM to sqlx in the sincap-common library. The migration maintains the existing Query API functionality while replacing GORM's ORM features with manual SQL query building using sqlx.

**Note**: This documentation uses singular PascalCase naming convention for database schemas (e.g., table names: `User`, `City`, `Corporate`; column names: `ID`, `Name`, `UserID`, `CityID`).

## ✅ Migration Status: COMPLETE

All core functionality has been migrated and is working. The API remains the same for backward compatibility with significant performance improvements.

---

## 📋 What Changed

### ✅ **Still Works (Same API)**
- Query API (`qapi.Query`) - filters, sorting, pagination
- Repository pattern and interfaces  
- Service layer pattern
- Database connection management
- JSON field handling
- Polymorphic and Many-to-Many relationship queries
- All 5 CRUD methods: `List`, `Create`, `Read`, `Update`, `Delete`, `DeleteAll`

### 🔄 **Changed Implementation (Same API)**
- Manual SQL query construction instead of ORM
- Connection pooling with sqlx
- Performance optimizations with interface-based reflection reduction
- Explicit join configuration instead of GORM tags

### ❌ **No Longer Available**
- GORM Associations (`Preload`, `Joins`) - use explicit join configuration
- GORM Migrations (`AutoMigrate`) - use goose (or golang-migrate)
- GORM Hooks (`BeforeCreate`, `AfterUpdate`) - implement manually
- GORM Scopes - implement as functions
- Built-in soft delete support - implement manually

---

## 🚀 Migration Steps

### 1. Update Dependencies

```bash
# Remove GORM
go mod edit -dropreplace gorm.io/gorm
go mod edit -dropreplace gorm.io/driver/mysql

# Add sqlx and migration tools
go get github.com/jmoiron/sqlx
go get github.com/go-sql-driver/mysql
go get github.com/pressly/goose/v3

go mod tidy
```

### 2. Update Database Configuration

**Before (GORM):**
```go
db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
```

**After (sqlx):**
```go
configs := []db.Config{
    {
        Name: "default",
        Args: []string{"user:password@tcp(localhost:3306)/database?parseTime=true"},
    },
}
db.Configure(configs)
database := db.DB()
```

### 3. Update Model Definitions

**Before (GORM):**
```go
type User struct {
    ID      uint    `gorm:"primary_key"`
    CityID  *uint
	City    *cities.City `gorm:"foreignKey:CityID"`
    Corporates []Corporate  `gorm:"many2many:user_corporates"`
}
```

**After (sqlx):**
```go
type User struct {
    ID     uint  `db:"ID"`
    CityID *uint `db:"CityID"`
    // No GORM tags needed - relationships handled explicitly
}
```

### 4. Implement Performance Interfaces (Optional but Recommended)

```go
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

---

## 🔗 Relationship Migration

The new system uses a **JoinRegistry** to define how tables should be joined for relationship queries. This replaces GORM tags with explicit configuration.

### GORM vs sqlx Relationship Handling

**Before (GORM):**
```go
type User struct {
    ID         uint        `gorm:"primary_key"`
    Profile    Profile     `gorm:"foreignKey:UserID"`
    Orders     []Order     `gorm:"foreignKey:UserID"`
    Corporates []Corporate `gorm:"many2many:user_corporates"`
}

// Query with preloads
db.Preload("Profile").Preload("Orders").Find(&users)

// Query with joins
db.Joins("Profile").Where("Profile.Bio LIKE ?", "%developer%").Find(&users)
```

**After (sqlx):**
```go
type User struct {
    ID uint `db:"ID"`
    // No relationship fields - handled explicitly
}

// Set up join registry
joinRegistry := queryapi.NewJoinRegistry()

// One-to-one
joinRegistry.Register("Profile", queryapi.JoinConfig{
    Type:       queryapi.OneToOne,
    Table:      "Profile",
    LocalKey:   "ID",
    ForeignKey: "UserID",
})

// One-to-many
joinRegistry.Register("Order", queryapi.JoinConfig{
    Type:       queryapi.OneToMany,
    Table:      "Order",
    LocalKey:   "ID",
    ForeignKey: "UserID",
})

// Many-to-many
joinRegistry.Register("Corporate", queryapi.JoinConfig{
    Type:            queryapi.ManyToMany,
    Table:           "Corporate",
    PivotTable:      "UserCorporate",
    PivotLocalKey:   "UserID",
    PivotForeignKey: "CorporateID",
})

// Query with relationships
query := &qapi.Query{
    Filter: []qapi.Filter{
        {Name: "Profile.Bio", Value: "developer", Operation: qapi.LK},
    },
}

options := &queryapi.QueryOptions{JoinRegistry: joinRegistry}
result, err := queryapi.GenerateDBWithOptions(query, User{}, options)
```

### Relationship Types Configuration

#### 1. One-to-One Relationships

```go
joinRegistry.Register("Profile", queryapi.JoinConfig{
    Type:       queryapi.OneToOne,
    Table:      "Profile",       // Table name
    LocalKey:   "ID",            // User.ID
    ForeignKey: "UserID",        // Profile.UserID
})
```

**Generated SQL:**
```sql
SELECT * FROM User 
LEFT JOIN Profile ON User.ID = Profile.UserID 
WHERE Profile.Bio LIKE ?
```

#### 2. One-to-Many Relationships

```go
joinRegistry.Register("Order", queryapi.JoinConfig{
    Type:       queryapi.OneToMany,
    Table:      "Order",         // Table name
    LocalKey:   "ID",            // User.ID
    ForeignKey: "UserID",        // Order.UserID
})
```

#### 3. Many-to-Many Relationships

```go
joinRegistry.Register("Corporate", queryapi.JoinConfig{
    Type:            queryapi.ManyToMany,
    Table:           "Corporate",         // Table name
    PivotTable:      "UserCorporate",     // Junction table
    PivotLocalKey:   "UserID",           // UserCorporate.UserID
    PivotForeignKey: "CorporateID",      // UserCorporate.CorporateID
})
```

**Generated SQL:**
```sql
SELECT * FROM User 
LEFT JOIN UserCorporate ON User.ID = UserCorporate.UserID 
LEFT JOIN Corporate ON UserCorporate.CorporateID = Corporate.ID 
WHERE Corporate.Name LIKE ?
```

#### 4. Polymorphic Relationships

```go
joinRegistry.Register("Comment", queryapi.JoinConfig{
    Type:             queryapi.Polymorphic,
    Table:            "Comment",          // Table name
    PolymorphicID:    "CommentableID",    // Comment.CommentableID
    PolymorphicType:  "CommentableType",  // Comment.CommentableType
    PolymorphicValue: "User",
})
```

---

## 🗄️ Database Migrations

### Replace GORM AutoMigrate with Goose

**Before (GORM):**
```go
db.AutoMigrate(&User{}, &Profile{}, &Order{})
```

**After (Goose):**

#### Setup Goose

```bash
# Install Goose
go install github.com/pressly/goose/v3/cmd/goose@latest

# Set environment variables
export GOOSE_DRIVER=mysql
export GOOSE_DBSTRING="user:password@tcp(localhost:3306)/database?parseTime=true"
```

#### Create and Run Migrations

```bash
# Create SQL migration
goose -dir ./migrations create create_user_table sql

# Run migrations
goose -dir ./migrations up

# Check status
goose -dir ./migrations status
```

#### Example Migration

```sql
-- +goose Up
CREATE TABLE User (
    ID INT AUTO_INCREMENT PRIMARY KEY,
    Name VARCHAR(255) NOT NULL,
    Email VARCHAR(255) UNIQUE NOT NULL,
    CityID INT,
    CreatedAt TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UpdatedAt TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);

CREATE TABLE UserCorporate (
    UserID INT,
    CorporateID INT,
    PRIMARY KEY (UserID, CorporateID)
);

-- +goose Down
DROP TABLE UserCorporate;
DROP TABLE User;
```

---

## 🔧 Advanced Migration Topics

### Custom Transactions

**Before (GORM):**
```go
db.Transaction(func(tx *gorm.DB) error {
    if err := tx.Create(&user).Error; err != nil {
        return err
    }
    if err := tx.Create(&profile).Error; err != nil {
        return err
    }
    return nil
})
```

**After (sqlx):**
```go
tx, err := database.Beginx()
if err != nil {
    return err
}
defer tx.Rollback()

err = mysql.Create(tx, &user)
if err != nil {
    return err
}

err = mysql.Create(tx, &profile)
if err != nil {
    return err
}

return tx.Commit()
```

### Error Handling

**Before (GORM):**
```go
if errors.Is(err, gorm.ErrRecordNotFound) {
    // Handle not found
}
```

**After (sqlx):**
```go
import "database/sql"

if errors.Is(err, sql.ErrNoRows) {
    // Handle not found
}
```

### Custom SQL Queries

**Before (GORM):**
```go
var results []Result
db.Raw("SELECT * FROM users WHERE age > ?", 25).Scan(&results)
```

**After (sqlx):**
```go
var results []Result
err := database.Select(&results, "SELECT * FROM User WHERE Age > ?", 25)
```

---

## 🎯 Migration Examples

### Basic CRUD Migration

**Before (GORM):**
```go
// Create
db.Create(&user)

// Read
db.First(&user, 1)

// Update
db.Save(&user)

// Delete
db.Delete(&user)

// Query with relationships
db.Preload("City").Preload("Corporates").Find(&users)

// Filter by relationship
db.Joins("City").Where("cities.name = ?", "Istanbul").Find(&users)
```

**After (sqlx):**
```go
// Create
mysql.Create(db, &user)

// Read
mysql.Read(db, &user, 1)

// Update
mysql.Update(db, &user)

// Delete
mysql.Delete(db, &user)

// Query with relationships 
joinRegistry := queryapi.NewJoinRegistry()
joinRegistry.Register("City", queryapi.JoinConfig{
    Type:       queryapi.OneToOne,
    Table:      "City",
    LocalKey:   "CityID",
    ForeignKey: "ID",
})
joinRegistry.Register("Corporate", queryapi.JoinConfig{
    Type:            queryapi.ManyToMany,
    Table:           "Corporate",
    PivotTable:      "UserCorporate",
    PivotLocalKey:   "UserID",
    PivotForeignKey: "CorporateID",
})

// Filter by relationship
mysql.List(db, &users, &qapi.Query{
    Filter: []qapi.Filter{
        {Name: "City.Name", Operation: qapi.EQ, Value: "Istanbul"},
    },
}, &queryapi.QueryOptions{JoinRegistry: joinRegistry})
```

---

## 🚨 Breaking Changes & Migration Notes

### 1. **No More AutoMigrate**
- **Before:** `db.AutoMigrate(&User{})`
- **After:** Use goose (or golang-migrate)

### 2. **No More Preloads**
- **Before:** `db.Preload("Profile").Find(&users)`
- **After:** Set up join registry and use `GenerateDBWithOptions`

### 3. **No More GORM Hooks**
- **Before:** `func (u *User) BeforeCreate(tx *gorm.DB) error`
- **After:** Implement in service layer

### 4. **Transaction Syntax**
- **Before:** `db.Transaction(func(tx *gorm.DB) error { ... })`
- **After:** `tx, err := db.Beginx(); defer tx.Rollback(); ... tx.Commit()`

### 5. **Error Handling**
- **Before:** `if errors.Is(err, gorm.ErrRecordNotFound)`
- **After:** `if errors.Is(err, sql.ErrNoRows)`

### 6. **Raw SQL**
- **Before:** `db.Raw(sql).Scan(&results)`
- **After:** `db.Select(&results, sql, args...)`

### 7. **Model Tags**
- **Before:** `gorm:"primary_key"`, `gorm:"foreignKey:UserID"`
- **After:** `db:"ID"` (relationships handled in join registry)

---

## ✅ Migration Checklist

After migration, verify these work:

- [ ] Database connection established
- [ ] All CRUD operations work
- [ ] Query API filters work
- [ ] Sorting and pagination work
- [ ] JSON fields work correctly
- [ ] Relationship queries work with join registry
- [ ] Transactions work
- [ ] Migrations run successfully
- [ ] Tests pass
- [ ] Performance is acceptable
- [ ] Error handling updated
- [ ] Custom SQL queries migrated

---

## 🎉 Migration Benefits

The migration to sqlx provides:

- ✅ **Same familiar API** - existing code works with minimal changes
- ✅ **Better performance** - especially with optimization interfaces
- ✅ **More control** - explicit relationship configuration
- ✅ **Cleaner dependencies** - no heavy ORM overhead
- ✅ **Future-proof** - easier to optimize and extend
- ✅ **Explicit relationships** - better understanding of database queries
- ✅ **Custom SQL support** - easier to write complex queries when needed

The Query API system remains fully functional, and the 5 core CRUD methods work exactly as before. The main difference is that relationships now require explicit configuration, giving you more control over your database queries.

**Migration is complete** - the project is now fully GORM-free! 🎉

---

## 🔗 Related Documentation

- **[README.md](./README.md)** - Quick start and API overview
- **[EXAMPLES.md](./EXAMPLES.md)** - Comprehensive code examples and best practices