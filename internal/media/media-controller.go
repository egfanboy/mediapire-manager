package media

import (
	"net/http"

	"github.com/egfanboy/mediapire-manager/internal/app"

	"github.com/egfanboy/mediapire-common/router"
)

const (
	basePath           = "/media"
	queryParamFilePath = "filePath"
	queryParamNodeId   = "nodeId"
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
		AddQueryParam(router.QueryParam{Name: queryParamFilePath, Required: true}).
		AddQueryParam(router.QueryParam{Name: queryParamNodeId, Required: true}).
		SetHandler(func(request *http.Request, p router.RouteParams) (interface{}, error) {
			return c.service.StreamMedia(request.Context(), p.Params[queryParamNodeId], p.Params[queryParamFilePath])
		})
}

func initController() mediaController {
	c := mediaController{service: newMediaService()}

	c.builders = append(c.builders, c.handleGetAll, c.StreamMedia)

	return c
}

func init() {
	app.GetApp().ControllerRegistry.Register(initController())
}
