# Go Pagination 🚀

Library pagination yang **dinamis, fleksibel, dan mudah digunakan** untuk Go dengan integrasi GORM. Library ini menyediakan berbagai pendekatan untuk mengimplementasikan pagination dalam aplikasi Go Anda dengan dukungan built-in untuk pencarian, sorting, filtering, dan berbagai database.

## ✨ Fitur Utama

- 🚀 **Generic Support**: Dukungan penuh untuk Go generics
- 🔍 **Smart Search**: Pencarian otomatis berdasarkan field yang ditentukan
- 🗂️ **Dynamic Filtering**: Filter dinamis dengan berbagai operator
- 🔗 **Relationship Support**: Preloading relasi dengan mudah
- 🛢️ **Multi-Database**: Support MySQL, PostgreSQL, SQLite, dan SQL Server
- 🛡️ **SQL Injection Protection**: Validasi built-in untuk mencegah SQL injection
- ⚡ **Performance Optimized**: Query count dan data yang efisien
- 🧪 **Well Tested**: Test coverage yang komprehensif
- 📚 **Multiple Patterns**: Dari simple helper sampai advanced builder

## 📦 Installation

```bash
go get github.com/Caknoooo/go-pagination
```

## 🚀 Quick Start

### 1. Simple API Response (Paling Mudah!)

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
        // Satu baris untuk pagination dengan search!
        response := pagination.PaginatedAPIResponse[User](
            db, c, "users", 
            []string{"name", "email"}, // field yang bisa di-search
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
- `GET /users?search=john&page=1&limit=10` - Search di name & email
- `GET /users?sort=name,desc&page=1&limit=10` - Sort by name descending

### 2. Custom Filter Pattern (Lebih Fleksibel!)

```go
type UserFilter struct {
    pagination.BaseFilter
    ID       int    `json:"id" form:"id"`
    Name     string `json:"name" form:"name"`
    Email    string `json:"email" form:"email"`
    IsActive *bool  `json:"is_active" form:"is_active"`
}

// Implementasi filter custom
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

// Tentukan field yang bisa di-search (DINAMIS!)
func (f *UserFilter) GetSearchFields() []string {
    return []string{"name", "email", "phone"}
}

func (f *UserFilter) GetTableName() string {
    return "users"
}

func (f *UserFilter) GetDefaultSort() string {
    return "id asc"
}

// Penggunaan di handler
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

### 3. Advanced Pattern dengan Relationships

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

// Handler dengan preload
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

### Automatic Search dengan GetSearchFields

Yang membuat library ini **sangat dinamis** adalah kemampuan `GetSearchFields()`:

```go
func (f *ProductFilter) GetSearchFields() []string {
    // Search akan otomatis bekerja di semua field ini!
    return []string{"name", "description", "brand", "category"}
}
```

Ketika user melakukan request:
- `GET /products?search=laptop` 
- Otomatis akan mencari di: `name LIKE '%laptop%' OR description LIKE '%laptop%' OR brand LIKE '%laptop%' OR category LIKE '%laptop%'`

### Database Compatibility

Search otomatis menyesuaikan dengan database:
- **MySQL/SQLite**: Menggunakan `LIKE`
- **PostgreSQL**: Menggunakan `ILIKE` (case-insensitive)

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

## 🎯 URL Parameters

| Parameter | Description | Example |
|-----------|-------------|---------|
| `page` | Halaman (default: 1) | `page=2` |
| `limit` | Jumlah per halaman (default: 10) | `limit=25` |
| `search` | Pencarian global | `search=john` |
| `sort` | Sorting | `sort=name,desc` |
| `include` | Preload relations | `include=profile,posts` |
| Custom Fields | Filter spesifik | `name=john&active=true` |

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

### 1. Gunakan Custom Filter untuk Logic Kompleks

```go
type ProductFilter struct {
    pagination.BaseFilter
    CategoryID  int     `json:"category_id" form:"category_id"`
    MinPrice    float64 `json:"min_price" form:"min_price"`
    MaxPrice    float64 `json:"max_price" form:"max_price"`
    IsActive    *bool   `json:"is_active" form:"is_active"`
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

1. **Index Database Fields**: Pastikan field yang sering di-search ter-index
2. **Limit Search Fields**: Jangan terlalu banyak field di `GetSearchFields()`
3. **Use Specific Filters**: Gunakan filter spesifik daripada hanya search
4. **Pagination Limits**: Set reasonable max limits

## 🧪 Testing

```go
func TestPagination(t *testing.T) {
    db := setupTestDB()
    
    filter := &UserFilter{
        BaseFilter: pagination.BaseFilter{
            Pagination: pagination.PaginationRequest{
                Page:   1,
                Limit:  10,
                Search: "john",
            },
        },
    }
    
    result, err := pagination.PaginatedQuery(db, filter, &[]User{})
    assert.NoError(t, err)
    assert.Equal(t, 1, result.Pagination.Page)
    assert.True(t, len(result.Data.([]User)) <= 10)
}
```

## 📚 Examples Repository

Lihat folder `examples/` untuk implementasi lengkap:

- **AthleteFilter**: Filter dengan relasi Province dan Sport
- **ProvinceFilter**: Filter sederhana dengan name dan code
- **SportFilter**: Filter dengan category dan description
- **EventFilter**: Filter dengan date range dan sport relation

## 🤝 Contributing

1. Fork repository
2. Create feature branch
3. Add tests for new features
4. Submit pull request

## 📄 License

MIT License

---

**Go Pagination** - Making pagination **dinamis serta mudah digunakan**! 🚀
}
```

#### PaginateModel
