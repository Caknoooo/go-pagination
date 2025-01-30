package pagination

import (
	"strconv"

	"github.com/gin-gonic/gin"
)

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
	Total       int  `json:"total"`
	LastPage    int  `json:"last_page"`
}

type PaginationLinks struct {
	First string  `json:"first"`
	Last  string  `json:"last"`
	Next  *string `json:"next"`
	Prev  *string `json:"prev"`
}

type InitPaginationResponse struct {
	Size   int  `json:"size"`
	Number int  `json:"number"`
	Offset int  `json:"offset"`
	From   *int `json:"from"`
	To     *int `json:"to"`
}

const (
	DEFAULT_PAGE_SIZE   = 10
	DEFAULT_PAGE_NUMBER = 1

	DEFAULT_PAGE_SIZE_QUERY   = "page[size]"
	DEFAULT_PAGE_NUMBER_QUERY = "page[number]"
)

func InitPagination(req PaginationRequest, lengthModel int) InitPaginationResponse {
	if req.Size <= 0 {
		req.Size = DEFAULT_PAGE_SIZE
	}

	if req.Number <= 0 {
		req.Number = DEFAULT_PAGE_NUMBER
	}

	offset := (req.Number - 1) * req.Size
	from := offset + 1
	to := offset + req.Size

	if from > lengthModel {
		from = lengthModel
	}

	if to > lengthModel {
		to = lengthModel
	}

	return InitPaginationResponse{
		Size:   req.Size,
		Number: req.Number,
		Offset: offset,
		From:   &from,
		To:     &to,
	}
}

func GeneratePaginationLinks(c *gin.Context, req PaginationRequest, lengthModel int) (PaginationResponse, error) {
	request := InitPagination(req, lengthModel)
	lastPage := (lengthModel + request.Size - 1) / request.Size

	queryParams := c.Request.URL.Query()
	queryParams.Set("page[size]", strconv.Itoa(request.Size))

	baseURL := c.Request.URL.Scheme + "://" + c.Request.Host + c.Request.URL.Path

	queryParams.Set("page[number]", strconv.Itoa(1))
	first := baseURL + "?" + queryParams.Encode()

	queryParams.Set("page[number]", strconv.Itoa(lastPage))
	last := baseURL + "?" + queryParams.Encode()

	var next *string
	var prev *string

	if request.Number > 1 {
		queryParams.Set("page[number]", strconv.Itoa(request.Number-1))
		prevStr := baseURL + "?" + queryParams.Encode()
		prev = &prevStr
	}

	if request.Number < lastPage {
		queryParams.Set("page[number]", strconv.Itoa(request.Number+1))
		nextStr := baseURL + "?" + queryParams.Encode()
		next = &nextStr
	}

	meta := MetaResponse{
		CurrentPage: request.Number,
		PerPage:     request.Size,
		From:        request.From,
		To:          request.To,
		Total:       lengthModel,
		LastPage:    lastPage,
	}

	return PaginationResponse{
		Meta: meta,
		Links: PaginationLinks{
			First: first,
			Last:  last,
			Next:  next,
			Prev:  prev,
		},
	}, nil
}
