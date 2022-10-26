package node

import (
	"net/http"

	"github.com/egfanboy/mediapire-manager/internal/app"
	"github.com/rs/zerolog/log"

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

func (c nodeController) getAllNodes() router.RouteBuilder {
	return router.NewV1RouteBuilder().
		SetMethod(http.MethodOptions, http.MethodGet).
		SetPath(basePath).
		SetReturnCode(http.StatusOK).
		SetHandler(func(request *http.Request, p router.RouteParams) (interface{}, error) {
			return c.service.GetAllNodes(request.Context())
		})
}

func initController() (nodeController, error) {

	nodeService, err := newNodeService()

	if err != nil {
		return nodeController{}, err
	}
	c := nodeController{service: nodeService}

	c.builders = append(c.builders, c.getAllNodes)

	return c, nil
}

func init() {
	controller, err := initController()

	if err != nil {
		log.Error().Err(err).Msg("Failed to instantiate node controller")
	} else {
		app.GetApp().ControllerRegistry.Register(controller)
	}

}
