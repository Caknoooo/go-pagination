package main

import (
	"time"

	"github.com/Caknoooo/go-pagination"
	"gorm.io/gorm"
)

// Event model
type Event struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	Name        string    `json:"name" gorm:"column:name"`
	Description string    `json:"description" gorm:"column:description"`
	StartDate   time.Time `json:"start_date" gorm:"column:start_date"`
	EndDate     time.Time `json:"end_date" gorm:"column:end_date"`
	Location    string    `json:"location" gorm:"column:location"`
	IsActive    bool      `json:"is_active" gorm:"column:is_active;default:true"`
}

// EventFilter - Custom filter untuk Event
type EventFilter struct {
	pagination.BaseFilter
	ID        int       `json:"id" form:"id"`
	Name      string    `json:"name" form:"name"`
	Location  string    `json:"location" form:"location"`
	IsActive  bool      `json:"is_active" form:"is_active"`
	Year      int       `json:"year" form:"year"`
	SportID   int       `json:"sport_id" form:"sport_id"`
	StartDate time.Time `json:"start_date" form:"start_date"`
	EndDate   time.Time `json:"end_date" form:"end_date"`
}

func (f *EventFilter) ApplyFilters(query *gorm.DB) *gorm.DB {
	if f.ID > 0 {
		query = query.Where("id = ?", f.ID)
	}
	if f.Name != "" {
		query = query.Where("name LIKE ?", "%"+f.Name+"%")
	}
	if f.Year > 0 {
		query = query.Where("YEAR(start_date) = ?", f.Year)
	}
	if f.SportID > 0 {
		query = query.Where("sport_id = ?", f.SportID)
	}
	if !f.StartDate.IsZero() {
		query = query.Where("start_date >= ?", f.StartDate)
	}
	if !f.EndDate.IsZero() {
		query = query.Where("end_date <= ?", f.EndDate)
	}

	return query
}

func (f *EventFilter) GetTableName() string {
	return "events"
}

func (f *EventFilter) GetSearchFields() []string {
	return []string{"name", "description", "location"}
}

func (f *EventFilter) GetDefaultSort() string {
	return "id asc"
}
