package pagination

import (
	"fmt"
	"strings"

	"gorm.io/gorm"
)

type QueryBuilder interface {
	ApplyFilters(query *gorm.DB) *gorm.DB
	GetTableName() string
	GetDefaultSort() string
	GetSearchFields() []string
}

type IncludableQueryBuilder interface {
	QueryBuilder
	GetIncludes() []string
	GetPagination() PaginationRequest
	Validate()
}

type AllowedIncludesProvider interface {
	GetAllowedIncludes() map[string]bool
}

// DatabaseProvider interface for query builders that need database access
type DatabaseProvider interface {
	GetDB() *gorm.DB
}

// QueryLayerBuilder interface that combines query building with database access
type QueryLayerBuilder interface {
	IncludableQueryBuilder
	DatabaseProvider
}

// applyAutoSearch applies search automatically based on provided search fields
func applyAutoSearch(query *gorm.DB, searchTerm string, searchFields []string, dialect DatabaseDialect) *gorm.DB {
	if len(searchFields) == 0 || searchTerm == "" {
		return query
	}

	searchPattern := "%" + searchTerm + "%"
	operator := getSearchOperator(dialect)

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

func getSearchOperator(dialect DatabaseDialect) string {
	switch dialect {
	case PostgreSQL:
		return "ILIKE"
	case MySQL, SQLite, SQLServer:
		return "LIKE"
	default:
		return "LIKE"
	}
}

// DatabaseDialect represents different database types for compatibility
type DatabaseDialect string

const (
	MySQL      DatabaseDialect = "mysql"
	PostgreSQL DatabaseDialect = "postgresql"
	SQLite     DatabaseDialect = "sqlite"
	SQLServer  DatabaseDialect = "sqlserver"
)

// PaginatedQueryOptions provides configuration for paginated queries
type PaginatedQueryOptions struct {
	Dialect          DatabaseDialect
	EnableSoftDelete bool
	CustomCountQuery string
}

func PaginatedQuery[T any](
	db *gorm.DB,
	builder QueryBuilder,
	pagination PaginationRequest,
	includes []string,
) ([]T, int64, error) {
	return PaginatedQueryWithOptions[T](db, builder, pagination, includes, PaginatedQueryOptions{
		Dialect: MySQL, // Default to MySQL for backward compatibility
	})
}

// PaginatedQueryWithIncludable handles queries with includable query builders
func PaginatedQueryWithIncludable[T any](
	db *gorm.DB,
	builder IncludableQueryBuilder,
) ([]T, int64, error) {
	// If db is nil, try to get it from the builder (for query layer pattern)
	if db == nil {
		if dbProvider, ok := builder.(DatabaseProvider); ok {
			db = dbProvider.GetDB()
		} else {
			return nil, 0, fmt.Errorf("database connection not provided")
		}
	}

	// Validate the builder
	builder.Validate()

	// Get pagination and includes from the builder
	pagination := builder.GetPagination()
	includes := builder.GetIncludes()

	return PaginatedQueryWithOptions[T](db, builder, pagination, includes, PaginatedQueryOptions{
		Dialect: MySQL, // Default to MySQL for backward compatibility
	})
}

// PaginatedQueryWithIncludableAndOptions handles queries with includable query builders and custom options
func PaginatedQueryWithIncludableAndOptions[T any](
	db *gorm.DB,
	builder IncludableQueryBuilder,
	options PaginatedQueryOptions,
) ([]T, int64, error) {
	// Validate the builder
	builder.Validate()

	// Get pagination and includes from the builder
	pagination := builder.GetPagination()
	includes := builder.GetIncludes()

	return PaginatedQueryWithOptions[T](db, builder, pagination, includes, options)
}

func PaginatedQueryWithOptions[T any](
	db *gorm.DB,
	builder QueryBuilder,
	pagination PaginationRequest,
	includes []string,
	options PaginatedQueryOptions,
) ([]T, int64, error) {
	var result []T
	var totalCount int64

	// Build count query
	countQuery := db.Table(builder.GetTableName())
	countQuery = builder.ApplyFilters(countQuery)

	// Apply soft delete handling if enabled
	if options.EnableSoftDelete {
		countQuery = countQuery.Where("deleted_at IS NULL")
	}

	// Execute count query
	if options.CustomCountQuery != "" {
		if err := countQuery.Raw(options.CustomCountQuery).Count(&totalCount).Error; err != nil {
			return nil, 0, fmt.Errorf("failed to count records: %w", err)
		}
	} else {
		if err := countQuery.Count(&totalCount).Error; err != nil {
			return nil, 0, fmt.Errorf("failed to count records: %w", err)
		}
	}

	// Build data query
	dataQuery := db.Table(builder.GetTableName())
	dataQuery = builder.ApplyFilters(dataQuery)

	if pagination.Search != "" {
		dataQuery = applyAutoSearch(dataQuery, pagination.Search, builder.GetSearchFields(), options.Dialect)
	}

	// Apply soft delete handling if enabled
	if options.EnableSoftDelete {
		dataQuery = dataQuery.Where("deleted_at IS NULL")
	}

	// Apply sorting
	if pagination.Sort != "" {
		// Validate sort field to prevent SQL injection
		if isValidSortField(pagination.Sort) {
			orderClause := pagination.Sort + " " + pagination.Order
			dataQuery = dataQuery.Order(orderClause)
		} else {
			dataQuery = dataQuery.Order(builder.GetDefaultSort())
		}
	} else {
		dataQuery = dataQuery.Order(builder.GetDefaultSort())
	}

	// Apply pagination
	dataQuery = dataQuery.Offset(pagination.GetOffset()).Limit(pagination.GetLimit())

	// Validate and apply preloads
	validatedIncludes := validateIncludes(builder, includes)
	for _, include := range validatedIncludes {
		dataQuery = dataQuery.Preload(include)
	}

	// Execute data query
	if err := dataQuery.Find(&result).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to fetch records: %w", err)
	}

	return result, totalCount, nil
}

