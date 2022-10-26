package node

import (
	"net/http"

	"github.com/egfanboy/mediapire-manager/internal/app"

	"github.com/egfanboy/mediapire-common/router"
)

const basePath = "/health"

type healthController struct {
	builders []func() router.RouteBuilder
	service  healthApi
}

func (c healthController) GetApis() (routes []router.RouteBuilder) {
	for _, b := range c.builders {

		routes = append(routes, b())
	}

	return
}

func (c healthController) getHealth() router.RouteBuilder {
	return router.NewV1RouteBuilder().
		SetMethod(http.MethodOptions, http.MethodGet).
		SetPath(basePath).
		SetReturnCode(http.StatusOK).
		SetHandler(func(request *http.Request, p router.RouteParams) (interface{}, error) {
			err := c.service.GetHealth(request.Context())
			return nil, err
		})
}

func initController() healthController {
	c := healthController{service: newNodeService()}

	c.builders = append(c.builders, c.getHealth)

	return c
}

func init() {
	app.GetApp().ControllerRegistry.Register(initController())
}
