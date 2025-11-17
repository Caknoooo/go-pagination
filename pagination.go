package pagination

import (
	"math"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

type PaginationRequest struct {
	Page       int    `json:"page" form:"page"`
	PerPage    int    `json:"per_page" form:"per_page"`
	Search     string `json:"search" form:"search"`
	Sort       string `json:"sort" form:"sort"`
	Order      string `json:"order" form:"order"`
	IsDisabled bool   `json:"is_disabled,omitempty" form:"is_disabled"`
}

type PaginationResponse struct {
	Page       int   `json:"page"`
	PerPage    int   `json:"per_page"`
	MaxPage    int64 `json:"max_page"`
	Total      int64 `json:"total"`
	IsDisabled bool  `json:"is_disabled,omitempty"`
}

type PaginatedResponse struct {
	Code       int                `json:"code"`
	Status     string             `json:"status"`
	Message    string             `json:"message"`
	Data       interface{}        `json:"data"`
	Pagination PaginationResponse `json:"pagination"`
}

func (p *PaginationRequest) GetOffset() int {
	if p.Page <= 0 {
		p.Page = 1
	}
	return (p.Page - 1) * p.PerPage
}

func (p *PaginationRequest) GetLimit() int {
	if p.PerPage <= 0 {
		p.PerPage = 10
	}
	return p.PerPage
}

func (p *PaginationRequest) Validate() {
	if p.Page <= 0 {
		p.Page = 1
	}

	if p.PerPage <= 0 {
		p.PerPage = 10
	}

	if p.Order == "" {
		p.Order = "asc"
	}

	if p.Order != "asc" && p.Order != "desc" {
		p.Order = "asc"
	}
}

func BindPagination(ctx *gin.Context) PaginationRequest {
	pagination := PaginationRequest{
		Page:       1,
		PerPage:    10,
		Search:     "",
		Sort:       "",
		Order:      "asc",
		IsDisabled: false,
	}

	if pageStr := ctx.Query("page"); pageStr != "" {
		if page, err := strconv.Atoi(pageStr); err == nil && page > 0 {
			pagination.Page = page
		}
	}

	if perPageStr := ctx.Query("per_page"); perPageStr != "" {
		if perPage, err := strconv.Atoi(perPageStr); err == nil && perPage > 0 && perPage <= 100 {
			pagination.PerPage = perPage
		}
	}

	pagination.Search = ctx.Query("search")

	pagination.Sort = ctx.Query("sort")

	if order := ctx.Query("order"); order == "desc" || order == "asc" {
		pagination.Order = order
	}

	if isDisabled := ctx.Query("is_disabled"); isDisabled != "" {
		switch strings.ToLower(isDisabled) {
		case "1", "true", "yes", "y", "on":
			pagination.IsDisabled = true
		default:
			pagination.IsDisabled = false
		}
	}

	pagination.Validate()
	return pagination
}

func CalculatePagination(pagination PaginationRequest, totalCount int64) PaginationResponse {
	// When pagination disabled, return minimal metadata
	if pagination.IsDisabled {
		return PaginationResponse{
			Page:       1,
			PerPage:    int(totalCount),
			MaxPage:    1,
			Total:      totalCount,
			IsDisabled: true,
		}
	}

	maxPage := int64(math.Ceil(float64(totalCount) / float64(pagination.PerPage)))

	if maxPage == 0 {
		maxPage = 1
	}

	return PaginationResponse{
		Page:       pagination.Page,
		PerPage:    pagination.PerPage,
		MaxPage:    maxPage,
		Total:      totalCount,
		IsDisabled: false,
	}
}

func NewPaginatedResponse(code int, message string, data interface{}, pagination PaginationResponse) PaginatedResponse {
	status := "success"
	if code >= 400 {
		status = "error"
	}

	return PaginatedResponse{
		Code:       code,
		Status:     status,
		Message:    message,
		Data:       data,
		Pagination: pagination,
	}
}
