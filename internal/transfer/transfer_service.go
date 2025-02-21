package transfer

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path"
	"time"

	"github.com/egfanboy/mediapire-common/exceptions"
	"github.com/egfanboy/mediapire-common/messaging"
	commonTypes "github.com/egfanboy/mediapire-common/types"
	"github.com/egfanboy/mediapire-manager/internal/app"
	"github.com/egfanboy/mediapire-manager/internal/rabbitmq"
	"github.com/egfanboy/mediapire-manager/pkg/types"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type transfersApi interface {
	Download(ctx context.Context, transferId string) ([]byte, error)
	GetTransfer(ctx context.Context, transferId string) (commonTypes.Transfer, error)
	CleanupTransfer(ctx context.Context, transferId string) error
	CreateTransfer(ctx context.Context, body types.TransferCreateRequest) (commonTypes.Transfer, error)
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

	if transfer.Expiry.After(time.Now()) {
		err = fmt.Errorf(
			"cannot download content from transfer %s since it is expired",
			transferId,
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

func (s *transfersService) getTransferModel(ctx context.Context, transferId string) (*Transfer, error) {
	log.Info().Msgf("Get Transfer model: start id %s", transferId)

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

	return transfer, nil
}

func (s *transfersService) GetTransfer(ctx context.Context, transferId string) (commonTypes.Transfer, error) {
	log.Info().Msgf("Get Transfer: start id %s", transferId)

	transfer, err := s.getTransferModel(ctx, transferId)
	if err != nil {
		return commonTypes.Transfer{}, err
	}

	return transfer.ToApiResponse(), nil
}

func (t *transfersService) CleanupTransfer(ctx context.Context, transferId string) error {
	log.Info().Msgf("Cleanup Transfer: start id %s", transferId)

	err := os.RemoveAll(path.Join(app.GetApp().Config.DownloadPath, transferId+".zip"))
	if err != nil {
		log.Err(err).Msgf("Failed to remove zip file for transfer %s", transferId)
	}

	return err
}

func (s *transfersService) CreateTransfer(ctx context.Context, request types.TransferCreateRequest) (commonTypes.Transfer, error) {
	log.Info().Msg("Create Transfer: start")

	transferRepo, err := NewTransferRepository(ctx)
	if err != nil {
		log.Err(err).Msg("failed to instantiate transfer repository")
		return commonTypes.Transfer{}, err
	}

	inputs := make(map[uuid.UUID][]string)

	for _, item := range request.Inputs {
		if _, ok := inputs[item.NodeId]; ok {
			inputs[item.NodeId] = append(inputs[item.NodeId], item.MediaId)
		} else {
			inputs[item.NodeId] = []string{item.MediaId}
		}
	}

	targetId := app.GetApp().NodeId
	if request.TargetId != nil {
		targetId = *request.TargetId
	}

	t := NewTransferModel(targetId, inputs)

	err = transferRepo.Save(ctx, t)
	if err != nil {
		log.Err(err).Msg("failed to save new transfer")
		return commonTypes.Transfer{}, err
	}

	msg := messaging.TransferMessage{
		Id:       t.Id.Hex(),
		TargetId: t.TargetId,
		Inputs:   t.Inputs,
	}

	err = rabbitmq.PublishMessage(ctx, messaging.TopicTransfer, msg)
	if err != nil {
		log.Err(err).Msg("failed to start async download")
	}

	return t.ToApiResponse(), err

}

func newTransfersService() transfersApi {
	return &transfersService{}
}