// isValidSortField validates sort field to prevent SQL injection
func isValidSortField(field string) bool {
	// Allow only alphanumeric characters, underscores, and dots
	for _, char := range field {
		if !((char >= 'a' && char <= 'z') ||
			(char >= 'A' && char <= 'Z') ||
			(char >= '0' && char <= '9') ||
			char == '_' || char == '.') {
			return false
		}
	}
	return len(field) > 0
}

// isValidInclude validates include field to prevent SQL injection
func isValidInclude(include string) bool {
	// Allow only alphanumeric characters, underscores, and dots
	for _, char := range include {
		if !((char >= 'a' && char <= 'z') ||
			(char >= 'A' && char <= 'Z') ||
			(char >= '0' && char <= '9') ||
			char == '_' || char == '.') {
			return false
		}
	}
	return len(include) > 0
}

// validateIncludes validates includes against allowed includes for the builder
func validateIncludes(builder interface{}, includes []string) []string {
	if includeValidator, ok := builder.(AllowedIncludesProvider); ok {
		allowedIncludes := includeValidator.GetAllowedIncludes()
		var validIncludes []string
		for _, include := range includes {
			if isValidInclude(include) && allowedIncludes[include] {
				validIncludes = append(validIncludes, include)
			}
		}
		return validIncludes
	}

	// Fallback: just validate syntax if no allowed includes defined
	var validIncludes []string
	for _, include := range includes {
		if isValidInclude(include) {
			validIncludes = append(validIncludes, include)
		}
	}
	return validIncludes
}

type SimpleQueryBuilder struct {
	TableName    string
	FilterFunc   func(*gorm.DB) *gorm.DB
	SearchFields []string
	DefaultSort  string
	Dialect      DatabaseDialect
}

func (s *SimpleQueryBuilder) ApplyFilters(query *gorm.DB) *gorm.DB {
	if s.FilterFunc != nil {
		return s.FilterFunc(query)
	}
	return query
}

