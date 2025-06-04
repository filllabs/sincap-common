package main

import (
	"fmt"
	"log"

	"github.com/filllabs/sincap-common/db"
	"github.com/filllabs/sincap-common/db/mysql"
	"github.com/filllabs/sincap-common/db/queryapi"
	"github.com/filllabs/sincap-common/middlewares/qapi"
)

// Example model with optimized interfaces (reduces reflection usage)
type User struct {
	ID    uint   `json:"id" db:"ID"`
	Name  string `json:"name" db:"Name"`
	Email string `json:"email" db:"Email"`
	Age   int    `json:"age" db:"Age"`
}

// TableName implements mysql.TableNamer interface (reduces reflection)
func (User) TableName() string {
	return "Users"
}

// GetID implements mysql.IDGetter interface (reduces reflection)
func (u User) GetID() interface{} {
	return u.ID
}

// SetID implements mysql.IDSetter interface (reduces reflection)
func (u *User) SetID(id interface{}) error {
	if idVal, ok := id.(uint64); ok {
		u.ID = uint(idVal)
		return nil
	}
	return fmt.Errorf("invalid ID type")
}

// GetFieldMap implements mysql.FieldMapper interface (eliminates reflection for CRUD)
func (u User) GetFieldMap() map[string]interface{} {
	return map[string]interface{}{
		"ID":    u.ID,
		"Name":  u.Name,
		"Email": u.Email,
		"Age":   u.Age,
	}
}

// Example model without interfaces (uses reflection fallback)
type Product struct {
	ID    uint    `json:"id" db:"ID"`
	Name  string  `json:"name" db:"Name"`
	Price float64 `json:"price" db:"Price"`
}

// TableName method (will be called via reflection)
func (Product) TableName() string {
	return "Products"
}

