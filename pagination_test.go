package pagination

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	Name  string
	Email string
}

func setupDB() *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		panic("Gagal koneksi ke database")
	}

	// Auto migrate model User
	db.AutoMigrate(&User{})
	return db
}

func setupGinContext(queryParams map[string]string) *gin.Context {
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = &http.Request{
		URL:    &url.URL{},
		Header: make(http.Header),
	}

	q := c.Request.URL.Query()
	for key, value := range queryParams {
		q.Add(key, value)
	}
	c.Request.URL.RawQuery = q.Encode()

	return c
}

func TestNew(t *testing.T) {
	db := setupDB()
	c := setupGinContext(map[string]string{
		"page[size]":   "10",
		"page[number]": "2",
	})

	p, err := New(db, c)
	assert.NoError(t, err)
	assert.Equal(t, 10, p.Req.Size)
	assert.Equal(t, 2, p.Req.Number)
}

func TestQuery(t *testing.T) {
	db := setupDB()
	c := setupGinContext(map[string]string{
		"page[size]":   "5",
		"page[number]": "1",
	})

	db.Create(&User{Name: "User 1", Email: "user1@example.com"})
	db.Create(&User{Name: "User 2", Email: "user2@example.com"})
	db.Create(&User{Name: "User 3", Email: "user3@example.com"})
	db.Create(&User{Name: "User 4", Email: "user4@example.com"})
	db.Create(&User{Name: "User 5", Email: "user5@example.com"})

	p, _ := New(db, c)
	query := p.Query()

	var result []User
	if err := query.Find(&result).Error; err != nil {
		t.Errorf("Query failed: %v", err)
	}

	assert.Equal(t, 5, len(result))
}

func TestCount(t *testing.T) {
	db := setupDB()
	c := setupGinContext(map[string]string{
		"page[size]":   "5",
		"page[number]": "1",
	})

	db.Create(&User{Name: "User 1", Email: "user1@example.com"})
	db.Create(&User{Name: "User 2", Email: "user2@example.com"})

	p, _ := New(db, c)
	err := p.Count(&User{})
	assert.NoError(t, err)
	assert.Equal(t, int64(2), p.TotalItems)
}

func TestGenerateResponse(t *testing.T) {
	db := setupDB()
	c := setupGinContext(map[string]string{
		"page[size]":   "5",
		"page[number]": "1",
	})

	db.Create(&User{Name: "User 1", Email: "user1@example.com"})
	db.Create(&User{Name: "User 2", Email: "user2@example.com"})

	p, _ := New(db, c)
	_ = p.Count(&User{})
	response := p.GenerateResponse(c)

	t.Log("First URL:", response.Links.First)
	t.Log("Last URL:", response.Links.Last)

	assert.Equal(t, 1, response.Meta.CurrentPage)
	assert.Equal(t, 5, response.Meta.PerPage)
	assert.Equal(t, 1, *response.Meta.From)
	assert.Equal(t, 2, *response.Meta.To)

	firstLink := generateLink(c.Request.URL.String(), 1, 5)
	lastLink := generateLink(c.Request.URL.String(), 1, 5)

	assert.Contains(t, response.Links.First, firstLink)
	assert.Contains(t, response.Links.Last, lastLink)
	assert.Nil(t, response.Links.Prev)
	assert.Nil(t, response.Links.Next)
}

func generateLink(baseURL string, pageNumber int, pageSize int) string {
	u, err := url.Parse(baseURL)
	if err != nil {
		panic(err)
	}
	q := u.Query()
	q.Set("page[number]", fmt.Sprintf("%d", pageNumber))
	q.Set("page[size]", fmt.Sprintf("%d", pageSize))
	u.RawQuery = q.Encode()
	return u.String()
}
