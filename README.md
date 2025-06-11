# sincap-common

common libs, utils

- [chi](https://github.com/go-chi/chi) for routing
- [structs](https://github.com/fatih/structs) for reflection
- [zap](https://github.com/uber-go/zap) for logging
- [melody](https://github.com/olahol/melody) for websockets
- [sqlx](https://github.com/jmoiron/sqlx) for database operations
- [mysql driver](https://github.com/go-sql-driver/mysql) for MySQL connectivity
- [goose](https://github.com/pressly/goose) for database migrations
- [testify](https://github.com/stretchr/testify) for asserting

# Test

```bash
go get -u github.com/gophertown/looper
looper
```


## 🚀 Key Features

- **High-Performance Database Layer**: Migrated from GORM to sqlx for performance improvements
- **Query API**: Advanced filtering, sorting, pagination, and full-text search
- **Interface-Based Optimizations**: Optional interfaces for maximum performance
- **Relationship Support**: Manual joins and relationship queries
- **Migration Ready**: Goose integration for database migrations


## Query API

Multi level searches only works with SingularTableNames for PolymorphicModel and for equals

### Field selection
Give the API consumer the ability to choose returned fields. This will also reduce the network traffic and speed up the usage of the API.

```
GET /cars?_fields=manufacturer,model,id,color
```
### Preload selection
Give the API consumer the ability to choose preloaded (eager) relations. This will also reduce the network traffic and speed up the usage of the API.

```
GET /cars?_preloads=manufacturer,model,id,color
```

### Paging

```
GET /cars?_offset=10&_limit=5
```
* Add  _offset and _limit (an X-Total-Count header is included in the response).

* To send the total entries back to the user use the custom HTTP header: X-Total-Count.

* Content-Range offset – limit / count.

	* offset: Index of the first element returned by the request.

	* limit: Index of the last element returned by the request.

	* count: Total number of elements in the collection.

* Accept-Range resource max.



### Sorting

* Allow ascending and descending sorting over multiple fields.
* Use sort with underscore as `_sort`.
* In code, descending describe as ` - `, ascending describe as ` + `.

```GET /cars?_sort=-manufactorer,+model```

### Operators
* Add `_filter` query parameter and continue with field names,operations and values separated by `,`.
* Pattern `_filter=<fieldname><operation><value>`.
* Supported operations.
	* `=` equal
	* `!=` not equal
	* `<` less
	* `<=` less or equals
	* `>` greater
	* `>=` greater or equals
	* `~=` like
	* `|=` in (values must be separated with `|`
	* `*=` in alternative (values must be separated with `*`
* NULL/mull/nil id reserved word. For ex. Name=NULL or Name!=NULL becomes IS NULL or IS NOT NULL
	
```GET http://127.0.0.1:8080/app/users?_filter=name=seray,active=true```

### Full-text search

* Add `_q`.

```GET /cars?_q=nissan```


## Database Configuration

```go
import (
    "github.com/filllabs/sincap-common/db"
    "github.com/filllabs/sincap-common/db/mysql"
    "github.com/filllabs/sincap-common/db/queryapi"
)

// Configure database
configs := []db.Config{
    {
        Name: "default",
        Args: []string{"user:password@tcp(localhost:3306)/database?parseTime=true"},
    },
}
db.Configure(configs)
database := db.DB()
```

### Basic CRUD Operations

```go
// Create
user := &User{Name: "John", Email: "john@example.com"}
err := mysql.Create(database, user)

// Read
var user User
err := mysql.Read(database, &user, 1)

// Update
user.Name = "Jane"
err := mysql.Update(database, &user)

// Delete
err := mysql.Delete(database, &user)

// List with query
var users []User
count, err := mysql.List(database, &users, nil)
```

### Performance Optimization

Implement these interfaces on your models for maximum performance:

```go
type User struct {
    ID    uint   `db:"ID"`
    Name  string `db:"Name"`
    Email string `db:"Email"`
}

// TableNamer - eliminates reflection for table name
func (User) TableName() string { return "User" }

// IDGetter - eliminates reflection for ID access
func (u User) GetID() interface{} { return u.ID }

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
    }
}
```

### Programmatic Usage

```go
import "github.com/filllabs/sincap-common/middlewares/qapi"

query := &qapi.Query{
    Limit:  10,
    Offset: 0,
    Sort:   []string{"Name ASC", "CreatedAt DESC"},
    Filter: []qapi.Filter{
        {Name: "Age", Operation: qapi.GT, Value: "25"},
        {Name: "Status", Operation: qapi.EQ, Value: "active"},
        {Name: "Name", Operation: qapi.LK, Value: "John"},
    },
    Q: []string{"developer"}, // Full-text search
}

var users []User
count, err := mysql.List(database, &users, query)
```

### Relationship Queries

For queries involving relationships, use the join registry system:

```go
import "github.com/filllabs/sincap-common/db/queryapi"

// Set up join registry
joinRegistry := queryapi.NewJoinRegistry()
joinRegistry.Register("Profile", queryapi.JoinConfig{
    Type:       queryapi.OneToOne,
    Table:      "Profile",
    LocalKey:   "ID",
    ForeignKey: "UserID",
})

// Query with relationship filters
query := &qapi.Query{
    Filter: []qapi.Filter{
        {Name: "Profile.Bio", Operation: qapi.LK, Value: "developer"},
    },
}

options := &queryapi.QueryOptions{JoinRegistry: joinRegistry}
result, err := queryapi.GenerateDBWithOptions(query, User{}, options)
```

## 🗄️ Database Migrations

This library uses Goose for database migrations:

```bash
# Install Goose
go install github.com/pressly/goose/v3/cmd/goose@latest

# Create migration
goose -dir ./migrations create create_users_table sql

# Run migrations
goose -dir ./migrations up

# Check status
goose -dir ./migrations status
```


## 📚 Documentation

For detailed information, see:

- **[`EXAMPLES.md`](./EXAMPLES.md)** - Comprehensive examples and practical usage patterns
- **[`MIGRATION.md`](./MIGRATION.md)** - Complete GORM to sqlx migration guide with detailed relationship handling

## 🎯 Migration from GORM

If you're migrating from GORM, the API remains largely the same:

**Before (GORM):**
```go
db.Create(&user)
db.Find(&users)
db.Where("age > ?", 25).Find(&users)
```

**After (sqlx):**
```go
mysql.Create(db, &user)
mysql.List(db, &users, nil)
mysql.List(db, &users, &qapi.Query{
    Filter: []qapi.Filter{
        {Name: "age", Operation: qapi.GT, Value: "25"},
    },
})
```
