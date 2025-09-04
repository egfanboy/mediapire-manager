package pagination

import (
	"errors"
	"strconv"

	"github.com/egfanboy/mediapire-common/exceptions"
	"github.com/egfanboy/mediapire-common/router"
)

const (
	pageQueryParamName  = "page"
	limitQueryParamName = "limit"

	defaultLimit = 10
)

var (
	PageQueryParam  = router.QueryParam{Name: pageQueryParamName, Required: false}
	LimitQueryParam = router.QueryParam{Name: limitQueryParamName, Required: false}
)

type ApiPaginationParams struct {
	Page  int
	Limit int
}

func (p ApiPaginationParams) Validate() error {
	if p.Page < 1 {
		return exceptions.NewBadRequestException(errors.New("invalid page parameter. Must be greater than 1"))
	}

	if p.Limit > 100 {
		return exceptions.NewBadRequestException(errors.New("invalid page limit. Must be 100 or smaller"))
	}

	return nil
}

func NewApiPaginationParams(p router.RouteParams) (result ApiPaginationParams, err error) {
	if page, ok := p.Params[pageQueryParamName]; !ok {
		err = exceptions.NewBadRequestException(errors.New("must provide page parameter"))
		return
	} else {
		pageInt, errConv := strconv.Atoi(page)
		if errConv != nil {
			err = exceptions.NewBadRequestException(errConv)
			return
		}
		result.Page = pageInt
	}

	if limit, ok := p.Params[limitQueryParamName]; !ok {
		result.Limit = defaultLimit
	} else {
		limitInt, errConv := strconv.Atoi(limit)
		if errConv != nil {
			err = exceptions.NewBadRequestException(errConv)
			return
		}

		result.Limit = limitInt
	}

	err = result.Validate()

	return
}
