package changeset

import (
	"context"
	"fmt"
	"net/http"

	"github.com/egfanboy/mediapire-common/router"
	"github.com/egfanboy/mediapire-manager/internal/app"
	"github.com/egfanboy/mediapire-manager/pkg/types"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

const (
	basePath         = "/changesets"
	paramChangesetId = "changesetId"
)

type changesetController struct {
	builders []func() router.RouteBuilder
	service  ChangesetApi
}

func (c changesetController) GetApis() (routes []router.RouteBuilder) {
	for _, b := range c.builders {
		routes = append(routes, b())
	}

	return
}

func (c changesetController) GetChangesets() router.RouteBuilder {
	return router.NewV1RouteBuilder().
		SetMethod(http.MethodOptions, http.MethodGet).
		SetPath(basePath).
		SetReturnCode(http.StatusOK).
		SetHandler(func(request *http.Request, p router.RouteParams) (interface{}, error) {
			r, err := c.service.GetChangesets(request.Context())
			if err != nil {
				return nil, err
			}

			result := make([]types.ChangesetItem, len(r))
			for i, v := range r {
				result[i] = v.ToApiResponse()
			}

			return result, nil
		})
}

func (c changesetController) GetChangesetId() router.RouteBuilder {
	return router.NewV1RouteBuilder().
		SetMethod(http.MethodOptions, http.MethodGet).
		SetPath(fmt.Sprintf("%s/{%s}", basePath, paramChangesetId)).
		SetReturnCode(http.StatusOK).
		SetHandler(func(request *http.Request, p router.RouteParams) (interface{}, error) {
			changesetId, ok := p.Params[paramChangesetId]
			if !ok {
				return nil, fmt.Errorf("%s not found in API path", paramChangesetId)
			}

			changesetObjectId, err := primitive.ObjectIDFromHex(changesetId)
			if err != nil {
				return nil, err
			}

			r, err := c.service.GetChangesetById(request.Context(), changesetObjectId)
			if err != nil {
				return nil, err
			}

			return r.ToApiResponse(), nil
		})

}

func (c changesetController) CreateChangeset() router.RouteBuilder {
	return router.NewV1RouteBuilder().
		SetMethod(http.MethodOptions, http.MethodPost).
		SetPath(basePath).
		SetReturnCode(http.StatusAccepted).
		SetHandler(func(request *http.Request, p router.RouteParams) (interface{}, error) {
			var body types.ChangesetCreateRequest
			err := p.PopulateBody(&body)
			if err != nil {
				return nil, err
			}

			r, err := c.service.CreateChangeset(request.Context(), body)
			if err != nil {
				return nil, err
			}

			return r.ToApiResponse(), nil
		})
}

func initController() changesetController {
	// TODO: Need to rethink this to handle errors
	service, _ := newChangesetService(context.Background())

	c := changesetController{service: service}

	c.builders = append(
		c.builders,
		c.GetChangesets,
		c.GetChangesetId,
		c.CreateChangeset,
	)

	return c
}

func init() {
	app.GetApp().ControllerRegistry.Register(initController())
}
