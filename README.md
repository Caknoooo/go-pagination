# Go Pagination 🚀

A **dynamic, flexible, and easy-to-use** pagination library for Go with GORM integration. This library provides various approaches to implement pagination in your Go applications with built-in support for searching, sorting, filtering, and multiple databases.

## ✨ Key Features

- 🚀 **Generic Support**: Full support for Go generics
- 🔍 **Smart Search**: Automatic search based on specified fields
- 🗂️ **Dynamic Filtering**: Dynamic filters with various operators
- 🔗 **Relationship Support**: Easy preloading of relations
- 🛢️ **Multi-Database**: Support for MySQL, PostgreSQL, SQLite, and SQL Server
- 🛡️ **SQL Injection Protection**: Built-in validation to prevent SQL injection
- ⚡ **Performance Optimized**: Efficient count and data queries
- 🧪 **Well Tested**: Comprehensive test coverage
- 📚 **Multiple Patterns**: From simple helpers to advanced builders

## 📦 Installation

```bash
go get github.com/Caknoooo/go-pagination
```

## 🚀 Quick Start

### 1. Simple API Response (Easiest Way!)

```go
package main

import (
    "github.com/Caknoooo/go-pagination"
    "github.com/gin-gonic/gin"
    "gorm.io/gorm"
)

type User struct {
    ID    uint   `json:"id" gorm:"primaryKey"`
    Name  string `json:"name"`
    Email string `json:"email"`
}

func GetUsers(db *gorm.DB) gin.HandlerFunc {
    return func(c *gin.Context) {
        // One line for pagination with search!
        response := pagination.PaginatedAPIResponse[User](
            db, c, "users", 
            []string{"name", "email"}, // searchable fields
            "Users retrieved successfully",
        )
        c.JSON(response.Code, response)
    }
}

func main() {
    r := gin.Default()
    r.GET("/users", GetUsers(db))
    r.Run(":8080")
}
```

**URL Examples:**
- `GET /users?page=1&limit=10` - Basic pagination
- `GET /users?search=john&page=1&limit=10` - Search in name & email
- `GET /users?sort=name,desc&page=1&limit=10` - Sort by name descending

### 2. Custom Filter Pattern with Include Support (More Flexible!)

```go
type UserFilter struct {
    pagination.BaseFilter
    ID       int    `json:"id" form:"id"`
    Name     string `json:"name" form:"name"`
    Email    string `json:"email" form:"email"`
    IsActive *bool  `json:"is_active" form:"is_active"`
}

// Custom filter implementation
func (f *UserFilter) ApplyFilters(query *gorm.DB) *gorm.DB {
    if f.ID > 0 {
        query = query.Where("id = ?", f.ID)
    }
    if f.Name != "" {
        query = query.Where("name LIKE ?", "%"+f.Name+"%")
    }
    if f.Email != "" {
        query = query.Where("email LIKE ?", "%"+f.Email+"%")
    }
    if f.IsActive != nil {
        query = query.Where("is_active = ?", *f.IsActive)
    }
    return query
}

// Define searchable fields (DYNAMIC!)
func (f *UserFilter) GetSearchFields() []string {
    return []string{"name", "email", "phone"}
}

func (f *UserFilter) GetTableName() string {
    return "users"
}

func (f *UserFilter) GetDefaultSort() string {
    return "id asc"
}

// IncludableQueryBuilder implementation for relationships
func (f *UserFilter) GetIncludes() []string {
    return f.Includes
}

func (f *UserFilter) GetPagination() pagination.PaginationRequest {
    return f.Pagination
}

func (f *UserFilter) Validate() {
    var validIncludes []string
    allowedIncludes := f.GetAllowedIncludes()
    for _, include := range f.Includes {
        if allowedIncludes[include] {
            validIncludes = append(validIncludes, include)
        }
    }
    f.Includes = validIncludes
}

func (f *UserFilter) GetAllowedIncludes() map[string]bool {
    return map[string]bool{
        "Profile": true,
        "Posts":   true,
        "Orders":  true,
    }
}

// Usage in handler with automatic include validation
func GetUsersWithFilter(db *gorm.DB) gin.HandlerFunc {
    return func(c *gin.Context) {
        var filter UserFilter
        if err := pagination.BindPagination(c, &filter); err != nil {
            c.JSON(400, gin.H{"error": err.Error()})
            return
        }

        // Use IncludableQueryBuilder for automatic validation
        users, total, err := pagination.PaginatedQueryWithIncludable[User](db, &filter)
        if err != nil {
            c.JSON(500, gin.H{"error": err.Error()})
            return
        }

        paginationResponse := pagination.CalculatePagination(filter.GetPagination(), total)
        response := pagination.NewPaginatedResponse(200, "Users retrieved successfully", users, paginationResponse)
        c.JSON(200, response)
    }
}
```

