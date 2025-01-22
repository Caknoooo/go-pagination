# go-pagination

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
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"github.com/Caknoooo/go-pagination"
)

type user struct {
	ID   int
	Name string
	Age  int
}

func seedUser(db *gorm.DB) {
	for i := 1; i <= 50; i++ {
		name := "user-" + strconv.Itoa(i)
		age := i
		db.Create(&user{Name: name, Age: age})
	}
}

func main() {
	dsn := "root:root@tcp(localhost:3306)/dbname?charset=utf8mb4&parseTime=True&loc=Local"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	db.AutoMigrate(&user{})
	seedUser(db)

	r := gin.Default()

	r.GET("/users", func(c *gin.Context) {
		var req pagination.PaginationRequest

		err := c.ShouldBind(&req)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
			return
		}

		var totalUsers int64
		db.Model(&user{}).Select("id").Count(&totalUsers)

		// Generate pagination links and meta
		resp, err := pagination.GeneratePaginationLinks(c, req, int(totalUsers))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate pagination links"})
			return
		}

		// Fetch the paginated users
		var users []user
		offset := (req.Number - 1) * req.Size
		db.Offset(offset).Limit(req.Size).Find(&users)

		c.JSON(http.StatusOK, gin.H{
			"data":  users,
			"meta":  resp.Meta,
			"links": resp.Links,
		})
	})

	r.Run(":8080")
}
```