package types

import (
	"errors"
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

func NewPaginatedResponse[T any](data []T, page int, limit int) (result PaginatedResponse[T], err error) {
	if page < 1 {
		err = exceptions.NewBadRequestException(errors.New("invalid page parameter. Must be greater than 1"))
		return
	}

	if limit > 100 {
		err = exceptions.NewBadRequestException(errors.New("invalid page limit. Must be 100 or smaller"))
		return
	}

	startIndex := (page - 1) * limit

	if len(data)-1 < startIndex {
		err = exceptions.NewBadRequestException(fmt.Errorf("no page %d for current data", page))
		return
	}

	expectedEndIndex := startIndex + limit
	endIndex := int(math.Min(float64(expectedEndIndex), float64(len(data)-1)))

	var paginatedData []T
	// only 1 record
	if startIndex == endIndex {
		paginatedData = data
	} else {

		paginatedData = data[startIndex:endIndex]
	}

	p := Pagination{
		CurrentPage: page,
	}

	if page == 1 {
		p.PreviousPage = nil
	} else {
		previous := page - 1
		p.PreviousPage = &previous
	}

	if expectedEndIndex != endIndex {
		p.NextPage = nil
	} else {
		next := page + 1
		p.NextPage = &next
	}

	return PaginatedResponse[T]{
		Results:    paginatedData,
		Pagination: p,
	}, nil
}
