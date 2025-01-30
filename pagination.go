package pagination

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

const (
	DefaultPageSize   = 10
	DefaultPageNumber = 1
	PageSizeQuery     = "page[size]"
	PageNumberQuery   = "page[number]"
)

type Pagination struct {
	DB         *gorm.DB
	Req        PaginationRequest
	TotalItems int64
}

type PaginationRequest struct {
	Size   int `form:"page[size]"`
	Number int `form:"page[number]"`
}

type PaginationResponse struct {
	Meta  MetaResponse    `json:"meta"`
	Links PaginationLinks `json:"links"`
}

type MetaResponse struct {
	CurrentPage int  `json:"current_page"`
	PerPage     int  `json:"per_page"`
	From        *int `json:"from"`
	To          *int `json:"to"`
}

type PaginationLinks struct {
	First string  `json:"first"`
	Last  string  `json:"last"`
	Next  *string `json:"next"`
	Prev  *string `json:"prev"`
}

func New(db *gorm.DB, c *gin.Context) (*Pagination, error) {
	var req PaginationRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		return nil, err
	}

	if req.Size <= 0 {
		req.Size = DefaultPageSize
	}
	if req.Number <= 0 {
		req.Number = DefaultPageNumber
	}

	return &Pagination{
		DB:  db,
		Req: req,
	}, nil
}

func (p *Pagination) Query() *gorm.DB {
	offset := (p.Req.Number - 1) * p.Req.Size
	return p.DB.Offset(offset).Limit(p.Req.Size)
}

func (p *Pagination) Count(model interface{}) error {
	return p.DB.Model(model).Count(&p.TotalItems).Error
}

func (p *Pagination) GenerateResponse(c *gin.Context) PaginationResponse {
	baseURL := SetBaseURL(c)
	queryParams := c.Request.URL.Query()

	queryParams.Del(PageSizeQuery)
	queryParams.Del(PageNumberQuery)

	offset := (p.Req.Number - 1) * p.Req.Size
	from := offset + 1
	to := offset + p.Req.Size
	if to > int(p.TotalItems) {
		to = int(p.TotalItems)
	}

	queryParams.Set(PageSizeQuery, strconv.Itoa(p.Req.Size))
	queryParams.Set(PageNumberQuery, strconv.Itoa(DefaultPageNumber))
	first := baseURL + "?" + queryParams.Encode()

	lastPage := (int(p.TotalItems) + p.Req.Size - 1) / p.Req.Size
	queryParams.Set(PageNumberQuery, strconv.Itoa(lastPage))
	last := baseURL + "?" + queryParams.Encode()

	var next, prev *string
	if p.Req.Number > 1 {
		queryParams.Set(PageNumberQuery, strconv.Itoa(p.Req.Number-1))
		prevStr := baseURL + "?" + queryParams.Encode()
		prev = &prevStr
	}
	if p.Req.Number < lastPage {
		queryParams.Set(PageNumberQuery, strconv.Itoa(p.Req.Number+1))
		nextStr := baseURL + "?" + queryParams.Encode()
		next = &nextStr
	}

	return PaginationResponse{
		Meta: MetaResponse{
			CurrentPage: p.Req.Number,
			PerPage:     p.Req.Size,
			From:        &from,
			To:          &to,
		},
		Links: PaginationLinks{
			First: first,
			Last:  last,
			Next:  next,
			Prev:  prev,
		},
	}
}

func SetBaseURL(c *gin.Context) string {
	scheme := "http"
	if c.Request.TLS != nil {
		scheme = "https"
	}

	baseURL := scheme + "://" + c.Request.Host + c.Request.URL.Path

	return baseURL
}
