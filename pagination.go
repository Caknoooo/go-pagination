package pagination

import (
	"math"
	"strconv"

	"github.com/gin-gonic/gin"
)

type PaginationRequest struct {
	Page    int    `json:"page" form:"page"`
	PerPage int    `json:"per_page" form:"per_page"`
	Search  string `json:"search" form:"search"`
	Sort    string `json:"sort" form:"sort"`
	Order   string `json:"order" form:"order"`
}

type PaginationResponse struct {
	Page    int   `json:"page"`
	PerPage int   `json:"per_page"`
	MaxPage int64 `json:"max_page"`
	Total   int64 `json:"total"`
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
		Page:    1,
		PerPage: 10,
		Search:  "",
		Sort:    "",
		Order:   "asc",
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

	pagination.Validate()
	return pagination
}

func CalculatePagination(pagination PaginationRequest, totalCount int64) PaginationResponse {
	maxPage := int64(math.Ceil(float64(totalCount) / float64(pagination.PerPage)))

	if maxPage == 0 {
		maxPage = 1
	}

	return PaginationResponse{
		Page:    pagination.Page,
		PerPage: pagination.PerPage,
		MaxPage: maxPage,
		Total:   totalCount,
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