**URL Examples with Includes:**
- `GET /users?page=1&limit=10` - Basic pagination
- `GET /users?includes=Profile,Posts&page=1&limit=10` - With relationships
- `GET /users?search=john&includes=Profile&page=1&limit=10` - Search with relationships

### 3. Advanced Pattern with Complex Relationships

```go
type UserWithProfile struct {
    ID      uint    `json:"id"`
    Name    string  `json:"name"`
    Email   string  `json:"email"`
    Profile Profile `json:"profile" gorm:"foreignKey:UserID"`
    Posts   []Post  `json:"posts" gorm:"foreignKey:UserID"`
}

type Profile struct {
    ID     uint   `json:"id"`
    UserID uint   `json:"user_id"`
    Bio    string `json:"bio"`
    Avatar string `json:"avatar"`
}

type Post struct {
    ID     uint   `json:"id"`
    UserID uint   `json:"user_id"`
    Title  string `json:"title"`
    Content string `json:"content"`
}

type UserProfileFilter struct {
    pagination.BaseFilter
    Name     string `json:"name" form:"name"`
    Bio      string `json:"bio" form:"bio"`
    PostTitle string `json:"post_title" form:"post_title"`
}

func (f *UserProfileFilter) ApplyFilters(query *gorm.DB) *gorm.DB {
    if f.Name != "" {
        query = query.Where("users.name LIKE ?", "%"+f.Name+"%")
    }
    if f.Bio != "" {
        query = query.Joins("JOIN profiles ON profiles.user_id = users.id").
               Where("profiles.bio LIKE ?", "%"+f.Bio+"%")
    }
    if f.PostTitle != "" {
        query = query.Joins("JOIN posts ON posts.user_id = users.id").
               Where("posts.title LIKE ?", "%"+f.PostTitle+"%")
    }
    return query
}

func (f *UserProfileFilter) GetSearchFields() []string {
    return []string{"users.name", "users.email", "profiles.bio", "posts.title"}
}

func (f *UserProfileFilter) GetTableName() string {
    return "users"
}

func (f *UserProfileFilter) GetDefaultSort() string {
    return "users.id asc"
}

// IncludableQueryBuilder implementation with multiple relationships
func (f *UserProfileFilter) GetIncludes() []string {
    return f.Includes
}

func (f *UserProfileFilter) GetPagination() pagination.PaginationRequest {
    return f.Pagination
}

func (f *UserProfileFilter) Validate() {
    var validIncludes []string
    allowedIncludes := f.GetAllowedIncludes()
    for _, include := range f.Includes {
        if allowedIncludes[include] {
            validIncludes = append(validIncludes, include)
        }
    }
    f.Includes = validIncludes
}

func (f *UserProfileFilter) GetAllowedIncludes() map[string]bool {
    return map[string]bool{
        "Profile": true,
        "Posts":   true,
        // Nested relationships also supported
        "Profile.Address": true,
        "Posts.Comments":  true,
    }
}

// Handler with automatic relationship loading and validation
func GetUsersWithProfiles(db *gorm.DB) gin.HandlerFunc {
    return func(c *gin.Context) {
        filter := &UserProfileFilter{}
        filter.BindPagination(c)
        c.ShouldBindQuery(filter)

        // Automatically validates includes and loads relationships
        users, total, err := pagination.PaginatedQueryWithIncludable[UserWithProfile](db, filter)
        if err != nil {
            c.JSON(500, gin.H{"error": err.Error()})
            return
        }

        paginationResponse := pagination.CalculatePagination(filter.GetPagination(), total)
        response := pagination.NewPaginatedResponse(200, "Users with profiles retrieved successfully", users, paginationResponse)
        c.JSON(200, response)
    }
}
```

