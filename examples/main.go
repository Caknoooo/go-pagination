package main

import (
	"log"

	"github.com/Caknoooo/go-pagination"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// User model example
type User struct {
	ID    uint   `json:"id" gorm:"primaryKey"`
	Name  string `json:"name" gorm:"column:name"`
	Email string `json:"email" gorm:"column:email"`
	Age   int    `json:"age" gorm:"column:age"`
	Posts []Post `json:"posts,omitempty" gorm:"foreignKey:UserID"`
}

// Post model example
type Post struct {
	ID     uint   `json:"id" gorm:"primaryKey"`
	Title  string `json:"title" gorm:"column:title"`
	Body   string `json:"body" gorm:"column:body"`
	UserID uint   `json:"user_id" gorm:"column:user_id"`
	User   *User  `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

// UserFilter for advanced filtering
type UserFilter struct {
	pagination.DynamicFilter
	MinAge int    `json:"min_age" form:"min_age"`
	MaxAge int    `json:"max_age" form:"max_age"`
	Status string `json:"status" form:"status"`
}

func (u *UserFilter) ApplyFilters(query *gorm.DB) *gorm.DB {
	// Apply base dynamic filters first
	query = u.DynamicFilter.ApplyFilters(query)

	// Apply custom age filters
	if u.MinAge > 0 {
		query = query.Where("age >= ?", u.MinAge)
	}
	if u.MaxAge > 0 {
		query = query.Where("age <= ?", u.MaxAge)
	}
	if u.Status != "" {
		query = query.Where("status = ?", u.Status)
	}

	return query
}

func main() {
	// Database connection
	dsn := "root:root@tcp(localhost:3306)/test?charset=utf8mb4&parseTime=True&loc=Local"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Auto migrate
	db.AutoMigrate(&User{}, &Post{})

	// Seed data if needed
	seedData(db)

	r := gin.Default()

	// Example 1: Simple pagination dengan helper function
	r.GET("/users/simple", func(c *gin.Context) {
		response := pagination.PaginatedAPIResponse[User](
			db, c, "users",
			[]string{"name", "email"}, // search fields
			"Users retrieved successfully",
		)
		c.JSON(response.Code, response)
	})

	// Example 2: Pagination dengan includes/preload
	r.GET("/users/with-posts", func(c *gin.Context) {
		response := pagination.PaginatedAPIResponseWithIncludes[User](
			db, c, "users",
			[]string{"name", "email"},
			[]string{"Posts"}, // preload Posts relationship
			"Users with posts retrieved successfully",
		)
		c.JSON(response.Code, response)
	})

	// Example 3: Manual pagination dengan custom builder
	r.GET("/users/manual", func(c *gin.Context) {
		paginationRequest := pagination.BindPagination(c)

		builder := pagination.NewSimpleQueryBuilder("users").
			WithSearchFields("name", "email").
			WithDefaultSort("created_at desc").
			WithDialect(pagination.MySQL)

		users, total, err := pagination.PaginatedQuery[User](
			db, builder, paginationRequest, []string{},
		)

		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}

		paginationResponse := pagination.CalculatePagination(paginationRequest, total)
		response := pagination.NewPaginatedResponse(200, "Success", users, paginationResponse)

		c.JSON(200, response)
	})

	// Example 4: Advanced filtering dengan DynamicFilter
	r.GET("/users/advanced", func(c *gin.Context) {
		filter := &UserFilter{
			DynamicFilter: pagination.DynamicFilter{
				TableName:    "users",
				Model:        User{},
				SearchFields: []string{"name", "email"},
				DefaultSort:  "id desc",
			},
		}

		// Bind pagination dan filter parameters
		filter.BindPagination(c)

		// Bind custom filter parameters
		if err := c.ShouldBindQuery(filter); err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}

		users, total, err := pagination.PaginatedQuery[User](
			db, filter, filter.GetPagination(), filter.GetIncludes(),
		)

		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}

		paginationResponse := pagination.CalculatePagination(filter.GetPagination(), total)
		response := pagination.NewPaginatedResponse(200, "Success", users, paginationResponse)

		c.JSON(200, response)
	})

	// Example 5: Pagination dengan custom filter function
	r.GET("/users/custom-filter", func(c *gin.Context) {
		// Custom filter untuk user yang aktif
		filterFunc := func(query *gorm.DB) *gorm.DB {
			return query.Where("active = ?", true)
		}

		users, paginationResponse, err := pagination.PaginateWithFilter[User](
			db, c, "users",
			[]string{"name", "email"},
			filterFunc,
		)

		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}

		response := pagination.NewPaginatedResponse(200, "Active users retrieved", users, paginationResponse)
		c.JSON(200, response)
	})

	// Example 6: Chainable Query Builder untuk query kompleks
	r.GET("/posts/with-users", func(c *gin.Context) {
		paginationRequest := pagination.BindPagination(c)

		builder := pagination.NewChainableQueryBuilder("posts").
			Join("LEFT JOIN users ON posts.user_id = users.id").
			Select("posts.*", "users.name as user_name").
			WithSearchFields("posts.title", "posts.body", "users.name").
			WithDefaultSort("posts.created_at desc")

		posts, total, err := pagination.PaginatedQuery[Post](
			db, builder, paginationRequest, []string{"User"},
		)

		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}

		paginationResponse := pagination.CalculatePagination(paginationRequest, total)
		response := pagination.NewPaginatedResponse(200, "Posts with users retrieved", posts, paginationResponse)

		c.JSON(200, response)
	})

	// Example 7: Pagination dengan PostgreSQL dialect
	r.GET("/users/postgresql", func(c *gin.Context) {
		paginationRequest := pagination.BindPagination(c)

		builder := pagination.NewSimpleQueryBuilder("users").
			WithSearchFields("name", "email").
			WithDialect(pagination.PostgreSQL). // Menggunakan ILIKE untuk PostgreSQL
			WithDefaultSort("id asc")

		users, total, err := pagination.PaginatedQueryWithOptions[User](
			db, builder, paginationRequest, []string{},
			pagination.PaginatedQueryOptions{
				Dialect:          pagination.PostgreSQL,
				EnableSoftDelete: true, // Enable soft delete handling
			},
		)

		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}

		paginationResponse := pagination.CalculatePagination(paginationRequest, total)
		response := pagination.NewPaginatedResponse(200, "Success", users, paginationResponse)

		c.JSON(200, response)
	})

	log.Println("Server starting on :8080")
	r.Run(":8080")
}

func seedData(db *gorm.DB) {
	// Check if data already exists
	var count int64
	db.Model(&User{}).Count(&count)
	if count > 0 {
		return
	}

	// Seed users
	users := []User{
		{Name: "John Doe", Email: "john@example.com", Age: 25},
		{Name: "Jane Smith", Email: "jane@example.com", Age: 30},
		{Name: "Bob Johnson", Email: "bob@example.com", Age: 35},
		{Name: "Alice Brown", Email: "alice@example.com", Age: 28},
		{Name: "Charlie Wilson", Email: "charlie@example.com", Age: 32},
	}

	for _, user := range users {
		db.Create(&user)
	}

	// Seed posts
	posts := []Post{
		{Title: "First Post", Body: "This is the first post", UserID: 1},
		{Title: "Second Post", Body: "This is the second post", UserID: 1},
		{Title: "Jane's Post", Body: "This is Jane's post", UserID: 2},
		{Title: "Bob's Article", Body: "This is Bob's article", UserID: 3},
		{Title: "Alice's Story", Body: "This is Alice's story", UserID: 4},
	}

	for _, post := range posts {
		db.Create(&post)
	}
}
