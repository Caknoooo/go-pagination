# Go Pagination 🚀

A **powerful, flexible, and production-ready** pagination library for Go with GORM integration. Built with modern Go practices including generics, this library provides multiple patterns to implement pagination in your applications with built-in support for searching, sorting, filtering, relationships, and database security.

## ✨ Key Features

- 🚀 **Generic Support**: Full support for Go generics for type safety
- 🔍 **Smart Search**: Automatic search across multiple fields with database optimization
- 🗂️ **Advanced Filtering**: Dynamic filters with custom operators and validation
- 🔗 **Relationship Support**: Easy preloading with security validation
- 🛢️ **Multi-Database**: MySQL, PostgreSQL, SQLite, and SQL Server support
- 🛡️ **Security First**: SQL injection protection and include validation
- ⚡ **High Performance**: Optimized queries with efficient counting
- 🧪 **Production Ready**: Comprehensive test coverage and real-world examples
- 📚 **Multiple Patterns**: From simple one-liners to complex builders
- 🌐 **API Ready**: Built-in response formatting for REST APIs

## 📦 Installation

```bash
go get github.com/Caknoooo/go-pagination
```

## 📋 Table of Contents

- [Quick Start](#-quick-start)
- [Basic Usage](#-basic-usage)
- [Advanced Filtering](#-advanced-filtering)
- [Relationship Loading](#-relationship-loading)
- [Search Functionality](#-search-functionality)
- [Sorting Examples](#-sorting-examples)
- [Security Features](#-security-features)
- [URL Parameters](#-url-parameters)
- [Response Format](#-response-format)
- [Real Examples](#-real-world-examples)
- [Performance Tips](#-performance-tips)

## 🚀 Quick Start

### 1. Simplest Way - One Line Pagination! 

Perfect for getting started quickly with minimal setup:

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
        // 🎯 One line pagination with automatic search!
        response := pagination.PaginatedAPIResponse[User](
            db, c, "users", 
            []string{"name", "email"}, // fields to search in
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

**Try these URLs:**
```bash
# Basic pagination
curl "http://localhost:8080/users?page=1&limit=10"

# Search in name and email fields
curl "http://localhost:8080/users?search=john&page=1&limit=10"

# Sort by name descending
curl "http://localhost:8080/users?sort=name,desc&page=1&limit=10"

# Combined: search + sort + pagination
curl "http://localhost:8080/users?search=admin&sort=id,desc&page=2&limit=5"
```

## 🔧 Basic Usage

### Simple Pagination without Search

```go
func GetUsersBasic(db *gorm.DB) gin.HandlerFunc {
    return func(c *gin.Context) {
        var pagination pagination.PaginationRequest
        if err := c.ShouldBindQuery(&pagination); err != nil {
            c.JSON(400, gin.H{"error": err.Error()})
            return
        }

        var users []User
        var total int64

        query := db.Model(&User{})
        
        // Apply sorting if provided
        if pagination.Sort != "" {
            query = query.Order(pagination.Sort + " " + pagination.Order)
        }

        // Count total records
        query.Count(&total)

        // Apply pagination
        query.Offset(pagination.GetOffset()).Limit(pagination.GetLimit()).Find(&users)

        paginationResponse := pagination.CalculatePagination(pagination, total)
        response := pagination.NewPaginatedResponse(200, "Success", users, paginationResponse)
        
        c.JSON(200, response)
    }
}
```

### Custom Pagination Response

```go
func GetUsersCustom(db *gorm.DB) gin.HandlerFunc {
    return func(c *gin.Context) {
        var filter struct {
            pagination.PaginationRequest
            Status string `form:"status"`
            Role   string `form:"role"`
        }
        
        c.ShouldBindQuery(&filter)

        var users []User
        var total int64

        query := db.Model(&User{})
        
        // Apply custom filters
        if filter.Status != "" {
            query = query.Where("status = ?", filter.Status)
        }
        if filter.Role != "" {
            query = query.Where("role = ?", filter.Role)
        }

        // Search functionality
        if filter.Search != "" {
            query = query.Where("name LIKE ? OR email LIKE ?", 
                "%"+filter.Search+"%", "%"+filter.Search+"%")
        }

        query.Count(&total)
        query.Offset(filter.GetOffset()).Limit(filter.GetLimit()).Find(&users)

        paginationResponse := pagination.CalculatePagination(filter.PaginationRequest, total)
        
        c.JSON(200, gin.H{
            "success": true,
            "message": "Users retrieved successfully",
            "data":    users,
            "pagination": paginationResponse,
            "filters_applied": gin.H{
                "status": filter.Status,
                "role":   filter.Role,
                "search": filter.Search,
            },
        })
    }
}
```

**Usage examples:**
```bash
# Filter by status and role
curl "http://localhost:8080/users?status=active&role=admin&page=1&limit=10"

# Combine filters with search
curl "http://localhost:8080/users?status=active&search=john&page=1&limit=10"
```

## 🗂️ Advanced Filtering

### Custom Filter Pattern with Validation

Create powerful, reusable filters with automatic validation:

```go
type UserFilter struct {
    pagination.BaseFilter
    ID       int    `json:"id" form:"id"`
    Name     string `json:"name" form:"name"`
    Email    string `json:"email" form:"email"`
    IsActive *bool  `json:"is_active" form:"is_active"`
    Role     string `json:"role" form:"role"`
    MinAge   int    `json:"min_age" form:"min_age"`
    MaxAge   int    `json:"max_age" form:"max_age"`
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
    if f.Role != "" {
        query = query.Where("role = ?", f.Role)
    }
    if f.MinAge > 0 {
        query = query.Where("age >= ?", f.MinAge)
    }
    if f.MaxAge > 0 {
        query = query.Where("age <= ?", f.MaxAge)
    }
    return query
}

// Define searchable fields (will be used for global search)
func (f *UserFilter) GetSearchFields() []string {
    return []string{"name", "email", "phone"}
}

func (f *UserFilter) GetTableName() string {
    return "users"
}

func (f *UserFilter) GetDefaultSort() string {
    return "id asc"
}

// Handler using the custom filter
func GetUsersWithFilter(db *gorm.DB) gin.HandlerFunc {
    return func(c *gin.Context) {
        var filter UserFilter
        if err := pagination.BindPagination(c, &filter); err != nil {
            c.JSON(400, gin.H{"error": err.Error()})
            return
        }

        users, total, err := pagination.PaginatedQueryWithFilter[User](db, &filter)
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

**Advanced filtering examples:**
```bash
# Filter by specific user ID
curl "http://localhost:8080/users?id=123"

# Filter by name pattern
curl "http://localhost:8080/users?name=john&page=1&limit=10"

# Filter by role and status
curl "http://localhost:8080/users?role=admin&is_active=true"

# Age range filtering
curl "http://localhost:8080/users?min_age=18&max_age=65"

# Complex combined filtering
curl "http://localhost:8080/users?role=user&is_active=true&min_age=25&search=developer&sort=name,asc"
```
## 🔗 Relationship Loading

### Basic Relationship Loading with Security

```go
type User struct {
    ID      uint    `json:"id" gorm:"primaryKey"`
    Name    string  `json:"name"`
    Email   string  `json:"email"`
    Profile Profile `json:"profile,omitempty" gorm:"foreignKey:UserID"`
    Posts   []Post  `json:"posts,omitempty" gorm:"foreignKey:UserID"`
    Orders  []Order `json:"orders,omitempty" gorm:"foreignKey:UserID"`
}

type Profile struct {
    ID     uint   `json:"id"`
    UserID uint   `json:"user_id"`
    Bio    string `json:"bio"`
    Avatar string `json:"avatar"`
}

type Post struct {
    ID      uint   `json:"id"`
    UserID  uint   `json:"user_id"`
    Title   string `json:"title"`
    Content string `json:"content"`
}

type UserFilter struct {
    pagination.BaseFilter
    Name   string `json:"name" form:"name"`
    Status string `json:"status" form:"status"`
}

// Implement IncludableQueryBuilder interface
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

// 🛡️ Security: Define which relationships can be loaded
func (f *UserFilter) GetAllowedIncludes() map[string]bool {
    return map[string]bool{
        "Profile": true,  // ✅ Allow loading user profile
        "Posts":   true,  // ✅ Allow loading user posts
        "Orders":  true,  // ✅ Allow loading user orders
        // "Secrets": false, // ❌ Sensitive data - not allowed
    }
}

func (f *UserFilter) ApplyFilters(query *gorm.DB) *gorm.DB {
    if f.Name != "" {
        query = query.Where("name LIKE ?", "%"+f.Name+"%")
    }
    if f.Status != "" {
        query = query.Where("status = ?", f.Status)
    }
    return query
}

func (f *UserFilter) GetSearchFields() []string {
    return []string{"name", "email"}
}

func (f *UserFilter) GetTableName() string {
    return "users"
}

func (f *UserFilter) GetDefaultSort() string {
    return "id asc"
}

// Handler with automatic include validation
func GetUsersWithRelations(db *gorm.DB) gin.HandlerFunc {
    return func(c *gin.Context) {
        filter := &UserFilter{}
        filter.BindPagination(c)
        c.ShouldBindQuery(filter)

        // 🔒 Automatically validates includes and loads relationships
        users, total, err := pagination.PaginatedQueryWithIncludable[User](db, filter)
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

**Relationship loading examples:**
```bash
# Basic pagination without relationships
curl "http://localhost:8080/users?page=1&limit=10"

# Load user profiles
curl "http://localhost:8080/users?includes=Profile&page=1&limit=10"

# Load multiple relationships
curl "http://localhost:8080/users?includes=Profile,Posts&page=1&limit=10"

# Load all allowed relationships
curl "http://localhost:8080/users?includes=Profile,Posts,Orders&page=1&limit=10"

# Combine with search and filters
curl "http://localhost:8080/users?includes=Profile,Posts&search=john&status=active&page=1&limit=10"

# Try loading unauthorized relationship (will be ignored)
curl "http://localhost:8080/users?includes=Profile,Secrets&page=1&limit=10"
# Only Profile will be loaded, Secrets will be ignored for security
```

### Advanced Relationships with Nested Loading

```go
type UserAdvancedFilter struct {
    pagination.BaseFilter
    Name      string `json:"name" form:"name"`
    CityName  string `json:"city_name" form:"city_name"`
    PostTitle string `json:"post_title" form:"post_title"`
}

func (f *UserAdvancedFilter) ApplyFilters(query *gorm.DB) *gorm.DB {
    if f.Name != "" {
        query = query.Where("users.name LIKE ?", "%"+f.Name+"%")
    }
    if f.CityName != "" {
        query = query.Joins("JOIN profiles ON profiles.user_id = users.id").
               Joins("JOIN addresses ON addresses.profile_id = profiles.id").
               Where("addresses.city LIKE ?", "%"+f.CityName+"%")
    }
    if f.PostTitle != "" {
        query = query.Joins("JOIN posts ON posts.user_id = users.id").
               Where("posts.title LIKE ?", "%"+f.PostTitle+"%")
    }
    return query
}

func (f *UserAdvancedFilter) GetSearchFields() []string {
    return []string{"users.name", "users.email", "profiles.bio"}
}

func (f *UserAdvancedFilter) GetAllowedIncludes() map[string]bool {
    return map[string]bool{
        "Profile":         true, // Load user profile
        "Posts":           true, // Load user posts
        "Profile.Address": true, // Load nested: profile with address
        "Posts.Comments":  true, // Load nested: posts with comments
        "Posts.Tags":      true, // Load nested: posts with tags
    }
}
```

**Nested relationship examples:**
```bash
# Load nested relationships
curl "http://localhost:8080/users/advanced?includes=Profile.Address,Posts.Comments"

# Complex filtering with nested loading
curl "http://localhost:8080/users/advanced?includes=Profile.Address&city_name=Jakarta&search=developer"

# Multiple nested relationships
curl "http://localhost:8080/users/advanced?includes=Profile.Address,Posts.Comments,Posts.Tags&page=1&limit=5"
```
## 🔍 Search Functionality

### Automatic Search with Multiple Fields

The library provides powerful automatic search functionality across multiple fields:

```go
type ProductFilter struct {
    pagination.BaseFilter
    CategoryID int     `json:"category_id" form:"category_id"`
    MinPrice   float64 `json:"min_price" form:"min_price"`
    MaxPrice   float64 `json:"max_price" form:"max_price"`
    InStock    *bool   `json:"in_stock" form:"in_stock"`
}

// Define which fields should be searchable
func (f *ProductFilter) GetSearchFields() []string {
    return []string{"name", "description", "brand", "sku"}
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
    if f.InStock != nil {
        query = query.Where("in_stock = ?", *f.InStock)
    }
    return query
}

func (f *ProductFilter) GetTableName() string {
    return "products"
}

func (f *ProductFilter) GetDefaultSort() string {
    return "created_at desc"
}
```

**Search examples:**
```bash
# Search across name, description, brand, and sku fields
curl "http://localhost:8080/products?search=laptop"
# Automatically generates: WHERE (name LIKE '%laptop%' OR description LIKE '%laptop%' OR brand LIKE '%laptop%' OR sku LIKE '%laptop%')

# Combine search with filters
curl "http://localhost:8080/products?search=gaming&category_id=1&min_price=500"

# Search with pagination and sorting
curl "http://localhost:8080/products?search=macbook&sort=price,asc&page=1&limit=10"
```

### Database-Specific Search Optimization

The library automatically optimizes search based on your database:

```go
// For PostgreSQL - Uses ILIKE for case-insensitive search
// Automatically generated: WHERE (name ILIKE '%search%' OR description ILIKE '%search%')

// For MySQL/SQLite - Uses LIKE 
// Automatically generated: WHERE (name LIKE '%search%' OR description LIKE '%search%')
```

### Advanced Search with Relationships

```go
type UserSearchFilter struct {
    pagination.BaseFilter
    Role       string `json:"role" form:"role"`
    Department string `json:"department" form:"department"`
}

func (f *UserSearchFilter) GetSearchFields() []string {
    return []string{
        "users.name", 
        "users.email", 
        "profiles.bio", 
        "departments.name",
    }
}

func (f *UserSearchFilter) ApplyFilters(query *gorm.DB) *gorm.DB {
    // Join tables for search functionality
    query = query.Joins("LEFT JOIN profiles ON profiles.user_id = users.id").
           Joins("LEFT JOIN departments ON departments.id = users.department_id")
    
    if f.Role != "" {
        query = query.Where("users.role = ?", f.Role)
    }
    if f.Department != "" {
        query = query.Where("departments.name = ?", f.Department)
    }
    return query
}
```

**Advanced search examples:**
```bash
# Search across multiple tables
curl "http://localhost:8080/users/search?search=developer"
# Searches in: users.name, users.email, profiles.bio, departments.name

# Search with relationship filters
curl "http://localhost:8080/users/search?search=john&role=admin&department=IT"
```

## 🔄 Sorting Examples

### Basic Sorting

```bash
# Sort by single field ascending (default)
curl "http://localhost:8080/users?sort=name"

# Sort by single field descending
curl "http://localhost:8080/users?sort=name,desc"

# Sort by multiple fields
curl "http://localhost:8080/users?sort=role,asc&sort=name,desc"

# Sort with pagination
curl "http://localhost:8080/users?sort=created_at,desc&page=1&limit=20"
```


**Custom sorting examples:**
```bash
# Sort by calculated posts count
curl "http://localhost:8080/users?sort=posts_count,desc"

# Sort by concatenated full name
curl "http://localhost:8080/users?sort=full_name,asc"

# Sort by related table field
curl "http://localhost:8080/users?sort=latest_login,desc"
```
## 🛡️ Security Features

### Include Validation and SQL Injection Protection

```go
type SecureUserFilter struct {
    pagination.BaseFilter
    Status string `json:"status" form:"status"`
}

func (f *SecureUserFilter) GetAllowedIncludes() map[string]bool {
    return map[string]bool{
        "Profile":        true,  // ✅ Safe to load
        "Posts":          true,  // ✅ Safe to load
        "PublicData":     true,  // ✅ Safe to load
        "SensitiveData":  false, // ❌ Blocked for security
        "PrivateNotes":   false, // ❌ Blocked for security
        "AdminData":      false, // ❌ Blocked for security
    }
}

func (f *SecureUserFilter) Validate() {
    // Automatic validation removes unauthorized includes
    var validIncludes []string
    allowedIncludes := f.GetAllowedIncludes()
    
    for _, include := range f.Includes {
        // Validate against whitelist
        if allowedIncludes[include] {
            // Additional regex validation for SQL injection protection
            if isValidInclude(include) {
                validIncludes = append(validIncludes, include)
            }
        }
    }
    f.Includes = validIncludes
}

// Built-in regex validation for includes
func isValidInclude(include string) bool {
    // Only allows alphanumeric, dots, and underscores
    matched, _ := regexp.MatchString(`^[a-zA-Z0-9_.]+$`, include)
    return matched
}
```

**Security examples:**
```bash
# Valid includes - will be processed
curl "http://localhost:8080/users?includes=Profile,Posts"

# Invalid includes - will be ignored
curl "http://localhost:8080/users?includes=Profile,SensitiveData,AdminData"
# Only Profile will be loaded

# SQL injection attempt - will be blocked
curl "http://localhost:8080/users?includes=Profile'; DROP TABLE users; --"
# Regex validation will reject this
```

### Input Validation and Sanitization

```go
type ValidatedFilter struct {
    pagination.BaseFilter
    Email    string `json:"email" form:"email" validate:"email"`
    Age      int    `json:"age" form:"age" validate:"min=0,max=120"`
    Status   string `json:"status" form:"status" validate:"oneof=active inactive pending"`
}

func (f *ValidatedFilter) Validate() error {
    // Built-in validation
    if f.Email != "" && !isValidEmail(f.Email) {
        return errors.New("invalid email format")
    }
    
    if f.Age < 0 || f.Age > 120 {
        return errors.New("age must be between 0 and 120")
    }
    
    validStatuses := map[string]bool{
        "active": true, "inactive": true, "pending": true,
    }
    if f.Status != "" && !validStatuses[f.Status] {
        return errors.New("invalid status value")
    }
    
    return nil
}

func (f *ValidatedFilter) ApplyFilters(query *gorm.DB) *gorm.DB {
    // All inputs are already validated
    if f.Email != "" {
        query = query.Where("email = ?", f.Email) // Safe to use
    }
    if f.Age > 0 {
        query = query.Where("age = ?", f.Age)
    }
    if f.Status != "" {
        query = query.Where("status = ?", f.Status)
    }
    return query
}
```

## � URL Parameters Reference

### Core Parameters

| Parameter | Type | Description | Example | Default |
|-----------|------|-------------|---------|---------|
| `page` | int | Page number | `page=2` | 1 |
| `limit` | int | Items per page | `limit=25` | 10 |
| `per_page` | int | Alias for limit | `per_page=25` | 10 |
| `search` | string | Global search term | `search=john` | "" |
| `sort` | string | Sort field | `sort=name` | "" |
| `order` | string | Sort direction | `order=desc` | "asc" |
| `includes` | string | Comma-separated relations | `includes=profile,posts` | "" |

### Sorting Formats

```bash
# Single field ascending (default)
?sort=name

# Single field with explicit direction
?sort=name&order=desc

# Alternative comma format
?sort=name,desc

# Multiple fields
?sort=name,asc&sort=created_at,desc
```

### Complex Query Examples

```bash
# Basic pagination
GET /api/users?page=1&limit=10

# Search with pagination
GET /api/users?search=developer&page=2&limit=5

# Filter with specific fields
GET /api/users?role=admin&status=active&page=1&limit=20

# Sort with relationships
GET /api/users?includes=profile,posts&sort=name,asc&page=1&limit=15

# Complex combined query
GET /api/users?search=john&role=user&status=active&includes=profile&sort=created_at,desc&page=1&limit=10

# Date range filtering (custom implementation)
GET /api/users?created_after=2023-01-01&created_before=2023-12-31&page=1&limit=10

# Numeric range filtering
GET /api/products?min_price=100&max_price=500&category_id=1&page=1&limit=10
```

## � Response Format

### Standard Response Structure

```json
{
  "code": 200,
  "status": "success", 
  "message": "Data retrieved successfully",
  "data": [
    {
      "id": 1,
      "name": "John Doe",
      "email": "john@example.com",
      "profile": {
        "id": 1,
        "bio": "Software Developer",
        "avatar": "avatar.jpg"
      },
      "posts": [
        {
          "id": 1,
          "title": "My First Post",
          "content": "Hello World!"
        }
      ]
    }
  ],
  "pagination": {
    "page": 1,
    "per_page": 10,
    "max_page": 15,
    "total": 142
  }
}
```

## 🚀 Running the Examples

The `examples/` folder contains a complete working implementation:

```bash
# Navigate to examples directory
cd examples/

# Install dependencies
go mod tidy

# Set up your database (update connection string in main.go)
# Default expects MySQL at localhost:3306 with database 'sports_db'

# Run the example server
go run .
```

The example server provides these endpoints:

- `GET /provinces` - Basic province pagination
- `GET /provinces/with-athletes` - Provinces with athlete relationships
- `GET /athletes` - Athletes with province/sport filtering
- `GET /sports` - Sports management
- `GET /events` - Events with date filtering

**Test the examples:**
```bash
# Basic athlete pagination
curl "http://localhost:8080/athletes?page=1&limit=5"

# Athletes with relationships
curl "http://localhost:8080/athletes?includes=Province,Sport&page=1&limit=5"

# Search athletes by name
curl "http://localhost:8080/athletes?search=john&includes=Province"

# Filter by province and sport
curl "http://localhost:8080/athletes?province_id=1&sport_id=2&includes=Province,Sport"

# Provinces with their athletes
curl "http://localhost:8080/provinces/with-athletes?includes=Athletes&page=1&limit=10"
```

## 🏆 Best Practices

### 1. Filter Design

```go
// ✅ Good - Clear, specific filters
type UserFilter struct {
    pagination.BaseFilter
    Status   string    `json:"status" form:"status" validate:"oneof=active inactive"`
    Role     string    `json:"role" form:"role" validate:"oneof=admin user moderator"`
    IsActive *bool     `json:"is_active" form:"is_active"`
    MinAge   int       `json:"min_age" form:"min_age" validate:"min=0,max=120"`
}

// ❌ Avoid - Generic string filters
type BadFilter struct {
    pagination.BaseFilter
    Filter1 string `json:"filter1" form:"filter1"` // Unclear purpose
    Filter2 string `json:"filter2" form:"filter2"` // No validation
}
```

### 2. Search Field Selection

```go
// ✅ Good - Strategic search fields
func (f *ProductFilter) GetSearchFields() []string {
    return []string{"name", "sku", "brand"} // Most relevant for users
}

// ✅ Good - Include table prefixes for joins
func (f *OrderFilter) GetSearchFields() []string {
    return []string{"orders.invoice_number", "customers.name", "customers.email"}
}
```

### 3. Security Implementation

```go
// ✅ Good - Whitelist approach
func (f *UserFilter) GetAllowedIncludes() map[string]bool {
    return map[string]bool{
        "Profile":    true,  // Safe public data
        "Posts":      true,  // User's own posts
    }
    // Note: Sensitive data like "BankAccount", "SSN" not included
}

// ✅ Good - Validate input in ApplyFilters
func (f *UserFilter) ApplyFilters(query *gorm.DB) *gorm.DB {
    if f.Status != "" {
        // Validate against allowed values
        allowedStatuses := []string{"active", "inactive", "pending"}
        if contains(allowedStatuses, f.Status) {
            query = query.Where("status = ?", f.Status)
        }
    }
    return query
}
```

### 4. Error Handling

```go
func GetUsersHandler(db *gorm.DB) gin.HandlerFunc {
    return func(c *gin.Context) {
        filter := &UserFilter{}
        
        // Validate input binding
        if err := filter.BindPagination(c); err != nil {
            c.JSON(400, gin.H{
                "error": "Invalid pagination parameters",
                "details": err.Error(),
            })
            return
        }
        
        // Validate custom filter fields
        if err := filter.Validate(); err != nil {
            c.JSON(400, gin.H{
                "error": "Invalid filter parameters", 
                "details": err.Error(),
            })
            return
        }

        users, total, err := pagination.PaginatedQueryWithIncludable[User](db, filter)
        if err != nil {
            log.Printf("Database error: %v", err)
            c.JSON(500, gin.H{
                "error": "Internal server error",
                "code": "DB_ERROR",
            })
            return
        }

        paginationResponse := pagination.CalculatePagination(filter.GetPagination(), total)
        response := pagination.NewPaginatedResponse(200, "Users retrieved successfully", users, paginationResponse)
        c.JSON(200, response)
    }
}
```

### 5. Testing Your Implementation

```go
func TestUserPagination(t *testing.T) {
    // Test basic pagination
    filter := &UserFilter{}
    filter.Pagination.Page = 1
    filter.Pagination.PerPage = 10
    
    users, total, err := pagination.PaginatedQueryWithFilter[User](db, filter)
    assert.NoError(t, err)
    assert.LessOrEqual(t, len(users), 10)
    
    // Test search functionality
    filter.Search = "john"
    users, total, err = pagination.PaginatedQueryWithFilter[User](db, filter)
    assert.NoError(t, err)
    
    // Test include validation
    filter.Includes = []string{"Profile", "InvalidRelation"}
    filter.Validate()
    assert.Equal(t, []string{"Profile"}, filter.Includes)
}
```

## 🤝 Contributing

We welcome contributions! Here's how you can help:

### Development Setup

```bash
# Fork and clone the repository
git clone https://github.com/yourusername/go-pagination.git
cd go-pagination

# Install dependencies
go mod tidy

# Run tests
go test ./...

# Run benchmarks
go test -bench=. ./...

# Run examples
cd examples/
go run .
```

### Contribution Guidelines

1. **Fork the repository**
2. **Create a feature branch**: `git checkout -b feature/amazing-feature`
3. **Make your changes**:
   - Add tests for new functionality
   - Update documentation
   - Follow Go conventions
4. **Run tests**: `go test ./...`
5. **Commit your changes**: `git commit -m 'Add amazing feature'`
6. **Push to branch**: `git push origin feature/amazing-feature`
7. **Open a Pull Request**

### What We're Looking For

- 🐛 **Bug fixes**
- ✨ **New features** (with tests)
- 📚 **Documentation improvements**
- 🚀 **Performance optimizations**
- 🧪 **Additional test coverage**
- 💡 **Usage examples**

### Code Style

- Follow standard Go formatting (`gofmt`)
- Add comments for exported functions
- Write meaningful test cases
- Update README for new features

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- **[GORM](https://gorm.io/)** - The fantastic Go ORM that makes database operations elegant
- **[Gin](https://gin-gonic.com/)** - The high-performance Go web framework
- **Go Community** - For continuous inspiration and feedback
- **Contributors** - Everyone who has contributed to making this library better

## 📞 Support

- 📖 **Documentation**: Check this README and examples
- 🐛 **Issues**: [GitHub Issues](https://github.com/Caknoooo/go-pagination/issues)
- 💬 **Discussions**: [GitHub Discussions](https://github.com/Caknoooo/go-pagination/discussions)
- ⭐ **Star the repo** if you find it useful!