**Advanced URL Examples:**
- `GET /users/profiles?includes=Profile,Posts&page=1&limit=10` - Multiple relationships
- `GET /users/profiles?includes=Profile.Address,Posts.Comments&page=1&limit=10` - Nested relationships
- `GET /users/profiles?bio=developer&includes=Profile&search=john&page=1&limit=10` - Complex filtering with relationships

### ✨ Security Features for Includes

- **Include Validation**: Only includes listed in `GetAllowedIncludes()` will be processed
- **SQL Injection Protection**: All includes are validated with regex patterns
- **Type Safety**: Uses interfaces to ensure required methods are available
- **Nested Relationships**: Support for deep relationship loading with validation

### 🎯 Real-World Example: Athlete Management System

```go
// Complete example from the examples/ folder
type AthleteFilter struct {
    pagination.BaseFilter
    ID         int `json:"id" form:"id"`
    ProvinceID int `json:"province_id" form:"province_id"`
    SportID    int `json:"sport_id" form:"sport_id"`
    EventID    int `json:"event_id" form:"event_id"`
}

func (f *AthleteFilter) GetAllowedIncludes() map[string]bool {
    return map[string]bool{
        "Province":      true,  // Load province data
        "Sport":         true,  // Load sport data
        "PlayersEvents": true,  // Load events participation
    }
}

// API Usage Examples:
// GET /athletes?includes=Province,Sport&page=1&limit=10
// GET /athletes?includes=Province&province_id=1&search=john
// GET /athletes?includes=Sport,PlayersEvents&sport_id=2
```

## 🔍 Search Features

### Automatic Search with GetSearchFields

What makes this library **extremely dynamic** is the `GetSearchFields()` capability:

```go
func (f *ProductFilter) GetSearchFields() []string {
    // Search will automatically work on all these fields!
    return []string{"name", "description", "brand", "category"}
}
```

When user makes a request:
- `GET /products?search=laptop` 
- Automatically searches in: `name LIKE '%laptop%' OR description LIKE '%laptop%' OR brand LIKE '%laptop%' OR category LIKE '%laptop%'`

### Database Compatibility

Automatic search adapts to database:
- **MySQL/SQLite**: Uses `LIKE`
- **PostgreSQL**: Uses `ILIKE` (case-insensitive)

## 📊 Response Format

```json
{
  "code": 200,
  "message": "Data retrieved successfully",
  "data": [...],
  "pagination": {
    "page": 1,
    "per_page": 10,
    "max_page": 150,
    "total": 15,
  }
}
```

## 🎯 URL Parameters

| Parameter | Description | Example |
|-----------|-------------|---------|
| `page` | Page number (default: 1) | `page=2` |
| `limit` | Items per page (default: 10) | `limit=25` |
| `search` | Global search | `search=john` |
| `sort` | Sorting | `sort=name,desc` |
| `includes` | Preload relations (validated) | `includes=profile,posts` |
| Custom Fields | Specific filters | `name=john&active=true` |

### Advanced Include Examples

```bash
# Single relationship
GET /athletes?includes=Province&page=1&limit=10

# Multiple relationships
GET /athletes?includes=Province,Sport,PlayersEvents&page=1&limit=10

# Nested relationships (if supported)
GET /users?includes=Profile.Address,Posts.Comments&page=1&limit=10

# Combined with search and filters
GET /athletes?includes=Province,Sport&search=john&province_id=1&page=1&limit=10
```

