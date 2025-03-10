package types

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/egfanboy/mediapire-common/exceptions"
	"github.com/egfanboy/mediapire-common/router"
)

const (
	sortByQueryParamName = "sortBy"
)

var (
	QueryParamSortBy = router.QueryParam{Name: sortByQueryParamName, Required: false}
	sortRegEx        = regexp.MustCompile(`\b(?:asc|desc)\(([^)]+)\)`)
)

type ApiFilteringParams struct {
	SortByField *string
	SortByOrder *string
}

func NewApiFilteringParams(p router.RouteParams) (ApiFilteringParams, error) {
	f := ApiFilteringParams{}

	if sortBy, ok := p.Params[sortByQueryParamName]; ok {
		if match := sortRegEx.FindStringSubmatch(sortBy); match == nil {
			return ApiFilteringParams{}, exceptions.NewBadRequestException(errors.New("sortBy query param does not match expected format"))
		} else {
			field := match[1]
			order := strings.ReplaceAll(match[0], fmt.Sprintf("(%s)", field), "")
			f.SortByField = &field
			f.SortByOrder = &order
		}
	}

	return f, nil
}
