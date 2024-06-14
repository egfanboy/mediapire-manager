package media

import (
	"context"

	"github.com/egfanboy/mediapire-common/messaging"
	commonTypes "github.com/egfanboy/mediapire-common/types"
	"github.com/egfanboy/mediapire-manager/internal/app"
	"github.com/egfanboy/mediapire-manager/internal/node"
	"github.com/egfanboy/mediapire-manager/internal/rabbitmq"
	"github.com/egfanboy/mediapire-manager/internal/transfer"
	"github.com/egfanboy/mediapire-manager/pkg/types"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"

	mhApi "github.com/egfanboy/mediapire-media-host/pkg/api"
	mhTypes "github.com/egfanboy/mediapire-media-host/pkg/types"
)

type mediaApi interface {
	GetMedia(ctx context.Context) (map[uuid.UUID][]mhTypes.MediaItem, error)
	StreamMedia(ctx context.Context, nodeId uuid.UUID, mediaId uuid.UUID) ([]byte, error)
	DownloadMediaAsync(ctx context.Context, request types.MediaDownloadRequest) (commonTypes.Transfer, error)
	DeleteMedia(ctx context.Context, request types.MediaDeleteRequest) error
}

type mediaService struct {
	nodeRepo     node.NodeRepo
	transferRepo transfer.TransferRepository
}

func (s *mediaService) DownloadMediaAsync(ctx context.Context, request types.MediaDownloadRequest) (commonTypes.Transfer, error) {
	log.Info().Msg("Starting async downloading")

	inputs := make(map[uuid.UUID][]uuid.UUID)

	for _, item := range request {
		if _, ok := inputs[item.NodeId]; ok {
			inputs[item.NodeId] = append(inputs[item.NodeId], item.MediaId)
		} else {
			inputs[item.NodeId] = []uuid.UUID{item.MediaId}
		}
	}
	app := app.GetApp()
	t := transfer.NewTransferModel(app.NodeId, inputs)

	err := s.transferRepo.Save(ctx, t)
	if err != nil {
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

func (s *mediaService) GetMedia(ctx context.Context) (result map[uuid.UUID][]mhTypes.MediaItem, err error) {
	log.Info().Msg("Getting all media from all nodes")
	result = map[uuid.UUID][]mhTypes.MediaItem{}

	nodes, err := s.nodeRepo.GetAllNodes(ctx)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get all nodes")
		return
	}

	for _, node := range nodes {
		if !node.IsUp {
			log.Warn().Msgf("Not fetching media from node %s since it is not up.", node.NodeHost)
			continue
		}

		items, _, errMedia := mhApi.NewClient(ctx).GetMedia(node)
		if errMedia != nil {
			log.Error().Err(errMedia).Msgf("Failed to get media from node %s", node.NodeHost)

			// do not fail the request if one node is unreachable.
			continue
		}

		result[node.Id] = items
	}

	return
}

func (s *mediaService) StreamMedia(ctx context.Context, nodeId uuid.UUID, mediaId uuid.UUID) ([]byte, error) {
	log.Info().Msgf("Streaming media %s from node %s", mediaId, nodeId)
	node, err := s.nodeRepo.GetNode(ctx, nodeId)

	if err != nil {
		log.Error().Err(err).Msgf("Failed to get node with id %s", nodeId)
		return nil, err
	}

	client := mhApi.NewClient(ctx)

	b, _, err := client.StreamMedia(node, mediaId)

	if err != nil {
		log.Error().Err(err).Msgf("Failed stream media on node %s", nodeId)
	}

	return b, err
}

func (s *mediaService) DeleteMedia(ctx context.Context, request types.MediaDeleteRequest) error {
	log.Info().Msgf("Start: delete media")

	inputs := make(map[uuid.UUID][]uuid.UUID)

	for _, item := range request {
		inputs[item.NodeId] = append(inputs[item.NodeId], item.MediaId)
	}

	msg := messaging.DeleteMediaMessage{
		MediaToDelete: inputs,
	}

	err := rabbitmq.PublishMessage(ctx, messaging.TopicDeleteMedia, msg)
	if err != nil {
		log.Err(err).Msgf("Failed to publish message to delete media")
	}

	return err
}

func newMediaService() (mediaApi, error) {
	nodeRepo, err := node.NewNodeRepo()

	if err != nil {
		return nil, err
	}

	// TODO: this whole function should be using context
	transferRepo, err := transfer.NewTransferRepository(context.Background())
	if err != nil {
		return nil, err
	}

	return &mediaService{
		nodeRepo:     nodeRepo,
		transferRepo: transferRepo,
	}, nil
}
