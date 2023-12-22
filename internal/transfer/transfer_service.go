package transfer

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path"

	"github.com/egfanboy/mediapire-common/exceptions"
	"github.com/egfanboy/mediapire-manager/internal/app"
	"github.com/rs/zerolog/log"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type transfersApi interface {
	Download(ctx context.Context, transferId string) ([]byte, error)
}

type transfersService struct {
}

func (s *transfersService) Download(ctx context.Context, transferId string) ([]byte, error) {
	log.Info().Msg("Download Transfer: start")

	transferObjectId, err := primitive.ObjectIDFromHex(transferId)
	if err != nil {
		log.Err(err).Msgf("Failed to convert transferId %s to an ObjectId", transferId)
		return nil, err
	}

	transferRepo, err := NewTransferRepository(ctx)
	if err != nil {
		log.Err(err).Msg("failed to instantiate transfer repository")
		return nil, err
	}

	transfer, err := transferRepo.GetById(ctx, transferObjectId)
	if err != nil {
		log.Err(err).Msgf("failed to get transfer with id %s from the database", transferId)
		return nil, err
	}

	if transfer.Status != StatusComplete {
		err = fmt.Errorf(
			"cannot download content from transfer %s since it is not in %s status. current status: %s",
			transferId, StatusProcessComplete,
			transfer.Status,
		)

		log.Err(err)
		return nil, &exceptions.ApiException{
			Err: err, StatusCode: http.StatusBadRequest,
		}
	}

	fileContent, err := os.ReadFile(path.Join(app.GetApp().Config.DownloadPath, transferId+".zip"))
	if err != nil {
		log.Err(err).Msgf("Failed to open item for transfer with id %s", transferId)
		return nil, err
	}

	return fileContent, nil
}

func newTransfersService() transfersApi {
	return &transfersService{}
}
