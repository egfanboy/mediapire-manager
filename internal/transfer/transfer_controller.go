package transfer

import (
	"errors"
	"net/http"

	"github.com/egfanboy/mediapire-common/router"
	"github.com/egfanboy/mediapire-manager/internal/app"
	"github.com/egfanboy/mediapire-manager/pkg/types"
)

const basePath = "/transfers"

type transfersController struct {
	builders []func() router.RouteBuilder
	service  transfersApi
}

func (c transfersController) GetApis() (routes []router.RouteBuilder) {
	for _, b := range c.builders {

		routes = append(routes, b())
	}

	return
}

func (c transfersController) DownloadTransfer() router.RouteBuilder {
	return router.NewV1RouteBuilder().
		SetMethod(http.MethodOptions, http.MethodGet).
		SetPath(basePath + "/{transferId}/download").
		SetDataType(router.DataTypeFile).
		SetReturnCode(http.StatusOK).
		SetHandler(func(request *http.Request, p router.RouteParams) (interface{}, error) {
			transferId, ok := p.Params["transferId"]
			if !ok {
				return nil, errors.New("transferId not found in API path")
			}

			return c.service.Download(request.Context(), transferId)
		})
}

func (c transfersController) GetTransferById() router.RouteBuilder {
	return router.NewV1RouteBuilder().
		SetMethod(http.MethodOptions, http.MethodGet).
		SetPath(basePath + "/{transferId}").
		SetReturnCode(http.StatusOK).
		SetHandler(func(request *http.Request, p router.RouteParams) (interface{}, error) {
			transferId, ok := p.Params["transferId"]
			if !ok {
				return nil, errors.New("transferId not found in API path")
			}

			return c.service.GetTransfer(request.Context(), transferId)
		})
}

func (c transfersController) CreateTransfer() router.RouteBuilder {
	return router.NewV1RouteBuilder().
		SetMethod(http.MethodOptions, http.MethodPost).
		SetPath(basePath).
		SetReturnCode(http.StatusAccepted).
		SetHandler(func(request *http.Request, p router.RouteParams) (interface{}, error) {
			var body types.TransferCreateRequest
			err := p.PopulateBody(&body)
			if err != nil {
				return nil, err
			}

			return c.service.CreateTransfer(request.Context(), body)
		})
}

func initController() transfersController {
	c := transfersController{service: newTransfersService()}

	c.builders = append(
		c.builders,
		c.DownloadTransfer,
		c.GetTransferById,
		c.CreateTransfer,
	)

	return c
}

func init() {
	app.GetApp().ControllerRegistry.Register(initController())
}
