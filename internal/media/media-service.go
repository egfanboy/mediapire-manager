package media

import (
	"context"
	"fmt"
	"time"

	"github.com/egfanboy/mediapire-common/exceptions"
	"github.com/egfanboy/mediapire-common/messaging"
	commonTypes "github.com/egfanboy/mediapire-common/types"
	"github.com/egfanboy/mediapire-manager/internal/app"
	media_update "github.com/egfanboy/mediapire-manager/internal/media/update"
	"github.com/egfanboy/mediapire-manager/internal/node"
	"github.com/egfanboy/mediapire-manager/internal/rabbitmq"
	"github.com/egfanboy/mediapire-manager/internal/transfer"
	"github.com/egfanboy/mediapire-manager/internal/utils"
	"github.com/egfanboy/mediapire-manager/pkg/types"
	"github.com/rs/zerolog/log"

	mhApi "github.com/egfanboy/mediapire-media-host/pkg/api"
	mhTypes "github.com/egfanboy/mediapire-media-host/pkg/types"
)

type MediaApi interface {
	GetMediaByNodeId(ctx context.Context, mediaTypes []string, nodeId string) ([]types.MediaItem, error)
	StreamMedia(ctx context.Context, nodeId string, mediaId string) ([]byte, error)
	DownloadMediaAsync(ctx context.Context, request types.MediaDownloadRequest) (commonTypes.Transfer, error)
	DeleteMedia(ctx context.Context, request types.MediaDeleteRequest) error
	GetMediaArt(ctx context.Context, nodeId string, mediaId string) ([]byte, error)
	GetMedia(ctx context.Context, mediaTypes []string, nodeIds []string) ([]types.MediaItem, error)
	GetMediaPaginated(
		ctx context.Context,
		mediaTypes []string,
		nodeIds []string,
		page int,
		limit int) (types.PaginatedResponse[types.MediaItem], error)
	// Used by other internal services, not to be exposed via API
	InternalUpdateMedia(ctx context.Context, changesetId string, request []types.Changeset) error
	InternalGetAllMediaFromNodes(ctx context.Context, nodeIds []string) ([]types.MediaItem, error)
}

type mediaService struct {
	nodeRepo     node.NodeRepo
	transferRepo transfer.TransferRepository
	repo         mediaRepo
}

