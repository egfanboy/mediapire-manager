package media

import (
	"net/http"
	"strings"

	"github.com/egfanboy/mediapire-manager/internal/app"
	"github.com/egfanboy/mediapire-manager/pkg/types"
	"github.com/egfanboy/mediapire-manager/pkg/types/pagination"
	"github.com/rs/zerolog/log"

	"github.com/egfanboy/mediapire-common/router"
)

const (
	basePath            = "/media"
	queryParamMediaId   = "mediaId"
	queryParamMediaIds  = "mediaIds"
	queryParamNodeId    = "nodeId"
	queryParamMediaType = "mediaType"

	queryParamSortBy = "sortBy"
)

type mediaController struct {
	builders []func() router.RouteBuilder
	service  MediaApi
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
		AddQueryParam(router.QueryParam{Name: queryParamMediaType, Required: false}).
		AddQueryParam(router.QueryParam{Name: queryParamNodeId, Required: false}).
		AddQueryParam(pagination.PageQueryParam).
		AddQueryParam(pagination.LimitQueryParam).
		AddQueryParam(types.QueryParamSortBy).
		AddQueryParam(router.QueryParam{Name: queryParamSortBy, Required: false}).
		SetHandler(func(request *http.Request, p router.RouteParams) (interface{}, error) {
			nodeIds := make([]string, 0)
			mediaTypes := make([]string, 0)

			if mediaTypeQuery, ok := p.Params[queryParamMediaType]; ok {
				mediaTypes = append(mediaTypes, strings.Split(mediaTypeQuery, ",")...)
			}

			if nodeIdQuery, ok := p.Params[queryParamNodeId]; ok {
				nodeIds = append(nodeIds, strings.Split(nodeIdQuery, ",")...)
			}

			paginationParams, err := pagination.NewApiPaginationParams(p)
			if err != nil {
				return nil, err
			}

			filteringParams, err := types.NewApiFilteringParams(p)
			if err != nil {
				return nil, err
			}

			return c.service.GetMediaPaginated(request.Context(), mediaTypes, nodeIds, filteringParams, paginationParams)
		})
}

// create a second route that will only match /media?mediaIds=1,2 to have it ignore pagination
func (c mediaController) getAllById() router.RouteBuilder {
	return router.NewV1RouteBuilder().
		SetMethod(http.MethodOptions, http.MethodGet).
		SetPath(basePath).
		SetReturnCode(http.StatusOK).
		AddQueryParam(router.QueryParam{Name: queryParamMediaIds, Required: true}).
		SetHandler(func(request *http.Request, p router.RouteParams) (interface{}, error) {
			mediaIds := make([]string, 0)

			if mediaIdsQuery, ok := p.Params[queryParamMediaIds]; ok {
				mediaIds = append(mediaIds, strings.Split(mediaIdsQuery, ",")...)
			}

			return c.service.GetMedia(request.Context(), []string{}, []string{}, mediaIds)
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
			return c.service.StreamMedia(request.Context(), p.Params[queryParamNodeId], p.Params[queryParamMediaId])
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

func (c mediaController) DeleteMedia() router.RouteBuilder {
	return router.NewV1RouteBuilder().
		SetMethod(http.MethodOptions, http.MethodDelete).
		SetPath(basePath).
		SetReturnCode(http.StatusAccepted).
		SetHandler(func(httpReq *http.Request, p router.RouteParams) (interface{}, error) {
			var request types.MediaDeleteRequest
			err := p.PopulateBody(&request)
			if err != nil {
				return nil, err
			}

			err = c.service.DeleteMedia(httpReq.Context(), request)
			return nil, err
		})
}

func (c mediaController) handleGetArt() router.RouteBuilder {
	return router.NewV1RouteBuilder().
		SetMethod(http.MethodOptions, http.MethodGet).
		SetPath(basePath + "/{mediaId}/art").
		SetReturnCode(http.StatusOK).
		SetDataType(router.DataTypeFile).
		AddQueryParam(router.QueryParam{Name: queryParamNodeId, Required: true}).
		SetHandler(func(request *http.Request, p router.RouteParams) (interface{}, error) {
			return c.service.GetMediaArt(request.Context(), p.Params[queryParamNodeId], p.Params[queryParamMediaId])
		})
}

func initController() (mediaController, error) {
	mediaService, err := NewMediaService()
	if err != nil {
		return mediaController{}, err
	}

	c := mediaController{service: mediaService}

	c.builders = append(
		c.builders,
		c.handleGetAll,
		c.StreamMedia,
		c.DownloadMedia,
		c.DeleteMedia,
		c.handleGetArt,
		c.getAllById,
	)

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
