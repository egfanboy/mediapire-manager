package node

import (
	"mediapire/manager/internal/app"
	"net/http"

	"github.com/egfanboy/mediapire-common/router"
)

const basePath = "/nodes"

type nodeController struct {
	builders []func() router.RouteBuilder
	service  nodeApi
}

func (c nodeController) GetApis() (routes []router.RouteBuilder) {
	for _, b := range c.builders {

		routes = append(routes, b())
	}

	return
}

func (c nodeController) HandleRegister() router.RouteBuilder {
	return router.NewV1RouteBuilder().
		SetMethod(http.MethodOptions, http.MethodPost).
		SetPath(basePath + "/register").
		SetReturnCode(http.StatusNoContent).
		SetHandler(func(request *http.Request, p router.RouteParams) (interface{}, error) {
			req := new(RegisterNodeRequest)

			err := p.PopulateBody(req)

			if err != nil {
				return nil, err
			}

			err = c.service.RegisterNode(request.Context(), *req)
			return nil, err
		})
}

func initController() nodeController {
	c := nodeController{service: newNodeService()}

	c.builders = append(c.builders, c.HandleRegister)

	return c
}

func init() {
	app.GetApp().ControllerRegistry.Register(initController())
}