func (s *SimpleQueryBuilder) GetSearchFields() []string {
	return s.SearchFields
}

func (s *SimpleQueryBuilder) GetTableName() string {
	return s.TableName
}

func (s *SimpleQueryBuilder) GetDefaultSort() string {
	if s.DefaultSort == "" {
		return "id asc"
	}
	return s.DefaultSort
}

// NewSimpleQueryBuilder creates a new SimpleQueryBuilder with default settings
func NewSimpleQueryBuilder(tableName string) *SimpleQueryBuilder {
	return &SimpleQueryBuilder{
		TableName:   tableName,
		DefaultSort: "id asc",
		Dialect:     MySQL,
	}
}

// WithSearchFields sets the search fields for the query builder
func (s *SimpleQueryBuilder) WithSearchFields(fields ...string) *SimpleQueryBuilder {
	s.SearchFields = fields
	return s
}

// WithDefaultSort sets the default sort for the query builder
func (s *SimpleQueryBuilder) WithDefaultSort(sort string) *SimpleQueryBuilder {
	s.DefaultSort = sort
	return s
}

// WithDialect sets the database dialect for the query builder
func (s *SimpleQueryBuilder) WithDialect(dialect DatabaseDialect) *SimpleQueryBuilder {
	s.Dialect = dialect
	return s
}

// WithFilters sets the filter function for the query builder
func (s *SimpleQueryBuilder) WithFilters(filterFunc func(*gorm.DB) *gorm.DB) *SimpleQueryBuilder {
	s.FilterFunc = filterFunc
	return s
}

// GetSearchOperator returns the search operator based on the current dialect
func (s *SimpleQueryBuilder) GetSearchOperator() string {
	return getSearchOperator(s.Dialect)
}

// ChainableQueryBuilder allows for method chaining to build complex queries
type ChainableQueryBuilder struct {
	*SimpleQueryBuilder
	joins   []string
	groupBy []string
	having  []string
	selects []string
}

// NewChainableQueryBuilder creates a new ChainableQueryBuilder
func NewChainableQueryBuilder(tableName string) *ChainableQueryBuilder {
	return &ChainableQueryBuilder{
		SimpleQueryBuilder: NewSimpleQueryBuilder(tableName),
		joins:              make([]string, 0),
		groupBy:            make([]string, 0),
		having:             make([]string, 0),
		selects:            make([]string, 0),
	}
}

// Join adds a JOIN clause to the query
func (c *ChainableQueryBuilder) Join(join string) *ChainableQueryBuilder {
	c.joins = append(c.joins, join)
	return c
}

// GroupBy adds a GROUP BY clause to the query
func (c *ChainableQueryBuilder) GroupBy(field string) *ChainableQueryBuilder {
	c.groupBy = append(c.groupBy, field)
	return c
}

// Having adds a HAVING clause to the query
func (c *ChainableQueryBuilder) Having(condition string) *ChainableQueryBuilder {
	c.having = append(c.having, condition)
	return c
}

// Select adds a SELECT clause to the query
func (c *ChainableQueryBuilder) Select(fields ...string) *ChainableQueryBuilder {
	c.selects = append(c.selects, fields...)
	return c
}

// ApplyFilters applies all the configured filters including joins, group by, etc.
func (c *ChainableQueryBuilder) ApplyFilters(query *gorm.DB) *gorm.DB {
	// Apply base filters first
	query = c.SimpleQueryBuilder.ApplyFilters(query)

	// Apply selects
	if len(c.selects) > 0 {
		query = query.Select(c.selects)
	}

	// Apply joins
	for _, join := range c.joins {
		query = query.Joins(join)
	}

	// Apply group by
	for _, groupBy := range c.groupBy {
		query = query.Group(groupBy)
	}

	// Apply having
	for _, having := range c.having {
		query = query.Having(having)
	}

	return query
}
