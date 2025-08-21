package main

import (
	"github.com/Caknoooo/go-pagination"
	"gorm.io/gorm"
)

// Province model
type Province struct {
	ID   uint   `json:"id" gorm:"primaryKey"`
	Name string `json:"name" gorm:"column:name"`
	Code string `json:"code" gorm:"column:code"`
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
