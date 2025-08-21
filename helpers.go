package pagination

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// Helper functions for common pagination scenarios

// PaginateModel provides a simple way to paginate any GORM model
func PaginateModel[T any](
	db *gorm.DB,
	ctx *gin.Context,
	tableName string,
	searchFields []string,
) ([]T, PaginationResponse, error) {
	pagination := BindPagination(ctx)

	builder := NewSimpleQueryBuilder(tableName).
		WithSearchFields(searchFields...)

	data, total, err := PaginatedQuery[T](db, builder, pagination, []string{})
	if err != nil {
		return nil, PaginationResponse{}, err
	}

	paginationResponse := CalculatePagination(pagination, total)
	return data, paginationResponse, nil
}

// PaginateWithIncludes provides pagination with preloaded relationships
func PaginateWithIncludes[T any](
	db *gorm.DB,
	ctx *gin.Context,
	tableName string,
	searchFields []string,
	includes []string,
) ([]T, PaginationResponse, error) {
	pagination := BindPagination(ctx)

	builder := NewSimpleQueryBuilder(tableName).
		WithSearchFields(searchFields...)

	data, total, err := PaginatedQuery[T](db, builder, pagination, includes)
	if err != nil {
		return nil, PaginationResponse{}, err
	}

	paginationResponse := CalculatePagination(pagination, total)
	return data, paginationResponse, nil
}

// PaginateWithFilter provides pagination with custom filters
func PaginateWithFilter[T any](
	db *gorm.DB,
	ctx *gin.Context,
	tableName string,
	searchFields []string,
	filterFunc func(*gorm.DB) *gorm.DB,
) ([]T, PaginationResponse, error) {
	pagination := BindPagination(ctx)

	builder := NewSimpleQueryBuilder(tableName).
		WithSearchFields(searchFields...).
		WithFilters(filterFunc)

	data, total, err := PaginatedQuery[T](db, builder, pagination, []string{})
	if err != nil {
		return nil, PaginationResponse{}, err
	}

	paginationResponse := CalculatePagination(pagination, total)
	return data, paginationResponse, nil
}

// QuickPaginate provides the simplest way to paginate with minimal configuration
func QuickPaginate[T any](
	db *gorm.DB,
	ctx *gin.Context,
	tableName string,
) ([]T, PaginationResponse, error) {
	pagination := BindPagination(ctx)

	builder := NewSimpleQueryBuilder(tableName)

	data, total, err := PaginatedQuery[T](db, builder, pagination, []string{})
	if err != nil {
		return nil, PaginationResponse{}, err
	}

	paginationResponse := CalculatePagination(pagination, total)
	return data, paginationResponse, nil
}

// PaginatedAPIResponse creates a complete API response with pagination
func PaginatedAPIResponse[T any](
	db *gorm.DB,
	ctx *gin.Context,
	tableName string,
	searchFields []string,
	message string,
) PaginatedResponse {
	data, paginationResponse, err := PaginateModel[T](db, ctx, tableName, searchFields)

	if err != nil {
		return NewPaginatedResponse(500, "Internal Server Error: "+err.Error(), nil, PaginationResponse{})
	}

	return NewPaginatedResponse(200, message, data, paginationResponse)
}

// PaginatedAPIResponseWithIncludes creates a complete API response with pagination and includes
func PaginatedAPIResponseWithIncludes[T any](
	db *gorm.DB,
	ctx *gin.Context,
	tableName string,
	searchFields []string,
	includes []string,
	message string,
) PaginatedResponse {
	data, paginationResponse, err := PaginateWithIncludes[T](db, ctx, tableName, searchFields, includes)

	if err != nil {
		return NewPaginatedResponse(500, "Internal Server Error: "+err.Error(), nil, PaginationResponse{})
	}

	return NewPaginatedResponse(200, message, data, paginationResponse)
}
