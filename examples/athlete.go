package main

import (
	"github.com/Caknoooo/go-pagination"
	"gorm.io/gorm"
)

type AthleteQuery struct {
	db *gorm.DB
}

func NewAthleteQuery(db *gorm.DB) *AthleteQuery {
	return &AthleteQuery{
		db: db,
	}
}

type Athlete struct {
	ID            int             `json:"id"`
	ProvinceID    int             `json:"province_id"`
	Province      *Province       `json:"province,omitempty"`
	SportID       int             `json:"sport_id"`
	Sport         *Sport          `json:"sport,omitempty"`
	Name          string          `json:"name"`
	Age           int             `json:"age"`
	Gender        string          `json:"gender"`
	BirthDate     string          `json:"birthdate"`
	Height        int             `json:"height"`
	Image         string          `json:"image"`
	PlayersEvents []PlayersEvents `json:"players_events,omitempty" gorm:"polymorphic:Player;polymorphicValue:athlete"`
}

type PlayersEvents struct {
	ID         int    `json:"id"`
	PlayerID   int    `json:"player_id"`
	PlayerType string `json:"player_type"`
	EventID    int    `json:"event_id"`
}

type AthleteFilter struct {
	pagination.BaseFilter
	ID         int `json:"id" form:"id"`
	ProvinceID int `json:"province_id" form:"province_id"`
	SportID    int `json:"sport_id" form:"sport_id"`
	EventID    int `json:"event_id" form:"event_id"`
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
		// You can add joins or subqueries here for EventID filtering
		// Example: query = query.Joins("JOIN players_events pe ON pe.player_id = athletes.id AND pe.player_type = 'athlete'").Where("pe.event_id = ?", f.EventID)
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

func (f *AthleteFilter) GetIncludes() []string {
	return f.Includes
}

func (f *AthleteFilter) GetPagination() pagination.PaginationRequest {
	return f.Pagination
}

func (f *AthleteFilter) Validate() {
	var validIncludes []string
	allowedIncludes := f.GetAllowedIncludes()
	for _, include := range f.Includes {
		if allowedIncludes[include] {
			validIncludes = append(validIncludes, include)
		}
	}
	f.Includes = validIncludes
}

func (f *AthleteFilter) GetAllowedIncludes() map[string]bool {
	return map[string]bool{
		"Province":      true,
		"Sport":         true,
		"PlayersEvents": true,
	}
}
