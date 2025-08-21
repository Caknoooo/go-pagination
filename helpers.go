package pagination

import (
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// PaginateWithCustomFilter provides pagination using custom filter that implements Filterable interface
func PaginateWithCustomFilter[T any](
	db *gorm.DB,
	ctx *gin.Context,
	filter Filterable,
) ([]T, PaginationResponse, error) {
	// Bind pagination from context
	if baseFilter, ok := filter.(interface{ BindPagination(*gin.Context) }); ok {
		baseFilter.BindPagination(ctx)
	}

	// Bind custom filter parameters
	if err := ctx.ShouldBindQuery(filter); err != nil {
		return nil, PaginationResponse{}, err
	}

	data, total, err := PaginatedQuery[T](db, filter, filter.GetPagination(), filter.GetIncludes())
	if err != nil {
		return nil, PaginationResponse{}, err
	}

	paginationResponse := CalculatePagination(filter.GetPagination(), total)
	return data, paginationResponse, nil
}

// PaginatedAPIResponseWithCustomFilter creates a complete API response using custom filter
func PaginatedAPIResponseWithCustomFilter[T any](
	db *gorm.DB,
	ctx *gin.Context,
	filter Filterable,
	message string,
) PaginatedResponse {
	data, paginationResponse, err := PaginateWithCustomFilter[T](db, ctx, filter)

	if err != nil {
		return NewPaginatedResponse(500, "Internal Server Error: "+err.Error(), nil, PaginationResponse{})
	}

	return NewPaginatedResponse(200, message, data, paginationResponse)
}

// CreateSearchableFilter creates a default search implementation for custom filters
func CreateSearchableFilter(searchFields []string, dialect DatabaseDialect) func(*gorm.DB, string) *gorm.DB {
	return func(query *gorm.DB, searchTerm string) *gorm.DB {
		if len(searchFields) == 0 || searchTerm == "" {
			return query
		}

		searchPattern := "%" + searchTerm + "%"
		operator := "LIKE"
		if dialect == PostgreSQL {
			operator = "ILIKE"
		}

		if len(searchFields) == 1 {
			return query.Where(searchFields[0]+" "+operator+" ?", searchPattern)
		}

		conditions := make([]string, len(searchFields))
		args := make([]interface{}, len(searchFields))

		for i, field := range searchFields {
			conditions[i] = field + " " + operator + " ?"
			args[i] = searchPattern
		}

		whereClause := "(" + strings.Join(conditions, " OR ") + ")"
		return query.Where(whereClause, args...)
	}
}

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
