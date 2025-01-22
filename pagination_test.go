package pagination_test

import (
	"go-pagination"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupRouter() *gin.Engine {
	r := gin.Default()
	return r
}

func TestInitPagination(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("could not connect to the database: %v", err)
	}

	paginator := pagination.NewPagination(db)

	req := pagination.PaginationRequest{Size: 0, Number: 0}
	lengthModel := 100
	result := paginator.InitPagiantion(req, lengthModel)

	if result.Size != pagination.DEFAULT_PAGE_SIZE {
		t.Errorf("expected %v, got %v", pagination.DEFAULT_PAGE_SIZE, result.Size)
	}
	if result.Number != pagination.DEFAULT_PAGE_NUMBER {
		t.Errorf("expected %v, got %v", pagination.DEFAULT_PAGE_NUMBER, result.Number)
	}
	if result.Offset != 0 {
		t.Errorf("expected %v, got %v", 0, result.Offset)
	}
	if *result.From != 1 {
		t.Errorf("expected %v, got %v", 1, *result.From)
	}
	if *result.To != 10 {
		t.Errorf("expected %v, got %v", 10, *result.To)
	}

	req = pagination.PaginationRequest{Size: 20, Number: 2}
	result = paginator.InitPagiantion(req, lengthModel)

	if result.Size != 20 {
		t.Errorf("expected %v, got %v", 20, result.Size)
	}
	if result.Number != 2 {
		t.Errorf("expected %v, got %v", 2, result.Number)
	}
	if result.Offset != 20 {
		t.Errorf("expected %v, got %v", 20, result.Offset)
	}
	if *result.From != 21 {
		t.Errorf("expected %v, got %v", 21, *result.From)
	}
	if *result.To != 40 {
		t.Errorf("expected %v, got %v", 40, *result.To)
	}
}

func TestGeneratePaginationLinks(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("could not connect to the database: %v", err)
	}

	paginator := pagination.NewPagination(db)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/test?"+pagination.DEFAULT_PAGE_SIZE_QUERY+"=10&"+pagination.DEFAULT_PAGE_NUMBER_QUERY+"=2", nil)

	req := pagination.PaginationRequest{Size: 10, Number: 2}
	lengthModel := 100
	resp, err := paginator.GeneratePaginationLinks(c, req, lengthModel)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if resp.Meta.CurrentPage != 2 {
		t.Errorf("expected %v, got %v", 2, resp.Meta.CurrentPage)
	}
	if resp.Meta.PerPage != 10 {
		t.Errorf("expected %v, got %v", 10, resp.Meta.PerPage)
	}
	if *resp.Meta.From != 11 {
		t.Errorf("expected %v, got %v", 11, *resp.Meta.From)
	}
	if *resp.Meta.To != 20 {
		t.Errorf("expected %v, got %v", 20, *resp.Meta.To)
	}

	if resp.Links.First == "" {
		t.Error("expected first link, got empty string")
	}
	if resp.Links.Last == "" {
		t.Error("expected last link, got empty string")
	}
	if resp.Links.Next == nil {
		t.Error("expected next link, got nil")
	}
	if resp.Links.Prev == nil {
		t.Error("expected prev link, got nil")
	}
}

func TestGeneratePaginationLinksWithEdgeCases(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("could not connect to the database: %v", err)
	}

	paginator := pagination.NewPagination(db)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/test?"+pagination.DEFAULT_PAGE_SIZE_QUERY+"=10&"+pagination.DEFAULT_PAGE_NUMBER_QUERY+"=1", nil)

	req := pagination.PaginationRequest{Size: 10, Number: 1}
	lengthModel := 5
	resp, err := paginator.GeneratePaginationLinks(c, req, lengthModel)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if resp.Meta.CurrentPage != 1 {
		t.Errorf("expected %v, got %v", 1, resp.Meta.CurrentPage)
	}
	if *resp.Meta.To != 5 {
		t.Errorf("expected %v, got %v", 5, *resp.Meta.To)
	}

	if resp.Links.First == "" {
		t.Error("expected first link, got empty string")
	}
	if resp.Links.Last == "" {
		t.Error("expected last link, got empty string")
	}
	if resp.Links.Next != nil {
		t.Error("expected next link to be nil")
	}
	if resp.Links.Prev != nil {
		t.Error("expected prev link to be nil")
	}
}
