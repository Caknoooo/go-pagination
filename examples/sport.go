package main

import (
	"github.com/Caknoooo/go-pagination"
	"gorm.io/gorm"
)

// Sport model
type Sport struct {
	ID          uint       `json:"id" gorm:"primaryKey"`
	Name        string     `json:"name" gorm:"column:name"`
	Category    string     `json:"category" gorm:"column:category"`
	Description string     `json:"description" gorm:"column:description"`
	IsActive    bool       `json:"is_active" gorm:"column:is_active;default:true"`
	Athletes    []Athlete  `json:"athletes,omitempty" gorm:"foreignKey:SportID"`
	Events      []Event    `json:"events,omitempty" gorm:"foreignKey:SportID"`
}

// SportFilter - Custom filter untuk Sport
type SportFilter struct {
	pagination.BaseFilter
	ID       int    `json:"id" form:"id"`
	Name     string `json:"name" form:"name"`
	Category string `json:"category" form:"category"`
	IsActive bool   `json:"is_active" form:"is_active"`
}

func (f *SportFilter) ApplyFilters(query *gorm.DB) *gorm.DB {
	if f.ID > 0 {
		query = query.Where("id = ?", f.ID)
	}
	if f.Name != "" {
		query = query.Where("name LIKE ?", "%"+f.Name+"%")
	}
	if f.Category != "" {
		query = query.Where("category = ?", f.Category)
	}

	return query
}

func (f *SportFilter) GetTableName() string {
	return "sports"
}

func (f *SportFilter) GetSearchFields() []string {
	return []string{"name", "category", "description"}
}

func (f *SportFilter) GetDefaultSort() string {
	return "id asc"
}

func (f *SportFilter) GetIncludes() []string {
	return f.Includes
}

func (f *SportFilter) GetPagination() pagination.PaginationRequest {
	return f.Pagination
}

func (f *SportFilter) Validate() {
	var validIncludes []string
	allowedIncludes := f.GetAllowedIncludes()
	for _, include := range f.Includes {
		if allowedIncludes[include] {
			validIncludes = append(validIncludes, include)
		}
	}
	f.Includes = validIncludes
}

func (f *SportFilter) GetAllowedIncludes() map[string]bool {
	return map[string]bool{
		// Sport memiliki relasi dengan Athletes dan Events
		"Athletes": true,
		"Events":   true,
	}
}
