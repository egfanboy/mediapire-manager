package settings

import (
	"net/http"

	"github.com/egfanboy/mediapire-manager/internal/app"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"

	"github.com/egfanboy/mediapire-common/router"
)

const (
	basePath         = "/settings"
	queryParamNodeId = "nodeId"
)

type settingsController struct {
	builders []func() router.RouteBuilder
	service  settingsApi
}

func (c settingsController) GetApis() (routes []router.RouteBuilder) {
	for _, b := range c.builders {

		routes = append(routes, b())
	}

	return
}

func (c settingsController) getSettings() router.RouteBuilder {
	return router.NewV1RouteBuilder().
		SetMethod(http.MethodOptions, http.MethodGet).
		SetPath(basePath).
		SetReturnCode(http.StatusOK).
		AddQueryParam(router.QueryParam{Name: queryParamNodeId, Required: false}).
		SetHandler(func(request *http.Request, p router.RouteParams) (interface{}, error) {
			if nodeId, ok := p.Params[queryParamNodeId]; ok {
				return c.service.GetNodeSettings(request.Context(), uuid.MustParse(nodeId))
			} else {
				return c.service.GetSettings(request.Context())
			}

		})
}

func initController() (settingsController, error) {
	service, err := newSettingsService()
	if err != nil {
		return settingsController{}, err
	}

	c := settingsController{service: service}

	c.builders = append(c.builders, c.getSettings)

	return c, nil
}

func init() {
	ctrl, err := initController()
	if err != nil {
		log.Error().Err(err).Msg("Failed to instantiate settings controller")
	} else {
		app.GetApp().ControllerRegistry.Register(ctrl)
	}
}
