package main

import (
	"github.com/Caknoooo/go-pagination"
	"gorm.io/gorm"
)

// Province model
type Province struct {
	ID       uint      `json:"id" gorm:"primaryKey"`
	Name     string    `json:"name" gorm:"column:name"`
	Code     string    `json:"code" gorm:"column:code"`
	Athletes []Athlete `json:"athletes,omitempty" gorm:"foreignKey:ProvinceID"`
}

// ProvinceFilter - Custom filter untuk Province
type ProvinceFilter struct {
	pagination.BaseFilter
	ID   int    `json:"id" form:"id"`
	Name string `json:"name" form:"name"`
	Code string `json:"code" form:"code"`
}

func (f *ProvinceFilter) ApplyFilters(query *gorm.DB) *gorm.DB {
	if f.ID > 0 {
		query = query.Where("id = ?", f.ID)
	}
	if f.Name != "" {
		query = query.Where("name LIKE ?", "%"+f.Name+"%")
	}
	if f.Code != "" {
		query = query.Where("code = ?", f.Code)
	}

	return query
}

func (f *ProvinceFilter) GetTableName() string {
	return "provinces"
}

func (f *ProvinceFilter) GetSearchFields() []string {
	return []string{"name", "code"}
}

func (f *ProvinceFilter) GetDefaultSort() string {
	return "id asc"
}

func (f *ProvinceFilter) GetIncludes() []string {
	return f.Includes
}

func (f *ProvinceFilter) GetPagination() pagination.PaginationRequest {
	return f.Pagination
}

func (f *ProvinceFilter) Validate() {
	var validIncludes []string
	allowedIncludes := f.GetAllowedIncludes()
	for _, include := range f.Includes {
		if allowedIncludes[include] {
			validIncludes = append(validIncludes, include)
		}
	}
	f.Includes = validIncludes
}

func (f *ProvinceFilter) GetAllowedIncludes() map[string]bool {
	return map[string]bool{
		// Province memiliki relasi dengan Athletes
		"Athletes": true,
	}
}
