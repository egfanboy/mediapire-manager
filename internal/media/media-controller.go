package media

import (
	"mediapire/manager/internal/app"
	"net/http"

	"github.com/egfanboy/mediapire-common/router"
)

const basePath = "/media"

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

func initController() mediaController {
	c := mediaController{service: newMediaService()}

	c.builders = append(c.builders, c.handleGetAll)

	return c
}

func init() {
	app.GetApp().ControllerRegistry.Register(initController())
}
