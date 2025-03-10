package pagination

import (
	"fmt"
	"math"

	"github.com/egfanboy/mediapire-common/exceptions"
)

type Pagination struct {
	CurrentPage  int  `json:"currentPage"`
	NextPage     *int `json:"nextPage"`
	PreviousPage *int `json:"previousPage"`
}

type PaginatedResponse[T any] struct {
	Results    []T        `json:"results"`
	Pagination Pagination `json:"pagination"`
}

func NewPaginatedResponse[T any](data []T, pagination ApiPaginationParams) (result PaginatedResponse[T], err error) {
	startIndex := (pagination.Page - 1) * pagination.Limit

	if len(data)-1 < startIndex {
		err = exceptions.NewBadRequestException(fmt.Errorf("no page %d for current data", pagination.Page))
		return
	}

	expectedEndIndex := startIndex + pagination.Limit
	endIndex := int(math.Min(float64(expectedEndIndex), float64(len(data)-1)))

	var paginatedData []T
	// only 1 record
	if startIndex == endIndex {
		paginatedData = data
	} else {

		paginatedData = data[startIndex:endIndex]
	}

	p := Pagination{
		CurrentPage: pagination.Page,
	}

	if pagination.Page == 1 {
		p.PreviousPage = nil
	} else {
		previous := pagination.Page - 1
		p.PreviousPage = &previous
	}

	if expectedEndIndex != endIndex {
		p.NextPage = nil
	} else {
		next := pagination.Page + 1
		p.NextPage = &next
	}

	return PaginatedResponse[T]{
		Results:    paginatedData,
		Pagination: p,
	}, nil
}
