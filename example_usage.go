package main

import (
	"fmt"
	"log"

	"github.com/filllabs/sincap-common/db"
	"github.com/filllabs/sincap-common/db/mysql"
	"github.com/filllabs/sincap-common/db/queryapi"
	"github.com/filllabs/sincap-common/middlewares/qapi"
)

// Example model
type User struct {
	ID    uint   `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
	Age   int    `json:"age"`
}

// TableName returns the table name for the User model
func (User) TableName() string {
	return "users"
}

func main() {
	// Example of how to use the new sqlx-based system
	fmt.Println("=== Sincap-Common sqlx Migration Example ===")

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

	// 2. Example: Create a user
	fmt.Println("\n--- Create Example ---")
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

	// 3. Example: Read a user
	fmt.Println("\n--- Read Example ---")
	var readUser User
	err = mysql.Read(database, &readUser, 1)
	if err != nil {
		log.Printf("Read error: %v", err)
	} else {
		fmt.Printf("Read user: %+v\n", readUser)
	}

	// 4. Example: Update a user
	fmt.Println("\n--- Update Example ---")
	readUser.Age = 31
	err = mysql.Update(database, &readUser)
	if err != nil {
		log.Printf("Update error: %v", err)
	} else {
		fmt.Println("User updated successfully")
	}

	// 5. Example: Partial update
	fmt.Println("\n--- Partial Update Example ---")
	err = mysql.Update(database, &User{ID: 1}, map[string]any{
		"email": "john.doe@example.com",
	})
	if err != nil {
		log.Printf("Partial update error: %v", err)
	} else {
		fmt.Println("User email updated successfully")
	}

	// 6. Example: List with query API
	fmt.Println("\n--- List with Query API Example ---")
	query := &qapi.Query{
		Limit:  10,
		Offset: 0,
		Sort:   []string{"name ASC"},
		Filter: []qapi.Filter{
			{
				Name:      "age",
				Operation: qapi.GT,
				Value:     "25",
			},
		},
	}

	var users []User
	count, err := mysql.List(database, &users, query)
	if err != nil {
		log.Printf("List error: %v", err)
	} else {
		fmt.Printf("Found %d users: %+v\n", count, users)
	}

	// 7. Example: Using GenerateSQL directly for custom queries
	fmt.Println("\n--- Custom Query Example ---")
	queryResult, err := queryapi.GenerateSQL(query, &users)
	if err != nil {
		log.Printf("GenerateSQL error: %v", err)
	} else {
		fmt.Printf("Generated SQL: %s\n", queryResult.Query)
		fmt.Printf("Parameters: %+v\n", queryResult.Args)
		fmt.Printf("Count SQL: %s\n", queryResult.CountQuery)
	}

	// 8. Example: Delete a user
	fmt.Println("\n--- Delete Example ---")
	err = mysql.Delete(database, &User{ID: 1})
	if err != nil {
		log.Printf("Delete error: %v", err)
	} else {
		fmt.Println("User deleted successfully")
	}

	// 9. Example: Bulk delete
	fmt.Println("\n--- Bulk Delete Example ---")
	err = mysql.DeleteAll(database, &User{}, 2, 3, 4)
	if err != nil {
		log.Printf("Bulk delete error: %v", err)
	} else {
		fmt.Println("Users deleted successfully")
	}

	fmt.Println("\n=== Migration Complete! ===")
	fmt.Println("Key Changes:")
	fmt.Println("1. GORM replaced with sqlx")
	fmt.Println("2. Manual SQL query building")
	fmt.Println("3. No automatic relationships (must be handled manually)")
	fmt.Println("4. No soft deletes (must be implemented manually)")
	fmt.Println("5. Use golang-migrate or goose for schema migrations")
	fmt.Println("6. GenerateSQL() returns SQL strings instead of GORM objects")
}
