package playback

import (
	"context"
	"net/http"

	"github.com/egfanboy/mediapire-common/router"
	"github.com/egfanboy/mediapire-manager/internal/app"
	"github.com/egfanboy/mediapire-manager/pkg/types"
	"github.com/rs/zerolog/log"
)

const basePath = "/playback/session"

type playbackController struct {
	builders []func() router.RouteBuilder
	service  PlaybackApi
}

func (c playbackController) GetApis() (routes []router.RouteBuilder) {
	for _, b := range c.builders {
		routes = append(routes, b())
	}

	return
}

func (c playbackController) getSession() router.RouteBuilder {
	return router.NewV1RouteBuilder().
		SetMethod(http.MethodOptions, http.MethodGet).
		SetPath(basePath).
		SetReturnCode(http.StatusOK).
		SetHandler(func(request *http.Request, p router.RouteParams) (interface{}, error) {
			return c.service.GetSession(request.Context())
		})
}

func (c playbackController) setQueue() router.RouteBuilder {
	return router.NewV1RouteBuilder().
		SetMethod(http.MethodOptions, http.MethodPost).
		SetPath(basePath + "/queue").
		SetReturnCode(http.StatusOK).
		SetHandler(func(request *http.Request, p router.RouteParams) (interface{}, error) {
			var body types.PlaybackSetQueueRequest
			if err := p.PopulateBody(&body); err != nil {
				return nil, err
			}

			return c.service.SetQueue(request.Context(), body)
		})
}

func (c playbackController) startFromMedia() router.RouteBuilder {
	return router.NewV1RouteBuilder().
		SetMethod(http.MethodOptions, http.MethodPost).
		SetPath(basePath + "/start").
		SetReturnCode(http.StatusOK).
		SetHandler(func(request *http.Request, p router.RouteParams) (interface{}, error) {
			var body types.PlaybackStartRequest
			if err := p.PopulateBody(&body); err != nil {
				return nil, err
			}

			return c.service.StartFromMedia(request.Context(), body)
		})
}

func (c playbackController) postCommand() router.RouteBuilder {
	return router.NewV1RouteBuilder().
		SetMethod(http.MethodOptions, http.MethodPost).
		SetPath(basePath + "/commands").
		SetReturnCode(http.StatusOK).
		SetHandler(func(request *http.Request, p router.RouteParams) (interface{}, error) {
			var body types.PlaybackCommandRequest
			if err := p.PopulateBody(&body); err != nil {
				return nil, err
			}

			return c.service.ApplyCommand(request.Context(), body)
		})
}

func initController() (playbackController, error) {
	service, err := newPlaybackService(context.Background())
	if err != nil {
		return playbackController{}, err
	}

	controller := playbackController{service: service}
	controller.builders = append(controller.builders, controller.getSession, controller.setQueue, controller.startFromMedia, controller.postCommand)

	return controller, nil
}

func init() {
	controller, err := initController()
	if err != nil {
		log.Error().Err(err).Msg("failed to instantiate playback controller")
		return
	}

	app.GetApp().ControllerRegistry.Register(controller)
}
