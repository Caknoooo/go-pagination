# go-pagination

```bash
go get "github.com/Caknoooo/go-pagination"
```

# testing with gin
docker-compose.yml
```docker
services:
  db:
    image: mysql:latest
    container_name: db
    ports:
      - "3306:3306"
    environment:
      MYSQL_ROOT_PASSWORD: root
      MYSQL_DATABASE: test22
    volumes:
      - db_data:/var/lib/mysql

volumes:
  db_data:
```

main.go
```go
package main

import (
	"fmt"
	"log"
	"net/http"
	"tested/pagination"

	"github.com/gin-gonic/gin"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type User struct {
	ID    uint   `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

func main() {
	// Connect to database
	dsn := "root:root@tcp(localhost:3306)/test22?charset=utf8mb4&parseTime=True&loc=Local"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database: ", err)
	}

	// Migrate User model to database
	if err := db.AutoMigrate(&User{}); err != nil {
		log.Fatal("Failed to migrate database: ", err)
	}

	// Seed initial users data
	seedUsers(db)

	// Setup Gin router
	r := gin.Default()

	// Define GET route for paginated users
	r.GET("/users", func(c *gin.Context) {
		// Initialize pagination
		p, err := pagination.New(db, c)
		if err != nil {
			log.Fatal("Error creating pagination: ", err)
		}

		var users []User
		// Query paginated data
		if err := p.Query().Find(&users).Error; err != nil {
			log.Fatal("Error querying database: ", err)
		}

		// Count total items for pagination metadata
		if err := p.Count(&User{}); err != nil {
			log.Fatal("Error counting items: ", err)
		}

		// Generate and return paginated response
		response := p.GenerateResponse(c)
		c.JSON(http.StatusOK, gin.H{
			"data":  users,
			"meta":  response.Meta,
			"links": response.Links,
		})
	})

	// Run the server
	if err := r.Run(":8081"); err != nil {
		log.Fatal("Failed to start server: ", err)
	}
}

// Seed dummy users data
func seedUsers(db *gorm.DB) {
	for i := 1; i <= 50; i++ {
		user := User{
			Name:  fmt.Sprintf("User %d", i),
			Email: fmt.Sprintf("user%d@example.com", i),
		}
		db.Create(&user)
	}
}
```