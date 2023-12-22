package media

import (
	"net/http"

	"github.com/egfanboy/mediapire-manager/internal/app"
	"github.com/egfanboy/mediapire-manager/pkg/types"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"

	"github.com/egfanboy/mediapire-common/router"
)

const (
	basePath          = "/media"
	queryParamMediaId = "mediaId"
	queryParamNodeId  = "nodeId"
)

type mediaController struct {
	builders []func() router.RouteBuilder
	service  mediaApi
}

func (c mediaController) GetApis() (routes []router.RouteBuilder) {
	for _, b := range c.builders {

		routes = append(routes, b())
	}

	return
}

func (c mediaController) handleGetAll() router.RouteBuilder {
	return router.NewV1RouteBuilder().
		SetMethod(http.MethodOptions, http.MethodGet).
		SetPath(basePath).
		SetReturnCode(http.StatusOK).
		SetHandler(func(request *http.Request, p router.RouteParams) (interface{}, error) {

			return c.service.GetMedia(request.Context())
		})
}

func (c mediaController) StreamMedia() router.RouteBuilder {
	return router.NewV1RouteBuilder().
		SetMethod(http.MethodOptions, http.MethodGet).
		SetPath(basePath + "/stream").
		SetDataType(router.DataTypeFile).
		SetReturnCode(http.StatusOK).
		AddQueryParam(router.QueryParam{Name: queryParamMediaId, Required: true}).
		AddQueryParam(router.QueryParam{Name: queryParamNodeId, Required: true}).
		SetHandler(func(request *http.Request, p router.RouteParams) (interface{}, error) {
			return c.service.StreamMedia(request.Context(), uuid.MustParse(p.Params[queryParamNodeId]), uuid.MustParse(p.Params[queryParamMediaId]))
		})
}

func (c mediaController) DownloadMedia() router.RouteBuilder {
	return router.NewV1RouteBuilder().
		SetMethod(http.MethodOptions, http.MethodPost).
		SetPath(basePath + "/download").
		SetReturnCode(http.StatusOK).
		SetHandler(func(httpReq *http.Request, p router.RouteParams) (interface{}, error) {
			var request types.MediaDownloadRequest
			err := p.PopulateBody(&request)
			if err != nil {
				return nil, err
			}

			return c.service.DownloadMediaAsync(httpReq.Context(), request)
		})
}

func initController() (mediaController, error) {
	mediaService, err := newMediaService()

	if err != nil {
		return mediaController{}, err
	}

	c := mediaController{service: mediaService}

	c.builders = append(c.builders, c.handleGetAll, c.StreamMedia, c.DownloadMedia)

	return c, nil
}

func init() {
	controller, err := initController()

	if err != nil {
		log.Error().Err(err).Msg("Failed to instantiate media controller")
	} else {
		app.GetApp().ControllerRegistry.Register(controller)
	}
}
