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

### 2. Custom Filter Pattern (More Flexible!)

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

// Usage in handler
func GetUsersWithFilter(db *gorm.DB) gin.HandlerFunc {
    return func(c *gin.Context) {
        var filter UserFilter
        if err := pagination.BindPagination(c, &filter); err != nil {
            c.JSON(400, gin.H{"error": err.Error()})
            return
        }

        result, err := pagination.PaginatedQuery(db, &filter, &[]User{})
        if err != nil {
            c.JSON(500, gin.H{"error": err.Error()})
            return
        }

        c.JSON(200, result)
    }
}
```

### 3. Advanced Pattern with Relationships

```go
type UserWithProfile struct {
    ID      uint    `json:"id"`
    Name    string  `json:"name"`
    Email   string  `json:"email"`
    Profile Profile `json:"profile" gorm:"foreignKey:UserID"`
}

type Profile struct {
    ID     uint   `json:"id"`
    UserID uint   `json:"user_id"`
    Bio    string `json:"bio"`
    Avatar string `json:"avatar"`
}

type UserProfileFilter struct {
    pagination.BaseFilter
    Name string `json:"name" form:"name"`
    Bio  string `json:"bio" form:"bio"`
}

func (f *UserProfileFilter) ApplyFilters(query *gorm.DB) *gorm.DB {
    if f.Name != "" {
        query = query.Where("users.name LIKE ?", "%"+f.Name+"%")
    }
    if f.Bio != "" {
        query = query.Joins("JOIN profiles ON profiles.user_id = users.id").
               Where("profiles.bio LIKE ?", "%"+f.Bio+"%")
    }
    return query
}

func (f *UserProfileFilter) GetSearchFields() []string {
    return []string{"users.name", "users.email", "profiles.bio"}
}

func (f *UserProfileFilter) GetTableName() string {
    return "users"
}

func (f *UserProfileFilter) GetDefaultSort() string {
    return "users.id asc"
}

// Handler with preload
func GetUsersWithProfiles(db *gorm.DB) gin.HandlerFunc {
    return func(c *gin.Context) {
        var filter UserProfileFilter
        filter.Includes = []string{"Profile"} // Auto preload!
        
        if err := pagination.BindPagination(c, &filter); err != nil {
            c.JSON(400, gin.H{"error": err.Error()})
            return
        }

        result, err := pagination.PaginatedQuery(db, &filter, &[]UserWithProfile{})
        if err != nil {
            c.JSON(500, gin.H{"error": err.Error()})
            return
        }

        c.JSON(200, result)
    }
}
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
    "limit": 10,
    "total_records": 150,
    "total_pages": 15,
    "has_next": true,
    "has_prev": false
  }
}
```

## 🔗 Include Relationships (New!)

This library now supports include relationships with safe validation!

### 1. IncludableQueryBuilder Interface

```go
type AthleteFilter struct {
    pagination.BaseFilter
    ID         int `json:"id" form:"id"`
    ProvinceID int `json:"province_id" form:"province_id"`
    SportID    int `json:"sport_id" form:"sport_id"`
}

// IncludableQueryBuilder implementation
func (f *AthleteFilter) ApplyFilters(query *gorm.DB) *gorm.DB {
    if f.ID > 0 {
        query = query.Where("id = ?", f.ID)
    }
    if f.ProvinceID > 0 {
        query = query.Where("province_id = ?", f.ProvinceID)
    }
    return query
}

func (f *AthleteFilter) GetTableName() string {
    return "athletes"
}

func (f *AthleteFilter) GetSearchFields() []string {
    return []string{"name"}
}

func (f *AthleteFilter) GetDefaultSort() string {
    return "id asc"
}

func (f *AthleteFilter) GetIncludes() []string {
    return f.Includes
}

func (f *AthleteFilter) GetPagination() pagination.PaginationRequest {
    return f.Pagination
}

func (f *AthleteFilter) Validate() {
    var validIncludes []string
    allowedIncludes := f.GetAllowedIncludes()
    for _, include := range f.Includes {
        if allowedIncludes[include] {
            validIncludes = append(validIncludes, include)
        }
    }
    f.Includes = validIncludes
}

func (f *AthleteFilter) GetAllowedIncludes() map[string]bool {
    return map[string]bool{
        "Province":      true,
        "Sport":         true,
        "PlayersEvents": true,
    }
}
```

### 2. Model with Relationships

```go
type Athlete struct {
    ID            int             `json:"id"`
    ProvinceID    int             `json:"province_id"`
    Province      *Province       `json:"province,omitempty"`
    SportID       int             `json:"sport_id"`
    Sport         *Sport          `json:"sport,omitempty"`
    Name          string          `json:"name"`
    Age           int             `json:"age"`
    PlayersEvents []PlayersEvents `json:"players_events,omitempty"`
}
```

### 3. Usage with IncludableQueryBuilder

```go
func GetAthletesWithIncludes(db *gorm.DB) gin.HandlerFunc {
    return func(c *gin.Context) {
        filter := &AthleteFilter{}
        filter.BindPagination(c)
        c.ShouldBindQuery(filter)

        // Automatically validate includes and load relationships
        athletes, total, err := pagination.PaginatedQueryWithIncludable[Athlete](db, filter)

        if err != nil {
            c.JSON(500, gin.H{"error": err.Error()})
            return
        }

        paginationResponse := pagination.CalculatePagination(filter.GetPagination(), total)
        response := pagination.NewPaginatedResponse(200, "Athletes retrieved successfully", athletes, paginationResponse)

        c.JSON(200, response)
    }
}
```

### 4. URL with Includes

```
GET /athletes?includes=Province,Sport&page=1&limit=10
```

The result will return athletes along with preloaded province and sport.

### 5. Security Features

- **Include Validation**: Only includes listed in `GetAllowedIncludes()` will be processed
- **SQL Injection Protection**: All includes are validated with regex patterns
- **Type Safety**: Uses interfaces to ensure required methods are available

## 🎯 URL Parameters

| Parameter | Description | Example |
|-----------|-------------|---------|
| `page` | Page number (default: 1) | `page=2` |
| `limit` | Items per page (default: 10) | `limit=25` |
| `search` | Global search | `search=john` |
| `sort` | Sorting | `sort=name,desc` |
| `include` | Preload relations | `include=profile,posts` |
| Custom Fields | Specific filters | `name=john&active=true` |

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

## 🧪 Testing

```go
func TestUserPagination(t *testing.T) {
    // Setup test database
    db := setupTestDB()
    
    filter := &UserFilter{
        BaseFilter: pagination.BaseFilter{
            Pagination: pagination.PaginationRequest{
                Page:    1,
                PerPage: 10,
            },
        },
        Name: "John",
    }
    
    users, total, err := pagination.PaginatedQuery[User](db, filter)
    
    assert.NoError(t, err)
    assert.Equal(t, int64(5), total)
    assert.Len(t, users, 5)
}
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