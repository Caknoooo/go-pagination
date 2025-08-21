package pagination

import (
	"fmt"
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupBenchmarkDB(recordCount int) *gorm.DB {
	db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	db.AutoMigrate(&TestUser{})

	batchSize := 1000
	for i := 0; i < recordCount; i += batchSize {
		var users []TestUser
		end := i + batchSize
		if end > recordCount {
			end = recordCount
		}

		for j := i; j < end; j++ {
			users = append(users, TestUser{
				Name:  fmt.Sprintf("User %d", j),
				Email: fmt.Sprintf("user%d@example.com", j),
				Age:   20 + (j % 50),
			})
		}
		db.CreateInBatches(users, batchSize)
	}

	return db
}

func BenchmarkPaginatedQuery_1000Records(b *testing.B) {
	db := setupBenchmarkDB(1000)
	builder := NewSimpleQueryBuilder("test_users").
		WithSearchFields("name", "email").
		WithDefaultSort("id asc")

	pagination := PaginationRequest{Page: 1, PerPage: 20}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _ = PaginatedQuery[TestUser](db, builder, pagination, []string{})
	}
}

func BenchmarkPaginatedQuery_10000Records(b *testing.B) {
	db := setupBenchmarkDB(10000)
	builder := NewSimpleQueryBuilder("test_users").
		WithSearchFields("name", "email").
		WithDefaultSort("id asc")

	pagination := PaginationRequest{Page: 1, PerPage: 20}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _ = PaginatedQuery[TestUser](db, builder, pagination, []string{})
	}
}

func BenchmarkPaginatedQuery_WithSearch(b *testing.B) {
	db := setupBenchmarkDB(5000)
	builder := NewSimpleQueryBuilder("test_users").
		WithSearchFields("name", "email").
		WithDefaultSort("id asc")

	pagination := PaginationRequest{Page: 1, PerPage: 20, Search: "User 100"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _ = PaginatedQuery[TestUser](db, builder, pagination, []string{})
	}
}

func BenchmarkPaginatedQuery_WithFilters(b *testing.B) {
	db := setupBenchmarkDB(5000)
	builder := NewSimpleQueryBuilder("test_users").
		WithSearchFields("name", "email").
		WithDefaultSort("id asc").
		WithFilters(func(query *gorm.DB) *gorm.DB {
			return query.Where("age > ?", 30)
		})

	pagination := PaginationRequest{Page: 1, PerPage: 20}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _ = PaginatedQuery[TestUser](db, builder, pagination, []string{})
	}
}

func BenchmarkDynamicFilter(b *testing.B) {
	db := setupBenchmarkDB(5000)
	filter := &DynamicFilter{
		TableName:    "test_users",
		Model:        TestUser{},
		SearchFields: []string{"name", "email"},
		DefaultSort:  "id asc",
		Filters: []FilterCondition{
			{Field: "age", Operator: ">", Value: 30, Logic: "AND"},
		},
	}

	pagination := PaginationRequest{Page: 1, PerPage: 20}
	filter.Pagination = pagination

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _ = PaginatedQuery[TestUser](db, filter, pagination, []string{})
	}
}

func BenchmarkChainableQueryBuilder(b *testing.B) {
	db := setupBenchmarkDB(5000)
	builder := NewChainableQueryBuilder("test_users").
		WithSearchFields("name", "email").
		WithDefaultSort("id asc").
		WithFilters(func(query *gorm.DB) *gorm.DB {
			return query.Where("age > ?", 25)
		})

	pagination := PaginationRequest{Page: 1, PerPage: 20}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _ = PaginatedQuery[TestUser](db, builder, pagination, []string{})
	}
}

func BenchmarkPaginatedQuery_PageSize10(b *testing.B) {
	db := setupBenchmarkDB(1000)
	builder := NewSimpleQueryBuilder("test_users")
	pagination := PaginationRequest{Page: 1, PerPage: 10}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _ = PaginatedQuery[TestUser](db, builder, pagination, []string{})
	}
}

func BenchmarkPaginatedQuery_PageSize50(b *testing.B) {
	db := setupBenchmarkDB(1000)
	builder := NewSimpleQueryBuilder("test_users")
	pagination := PaginationRequest{Page: 1, PerPage: 50}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _ = PaginatedQuery[TestUser](db, builder, pagination, []string{})
	}
}

func BenchmarkPaginatedQuery_PageSize100(b *testing.B) {
	db := setupBenchmarkDB(1000)
	builder := NewSimpleQueryBuilder("test_users")
	pagination := PaginationRequest{Page: 1, PerPage: 100}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _ = PaginatedQuery[TestUser](db, builder, pagination, []string{})
	}
}
