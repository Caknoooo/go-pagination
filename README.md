# Go Pagination

A powerful, flexible, and easy-to-use pagination library for Go with GORM integration. This library provides multiple approaches to implement pagination in your Go applications with built-in support for searching, sorting, filtering, and multiple database dialects.

## Features

- 🚀 **Generic Support**: Full support for Go generics
- 🔍 **Advanced Search**: Built-in search functionality across multiple fields
- 🗂️ **Flexible Filtering**: Dynamic filtering with multiple operators
- 🔗 **Relationship Support**: Easy preloading of related data
- 🛢️ **Multi-Database**: Support for MySQL, PostgreSQL, SQLite, and SQL Server
- 🛡️ **SQL Injection Protection**: Built-in validation to prevent SQL injection
- ⚡ **Performance Optimized**: Efficient count and data queries
- 🧪 **Well Tested**: Comprehensive test coverage
- 📚 **Multiple APIs**: Simple helpers to advanced builders

## Installation

```bash
go get github.com/Caknoooo/go-pagination
```

## Quick Start

### Basic Usage with Gin

```go
package main

import (
    "github.com/Caknoooo/go-pagination"
    "github.com/gin-gonic/gin"
    "gorm.io/gorm"
)

type User struct {
    ID    uint   `json:"id"`
    Name  string `json:"name"`
    Email string `json:"email"`
}

func GetUsers(db *gorm.DB) gin.HandlerFunc {
    return func(c *gin.Context) {
        response := pagination.PaginatedAPIResponse[User](
            db, c, "users", 
            []string{"name", "email"}, // search fields
            "Users retrieved successfully",
        )
        c.JSON(response.Code, response)
    }
}
```

### URL Query Parameters

The library automatically handles these query parameters:

- `page`: Page number (default: 1)
- `per_page`: Items per page (default: 10, max: 100)
- `search`: Search term to filter results
- `sort`: Field to sort by
- `order`: Sort order (`asc` or `desc`, default: `asc`)
- `includes`: Comma-separated list of relations to preload

Example request:
```
GET /users?page=2&per_page=20&search=john&sort=created_at&order=desc&includes=Posts,Profile
```

## API Reference

### 1. Simple Helper Functions

#### PaginatedAPIResponse
The easiest way to implement pagination:

```go
func GetUsers(c *gin.Context) {
    response := pagination.PaginatedAPIResponse[User](
        db, c, "users", 
        []string{"name", "email"}, // search fields
        "Users retrieved successfully",
    )
    c.JSON(response.Code, response)
}
```

#### PaginateModel
For more control over the response:

```go
func GetUsers(c *gin.Context) {
    users, paginationInfo, err := pagination.PaginateModel[User](
        db, c, "users", []string{"name", "email"},
    )
    if err != nil {
        c.JSON(500, gin.H{"error": err.Error()})
        return
    }
    
    c.JSON(200, gin.H{
        "data": users,
        "pagination": paginationInfo,
    })
}
```

### 2. Advanced Query Builder

#### SimpleQueryBuilder
For custom filtering and more control:

```go
func GetActiveUsers(c *gin.Context) {
    paginationRequest := pagination.BindPagination(c)
    
    builder := pagination.NewSimpleQueryBuilder("users").
        WithSearchFields("name", "email").
        WithDefaultSort("created_at desc").
        WithDialect(pagination.MySQL).
        WithFilters(func(query *gorm.DB) *gorm.DB {
            return query.Where("active = ?", true)
        })
    
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
}
```

#### ChainableQueryBuilder
For complex queries with joins:

```go
func GetPostsWithUsers(c *gin.Context) {
    paginationRequest := pagination.BindPagination(c)
    
    builder := pagination.NewChainableQueryBuilder("posts").
        Join("LEFT JOIN users ON posts.user_id = users.id").
        Select("posts.*", "users.name as user_name").
        WithSearchFields("posts.title", "posts.body", "users.name").
        WithDefaultSort("posts.created_at desc")
    
    posts, total, err := pagination.PaginatedQuery[Post](
        db, builder, paginationRequest, []string{"User"},
    )
    
    // Handle response...
}
```

### 3. Dynamic Filtering