func main() {
	// Example of how to use the new sqlx-based system with reduced reflection
	fmt.Println("=== Sincap-Common sqlx Migration Example (Optimized) ===")

	// 1. Database Configuration (you would normally do this in your app initialization)
	configs := []db.Config{
		{
			Name: "default",
			Args: []string{"user:password@tcp(localhost:3306)/database?parseTime=true"},
		},
	}

	// Configure database connections
	db.Configure(configs)

	// Get database connection
	database := db.DB()

	// 2. Example: Create a user (OPTIMIZED - no reflection)
	fmt.Println("\n--- Create Example (Optimized) ---")
	user := &User{
		Name:  "John Doe",
		Email: "john@example.com",
		Age:   30,
	}

	// This will use the FieldMapper interface - NO REFLECTION!
	err := mysql.Create(database, user)
	if err != nil {
		log.Printf("Create error: %v", err)
	} else {
		fmt.Printf("Created user with ID: %d (no reflection used!)\n", user.ID)
	}

	// 3. Example: Create using field map directly (NO REFLECTION)
	fmt.Println("\n--- Create with Field Map (Zero Reflection) ---")
	userFieldMap := map[string]interface{}{
		"Name":  "Jane Smith",
		"Email": "jane@example.com",
		"Age":   25,
	}

	var newUser User
	err = mysql.CreateWithFieldMap(database, "Users", userFieldMap, &newUser)
	if err != nil {
		log.Printf("CreateWithFieldMap error: %v", err)
	} else {
		fmt.Printf("Created user with field map, ID: %d (zero reflection!)\n", newUser.ID)
	}

	// 4. Example: Read a user (OPTIMIZED - minimal reflection)
	fmt.Println("\n--- Read Example (Optimized) ---")
	var readUser User
	// This will use the TableName() interface method - minimal reflection!
	err = mysql.Read(database, &readUser, 1)
	if err != nil {
		log.Printf("Read error: %v", err)
	} else {
		fmt.Printf("Read user: %+v (minimal reflection used)\n", readUser)
	}

	// 5. Example: Read by ID directly (NO REFLECTION)
	fmt.Println("\n--- Read by ID (Zero Reflection) ---")
	var directReadUser User
	err = mysql.ReadByID(database, &directReadUser, "Users", 1)
	if err != nil {
		log.Printf("ReadByID error: %v", err)
	} else {
		fmt.Printf("Read user directly: %+v (zero reflection!)\n", directReadUser)
	}

	// 6. Example: Update a user (OPTIMIZED - no reflection)
	fmt.Println("\n--- Update Example (Optimized) ---")
	readUser.Age = 31
	// This will use the FieldMapper and IDGetter interfaces - NO REFLECTION!
	err = mysql.Update(database, &readUser)
	if err != nil {
		log.Printf("Update error: %v", err)
	} else {
		fmt.Println("User updated successfully (no reflection used!)")
	}

	// 7. Example: Partial update with field map (NO REFLECTION)
	fmt.Println("\n--- Partial Update with Field Map (Zero Reflection) ---")
	err = mysql.UpdateWithFieldMap(database, "Users", 1, map[string]interface{}{
		"Email": "john.doe@example.com",
		"Age":   32,
	})
	if err != nil {
		log.Printf("UpdateWithFieldMap error: %v", err)
	} else {
		fmt.Println("User updated with field map (zero reflection!)")
	}

	// 8. Example: Delete a user (OPTIMIZED - no reflection)
	fmt.Println("\n--- Delete Example (Optimized) ---")
	// This will use the TableName() and GetID() interfaces - NO REFLECTION!
	err = mysql.Delete(database, &readUser)
	if err != nil {
		log.Printf("Delete error: %v", err)
	} else {
		fmt.Println("User deleted successfully (no reflection used!)")
	}

	// 9. Example: Delete by ID directly (NO REFLECTION)
	fmt.Println("\n--- Delete by ID (Zero Reflection) ---")
	err = mysql.DeleteByID(database, "Users", 2)
	if err != nil {
		log.Printf("DeleteByID error: %v", err)
	} else {
		fmt.Println("User deleted by ID (zero reflection!)")
	}

	// 10. Example: Comparison with reflection-based model
	fmt.Println("\n--- Comparison: Reflection-based Model ---")
	product := &Product{
		Name:  "Laptop",
		Price: 999.99,
	}

	// This will fall back to reflection since Product doesn't implement the interfaces
	err = mysql.Create(database, product)
	if err != nil {
		log.Printf("Create product error: %v", err)
	} else {
		fmt.Printf("Created product with ID: %d (reflection used as fallback)\n", product.ID)
	}

	// 11. Example: Query API with optimized table name resolution
	fmt.Println("\n--- Query API Example (Optimized) ---")
	var users []User
	query := &qapi.Query{
		Filter: []qapi.Filter{
			{Name: "Age", Value: "25", Operation: qapi.GT},
		},
		Sort:   []string{"Name ASC"},
		Limit:  10,
		Offset: 0,
	}

	count, err := mysql.List(database, &users, query)
	if err != nil {
		log.Printf("List error: %v", err)
	} else {
		fmt.Printf("Found %d users (table name resolved via interface)\n", count)
	}

	// 12. Example: Query API with joins (advanced)
	fmt.Println("\n--- Query API with Joins Example ---")

	// Set up join registry for relationship queries
	joinRegistry := queryapi.NewJoinRegistry()
	joinRegistry.Register("Profile", queryapi.JoinConfig{
		Type:       queryapi.OneToOne,
		Table:      "Profiles",
		LocalKey:   "ID",
		ForeignKey: "UserID",
	})

	// Complex query with relationships
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

	result, err := queryapi.GenerateSQLWithOptions(complexQuery, User{}, options)
	if err != nil {
		log.Printf("GenerateSQLWithOptions error: %v", err)
	} else {
		fmt.Printf("Generated optimized query: %s\n", result.Query)
		fmt.Printf("Query args: %v\n", result.Args)
	}

	fmt.Println("\n=== Performance Benefits ===")
	fmt.Println("✅ User model: Uses interfaces - NO reflection for CRUD operations")
	fmt.Println("✅ Direct functions: CreateWithFieldMap, ReadByID, etc. - ZERO reflection")
	fmt.Println("⚠️  Product model: Falls back to reflection (still works)")
	fmt.Println("🚀 Result: Significantly faster CRUD operations for optimized models")
}
