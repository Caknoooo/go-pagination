package main

import (
	"github.com/Caknoooo/go-pagination"
	"gorm.io/gorm"
)

// Athlete model
type Athlete struct {
	ID         uint   `json:"id" gorm:"primaryKey"`
	Name       string `json:"name" gorm:"column:name"`
	ProvinceID uint   `json:"province_id" gorm:"column:province_id"`
	SportID    uint   `json:"sport_id" gorm:"column:sport_id"`
	EventID    uint   `json:"event_id" gorm:"column:event_id"`
	Age        int    `json:"age" gorm:"column:age"`
	IsActive   bool   `json:"is_active" gorm:"column:is_active;default:true"`
}

// AthleteFilter - Custom filter untuk Athlete
type AthleteFilter struct {
	pagination.BaseFilter
	ID         int  `json:"id" form:"id"`
	ProvinceID int  `json:"province_id" form:"province_id"`
	SportID    int  `json:"sport_id" form:"sport_id"`
	EventID    int  `json:"event_id" form:"event_id"`
	MinAge     int  `json:"min_age" form:"min_age"`
	MaxAge     int  `json:"max_age" form:"max_age"`
	IsActive   bool `json:"is_active" form:"is_active"`
}

func (f *AthleteFilter) ApplyFilters(query *gorm.DB) *gorm.DB {
	if f.ID > 0 {
		query = query.Where("id = ?", f.ID)
	}
	if f.ProvinceID > 0 {
		query = query.Where("province_id = ?", f.ProvinceID)
	}
	if f.SportID > 0 {
		query = query.Where("sport_id = ?", f.SportID)
	}
	if f.EventID > 0 {
		query = query.Where("event_id = ?", f.EventID)
	}
	if f.MinAge > 0 {
		query = query.Where("age >= ?", f.MinAge)
	}
	if f.MaxAge > 0 {
		query = query.Where("age <= ?", f.MaxAge)
	}

	return query
}

func (f *AthleteFilter) GetTableName() string {
	return "athletes"
}

func (f *AthleteFilter) GetSearchFields() []string {
	return []string{"name"}
}

func (f *AthleteFilter) GetDefaultSort() string {
	return "id asc"
}