func (s *mediaService) DownloadMediaAsync(ctx context.Context, request types.MediaDownloadRequest) (commonTypes.Transfer, error) {
	log.Info().Msg("Starting async downloading")

	inputs := make(map[string][]string)

	for _, item := range request {
		if _, ok := inputs[item.NodeId]; ok {
			inputs[item.NodeId] = append(inputs[item.NodeId], item.MediaId)
		} else {
			inputs[item.NodeId] = []string{item.MediaId}
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

func (s *mediaService) GetMedia(ctx context.Context, mediaTypes []string, nodeIds []string) (result []types.MediaItem, err error) {
	log.Info().Msg("Getting all media from all nodes")

	nodes, err := s.nodeRepo.GetAllNodes(ctx)
	if err != nil {
		return
	}

	downNodeIds := make([]string, 0)

	for _, node := range nodes {
		if !node.IsUp {
			downNodeIds = append(downNodeIds, node.Id)
		}
	}

	for _, nodeId := range nodeIds {
		for _, downId := range downNodeIds {
			if nodeId == downId {
				err = exceptions.NewBadRequestException(fmt.Errorf("cannot fetch media for node %s since it is down", nodeId))
				return
			}
		}
	}

	result, err = s.repo.GetMedia(
		ctx,
		getMediaFilter{
			MediaTypes: mediaTypes,
			NodeIds:    nodeIds,
			Exclude:    newExcludeFilter("nodeId", downNodeIds),
		},
	)

	return
}

func (s *mediaService) InternalGetAllMediaFromNodes(ctx context.Context, nodeIds []string) (result []types.MediaItem, err error) {
	for _, nodeId := range nodeIds {
		media, getErr := s.GetMediaByNodeId(ctx, []string{}, nodeId)
		if getErr != nil {
			err = getErr
			return
		}

		result = append(result, media...)
	}

	return
}

func (s *mediaService) GetMediaByNodeId(ctx context.Context, mediaTypes []string, nodeId string) (result []types.MediaItem, err error) {
	log.Info().Msgf("Getting all media from node %s", nodeId)

	node, err := s.nodeRepo.GetNode(ctx, nodeId)
	if err != nil {
		log.Error().Err(err).Msgf("Failed to get node %s", nodeId)

		return
	}

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	items, _, err := mhApi.NewClient(node).GetMedia(ctx, &mediaTypes)
	if err != nil {
		log.Error().Err(err).Msgf("Failed to get media from node %s", node.NodeHost)

		// do not fail the request if one node is unreachable.
		return
	}

	for _, item := range items {
		convertedItem, err := utils.ConvertStruct[mhTypes.MediaItem, types.MediaItem](item)
		if err != nil {
			return nil, err
		}

		convertedItem.NodeId = node.Id
		result = append(result, convertedItem)
	}

	return
}

func (s *mediaService) StreamMedia(ctx context.Context, nodeId string, mediaId string) ([]byte, error) {
	log.Info().Msgf("Streaming media %s from node %s", mediaId, nodeId)
	node, err := s.nodeRepo.GetNode(ctx, nodeId)

	if err != nil {
		log.Error().Err(err).Msgf("Failed to get node with id %s", nodeId)
		return nil, err
	}

	client := mhApi.NewClient(node)

	b, _, err := client.StreamMedia(ctx, mediaId)
	if err != nil {
		log.Error().Err(err).Msgf("Failed stream media on node %s", nodeId)
	}

	return b, err
}

func (s *mediaService) DeleteMedia(ctx context.Context, request types.MediaDeleteRequest) error {
	log.Info().Msgf("Start: delete media")

	inputs := make(map[string][]string)

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

func (s *mediaService) GetMediaArt(ctx context.Context, nodeId string, mediaId string) ([]byte, error) {
	log.Info().Msgf("Getting art for media %s from node %s", mediaId, nodeId)
	node, err := s.nodeRepo.GetNode(ctx, nodeId)

	if err != nil {
		log.Error().Err(err).Msgf("Failed to get node with id %s", nodeId)
		return nil, err
	}

	client := mhApi.NewClient(node)

	b, _, err := client.GetMediaArt(ctx, mediaId)
	if err != nil {
		log.Error().Err(err).Msgf("Failed to get art for media on node %s", nodeId)
	}

	return b, err
}

func (s *mediaService) InternalUpdateMedia(ctx context.Context, changeSetId string, changes []types.Changeset) error {
	log.Info().Msg("Start: Update media")
	result := messaging.UpdateMediaMessage{ChangesetId: changeSetId, Items: make(map[string][]messaging.UpdatedItem)}

	clients := make(map[string]mhApi.MediaHostApi)
	var client mhApi.MediaHostApi
	for _, change := range changes {
		if cachedClient, ok := clients[change.NodeId]; ok {
			client = cachedClient
		} else {
			node, err := s.nodeRepo.GetNode(ctx, change.NodeId)
			if err != nil {
				log.Err(err).Msgf("failed to get node %s", change.NodeId)
				return err
			}

			client = mhApi.NewClient(mhTypes.NewHttpHost(node.NodeHost, node.Port()))

			clients[change.NodeId] = client
		}

		mediaItem, _, err := client.GetMediaByIdWithContent(ctx, change.MediaId)
		if err != nil {
			log.Err(err).Msgf("failed to get content for media %s on node %s", change.MediaId, change.NodeId)
			return err
		}

		builder := media_update.GetBuilder(mediaItem)

		if change.Change.Title != "" {
			builder.Title(change.Change.Title)
		}

		if change.Change.Artist != "" {
			builder.Title(change.Change.Artist)
		}

		if change.Change.Album != "" {
			builder.Title(change.Change.Album)
		}

		if change.Change.Comment != "" {
			builder.Title(change.Change.Comment)
		}

		if change.Change.Genre != "" {
			builder.Title(change.Change.Genre)
		}

		if change.Change.TrackNumber != 0 {
			builder.TrackNumber(change.Change.TrackNumber)
		}

		if change.Change.Art != "" {
			builder.Art(change.Change.Art)
		}

		newContent, err := media_update.UpdateMedia(builder)
		if err != nil {
			log.Err(err).Msgf("Failed to update media item %s", change.MediaId)
			return err
		}

		result.Items[change.NodeId] = append(result.Items[change.NodeId], messaging.UpdatedItem{MediaId: change.MediaId, Content: newContent})
	}

	// send message
	err := rabbitmq.PublishMessage(ctx, messaging.TopicUpdateMedia, result)
	if err != nil {
		log.Err(err).Msgf("Failed to publish message to topic %s", messaging.TopicUpdateMedia)
		return err
	}

	log.Info().Msg("End: Update media")
	return nil
}

func (s *mediaService) GetMediaPaginated(
	ctx context.Context,
	mediaTypes []string,
	nodeIds []string,
	page int,
	limit int) (result types.PaginatedResponse[types.MediaItem], err error) {
	log.Info().Msg("Getting paginated media")
	media, err := s.repo.GetMedia(ctx, getMediaFilter{NodeIds: nodeIds, MediaTypes: mediaTypes})
	if err != nil {
		return
	}

	result, err = types.NewPaginatedResponse(media, page, limit)
	return
}
func NewMediaService() (MediaApi, error) {
	nodeRepo, err := node.NewNodeRepo()

	if err != nil {
		return nil, err
	}

	// TODO: this whole function should be using context
	transferRepo, err := transfer.NewTransferRepository(context.Background())
	if err != nil {
		return nil, err
	}

	mediaRepo, err := newMediaRepo(context.Background())
	if err != nil {
		return nil, err
	}

	return &mediaService{
		nodeRepo:     nodeRepo,
		repo:         mediaRepo,
		transferRepo: transferRepo,
	}, nil
}