## 🔧 Configuration Options

```go
options := pagination.PaginatedQueryOptions{
    Dialect:           pagination.MySQL,
    EnableSoftDelete:  true,
    MaxLimit:         100,
    DefaultLimit:     10,
}

result, err := pagination.PaginatedQueryWithOptions(db, filter, &users, options)
```

## 🏆 Best Practices

### 1. Use Custom Filters for Complex Logic

```go
type ProductFilter struct {
    pagination.BaseFilter
    CategoryID int     `json:"category_id" form:"category_id"`
    MinPrice   float64 `json:"min_price" form:"min_price"`
    MaxPrice   float64 `json:"max_price" form:"max_price"`
    IsActive   *bool   `json:"is_active" form:"is_active"`
}

func (f *ProductFilter) ApplyFilters(query *gorm.DB) *gorm.DB {
    if f.CategoryID > 0 {
        query = query.Where("category_id = ?", f.CategoryID)
    }
    if f.MinPrice > 0 {
        query = query.Where("price >= ?", f.MinPrice)
    }
    if f.MaxPrice > 0 {
        query = query.Where("price <= ?", f.MaxPrice)
    }
    if f.IsActive != nil {
        query = query.Where("is_active = ?", *f.IsActive)
    }
    return query
}

func (f *ProductFilter) GetSearchFields() []string {
    return []string{"name", "description", "sku", "brand"}
}
```

### 2. Optimal Search Fields

```go
// ✅ Good - Specific searchable fields
func (f *UserFilter) GetSearchFields() []string {
    return []string{"name", "email", "username"}
}

// ❌ Avoid - Too many fields can impact performance
func (f *UserFilter) GetSearchFields() []string {
    return []string{"name", "email", "phone", "address", "bio", "notes", "description"}
}
```

### 3. Handle Relationships Efficiently

```go
type OrderFilter struct {
    pagination.BaseFilter
    CustomerName string `json:"customer_name" form:"customer_name"`
    Status       string `json:"status" form:"status"`
}

func (f *OrderFilter) ApplyFilters(query *gorm.DB) *gorm.DB {
    if f.CustomerName != "" {
        // Join only when needed
        query = query.Joins("JOIN customers ON customers.id = orders.customer_id").
               Where("customers.name LIKE ?", "%"+f.CustomerName+"%")
    }
    if f.Status != "" {
        query = query.Where("orders.status = ?", f.Status)
    }
    return query
}

func (f *OrderFilter) GetSearchFields() []string {
    return []string{"orders.invoice_number", "customers.name", "customers.email"}
}
```

## 📈 Performance Tips

1. **Index Database Fields**: Ensure frequently searched fields are indexed
2. **Limit Search Fields**: Don't include too many fields in `GetSearchFields()`
3. **Use Specific Filters**: Use specific filters instead of just search
4. **Pagination Limits**: Set reasonable max limits

## 📚 Complete Examples

Check the `examples/` folder for real-world implementations:

- **AthleteFilter**: Complete filter with Province, Sport relations + include support
- **ProvinceFilter**: Simple filter with Athletes relationship + include validation  
- **SportFilter**: Filter with Athletes/Events relationships + secure includes
- **EventFilter**: Date range filter with Sport relation + include support

All examples include:
- ✅ Complete IncludableQueryBuilder implementation
- ✅ Secure include validation with `GetAllowedIncludes()`
- ✅ Real-world relationship models  
- ✅ API endpoints ready to test with includes

**Run the examples:**
```bash
cd examples/
go run .
# Server starts on :8080 with endpoints like:
# GET /athletes?includes=Province,Sport&page=1&limit=10
# GET /provinces/with-athletes?includes=Athletes
# GET /sports/with-relations?includes=Athletes,Events
```

## 🤝 Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- [GORM](https://gorm.io/) for the excellent ORM
- [Gin](https://gin-gonic.com/) for the web framework
- The Go community for continuous inspiration