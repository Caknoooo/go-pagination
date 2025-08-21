package pagination

import (
	"reflect"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type BaseFilter struct {
	Pagination PaginationRequest `json:"pagination"`
	Includes   []string          `json:"includes"`
}

func (f *BaseFilter) BindPagination(ctx *gin.Context) {
	f.Pagination = BindPagination(ctx)

	// Bind includes from query parameter
	if includesStr := ctx.Query("includes"); includesStr != "" {
		f.Includes = strings.Split(includesStr, ",")
		// Clean whitespace from includes
		for i, include := range f.Includes {
			f.Includes[i] = strings.TrimSpace(include)
		}
	}
}

func (f *BaseFilter) GetOffset() int {
	return f.Pagination.GetOffset()
}

func (f *BaseFilter) GetLimit() int {
	return f.Pagination.GetLimit()
}

func (f *BaseFilter) ValidatePagination() {
	f.Pagination.Validate()
}

func (f *BaseFilter) GetPagination() PaginationRequest {
	return f.Pagination
}

func (f *BaseFilter) GetIncludes() []string {
	return f.Includes
}

type Filterable interface {
	ApplyFilters(query *gorm.DB) *gorm.DB
	GetTableName() string
	GetSearchFields() []string
	GetDefaultSort() string
	GetIncludes() []string
	GetPagination() PaginationRequest
}

// AdvancedQueryBuilder provides more sophisticated query building capabilities
type AdvancedQueryBuilder struct {
	SimpleQueryBuilder
	JoinClauses    []string
	GroupByClauses []string
	HavingClauses  []string
	SelectFields   []string
}

func (a *AdvancedQueryBuilder) ApplyJoins(query *gorm.DB) *gorm.DB {
	for _, join := range a.JoinClauses {
		query = query.Joins(join)
	}
	return query
}

func (a *AdvancedQueryBuilder) ApplyGroupBy(query *gorm.DB) *gorm.DB {
	for _, groupBy := range a.GroupByClauses {
		query = query.Group(groupBy)
	}
	return query
}

func (a *AdvancedQueryBuilder) ApplyHaving(query *gorm.DB) *gorm.DB {
	for _, having := range a.HavingClauses {
		query = query.Having(having)
	}
	return query
}

func (a *AdvancedQueryBuilder) ApplySelect(query *gorm.DB) *gorm.DB {
	if len(a.SelectFields) > 0 {
		query = query.Select(a.SelectFields)
	}
	return query
}

// FilterCondition represents a single filter condition
type FilterCondition struct {
	Field    string
	Operator string
	Value    interface{}
	Logic    string // AND, OR
}

// DynamicFilter allows for dynamic filtering based on struct tags
type DynamicFilter struct {
	BaseFilter
	Filters      []FilterCondition `json:"filters"`
	TableName    string            `json:"-"`
	Model        interface{}       `json:"-"`
	SearchFields []string          `json:"-"`
	DefaultSort  string            `json:"-"`
}

func (d *DynamicFilter) ApplyFilters(query *gorm.DB) *gorm.DB {
	for i, filter := range d.Filters {
		if filter.Field == "" || filter.Value == nil {
			continue
		}

		// Prevent SQL injection by validating field names
		if !d.isValidField(filter.Field) {
			continue
		}

		condition := d.buildCondition(filter)
		if condition == "" {
			continue
		}

		if i == 0 {
			query = query.Where(condition, filter.Value)
		} else {
			logic := strings.ToUpper(filter.Logic)
			if logic == "OR" {
				query = query.Or(condition, filter.Value)
			} else {
				query = query.Where(condition, filter.Value)
			}
		}
	}
	return query
}

func (d *DynamicFilter) isValidField(fieldName string) bool {
	if d.Model == nil {
		return false
	}

	modelType := reflect.TypeOf(d.Model)
	if modelType.Kind() == reflect.Ptr {
		modelType = modelType.Elem()
	}

	for i := 0; i < modelType.NumField(); i++ {
		field := modelType.Field(i)
		dbTag := field.Tag.Get("gorm")
		jsonTag := field.Tag.Get("json")

		// Check various field name formats
		if field.Name == fieldName ||
			strings.EqualFold(field.Name, fieldName) ||
			d.extractColumnName(dbTag) == fieldName ||
			d.extractJSONName(jsonTag) == fieldName {
			return true
		}
	}

	return false
}

func (d *DynamicFilter) extractColumnName(gormTag string) string {
	parts := strings.Split(gormTag, ";")
	for _, part := range parts {
		if strings.HasPrefix(part, "column:") {
			return strings.TrimPrefix(part, "column:")
		}
	}
	return ""
}

func (d *DynamicFilter) extractJSONName(jsonTag string) string {
	parts := strings.Split(jsonTag, ",")
	if len(parts) > 0 && parts[0] != "-" {
		return parts[0]
	}
	return ""
}

func (d *DynamicFilter) buildCondition(filter FilterCondition) string {
	switch strings.ToUpper(filter.Operator) {
	case "=", "EQ", "EQUALS":
		return filter.Field + " = ?"
	case "!=", "NE", "NOT_EQUALS":
		return filter.Field + " != ?"
	case ">", "GT", "GREATER_THAN":
		return filter.Field + " > ?"
	case ">=", "GTE", "GREATER_THAN_EQUALS":
		return filter.Field + " >= ?"
	case "<", "LT", "LESS_THAN":
		return filter.Field + " < ?"
	case "<=", "LTE", "LESS_THAN_EQUALS":
		return filter.Field + " <= ?"
	case "LIKE", "CONTAINS":
		return filter.Field + " LIKE ?"
	case "ILIKE", "ICONTAINS":
		return filter.Field + " ILIKE ?"
	case "IN":
		return filter.Field + " IN ?"
	case "NOT_IN":
		return filter.Field + " NOT IN ?"
	case "IS_NULL":
		return filter.Field + " IS NULL"
	case "IS_NOT_NULL":
		return filter.Field + " IS NOT NULL"
	default:
		return filter.Field + " = ?"
	}
}

func (d *DynamicFilter) GetTableName() string {
	return d.TableName
}

func (d *DynamicFilter) GetSearchFields() []string {
	return d.SearchFields
}

func (d *DynamicFilter) GetDefaultSort() string {
	if d.DefaultSort == "" {
		return "id asc"
	}
	return d.DefaultSort
}

func (d *DynamicFilter) ApplySearch(query *gorm.DB, searchTerm string) *gorm.DB {
	if len(d.SearchFields) == 0 || searchTerm == "" {
		return query
	}

	searchPattern := "%" + searchTerm + "%"

	// Build OR conditions for search
	conditions := make([]string, len(d.SearchFields))
	args := make([]interface{}, len(d.SearchFields))

	for i, field := range d.SearchFields {
		// Use LIKE for better database compatibility
		conditions[i] = field + " LIKE ?"
		args[i] = searchPattern
	}

	whereClause := "(" + strings.Join(conditions, " OR ") + ")"
	return query.Where(whereClause, args...)
}