#### DynamicFilter
For advanced filtering based on struct fields:

```go
type UserFilter struct {
    pagination.DynamicFilter
    MinAge int    `json:"min_age" form:"min_age"`
    MaxAge int    `json:"max_age" form:"max_age"`
    Status string `json:"status" form:"status"`
}

func (u *UserFilter) ApplyFilters(query *gorm.DB) *gorm.DB {
    // Apply base dynamic filters
    query = u.DynamicFilter.ApplyFilters(query)
    
    // Apply custom filters
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

func GetUsersWithFilter(c *gin.Context) {
    filter := &UserFilter{
        DynamicFilter: pagination.DynamicFilter{
            TableName:    "users",
            Model:        User{},
            SearchFields: []string{"name", "email"},
            DefaultSort:  "id desc",
        },
    }
    
    // Bind all parameters
    filter.BindPagination(c)
    c.ShouldBindQuery(filter)
    
    users, total, err := pagination.PaginatedQuery[User](
        db, filter, filter.GetPagination(), filter.GetIncludes(),
    )
    
    // Handle response...
}
```

### 4. Database Dialect Support

The library supports multiple database dialects:

```go
builder := pagination.NewSimpleQueryBuilder("users").
    WithDialect(pagination.PostgreSQL) // Uses ILIKE for case-insensitive search

// Available dialects:
// pagination.MySQL      - Uses LIKE
// pagination.PostgreSQL - Uses ILIKE  
// pagination.SQLite     - Uses LIKE
// pagination.SQLServer  - Uses LIKE
```

### 5. Advanced Options

#### PaginatedQueryWithOptions
For more control over query execution:

```go
users, total, err := pagination.PaginatedQueryWithOptions[User](
    db, builder, paginationRequest, []string{},
    pagination.PaginatedQueryOptions{
        Dialect:          pagination.PostgreSQL,
        EnableSoftDelete: true, // Automatically filter soft-deleted records
        CustomCountQuery: "SELECT COUNT(*) FROM users WHERE active = true",
    },
)
```

## Filter Operators

When using DynamicFilter, you can use these operators:

- `=`, `EQ`, `EQUALS`: Equal to
- `!=`, `NE`, `NOT_EQUALS`: Not equal to  
- `>`, `GT`, `GREATER_THAN`: Greater than
- `>=`, `GTE`, `GREATER_THAN_EQUALS`: Greater than or equal
- `<`, `LT`, `LESS_THAN`: Less than
- `<=`, `LTE`, `LESS_THAN_EQUALS`: Less than or equal
- `LIKE`, `CONTAINS`: Contains (case-sensitive)
- `ILIKE`, `ICONTAINS`: Contains (case-insensitive)
- `IN`: In list
- `NOT_IN`: Not in list
- `IS_NULL`: Is null
- `IS_NOT_NULL`: Is not null

## Response Format

All paginated responses follow this structure:

```json
{
  "code": 200,
  "status": "success",
  "message": "Data retrieved successfully",
  "data": [...],
  "pagination": {
    "page": 1,
    "per_page": 10,
    "max_page": 5,
    "total": 50
  }
}
```

## Security

The library includes built-in protection against SQL injection:

- Field names are validated using regex patterns
- Sort fields are sanitized
- Include fields are validated
- Filter operators are whitelisted

## Testing

Run the tests:

```bash
go test ./...
```

The library includes comprehensive tests covering:
- Pagination logic
- Query building
- SQL injection prevention
- Database dialect support
- Error handling

## Performance Tips

1. **Use appropriate indexes**: Make sure your database has indexes on fields used for sorting and filtering
2. **Limit per_page**: The library limits per_page to 100 by default
3. **Use SelectFields**: When using ChainableQueryBuilder, select only needed fields
4. **Optimize count queries**: For large datasets, consider using CustomCountQuery with optimized counting logic

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Examples

Check the [examples](examples/) directory for complete working examples including:

- Basic pagination with Gin
- Advanced filtering
- Multiple database support
- Custom query builders
- Error handling

## Changelog

### v1.0.0
- Initial release
- Generic type support
- Multiple database dialects
- Advanced filtering
- SQL injection protection
- Comprehensive test coverage
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