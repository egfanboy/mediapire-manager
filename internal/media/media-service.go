package media

import (
	"context"

	"github.com/egfanboy/mediapire-manager/internal/node"
	"github.com/egfanboy/mediapire-manager/pkg/types"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"

	mhApi "github.com/egfanboy/mediapire-media-host/pkg/api"
	mhTypes "github.com/egfanboy/mediapire-media-host/pkg/types"
)

type mediaApi interface {
	GetMedia(ctx context.Context) (map[string][]mhTypes.MediaItem, error)
	StreamMedia(ctx context.Context, nodeId string, mediaId uuid.UUID) ([]byte, error)
	DownloadMedia(ctx context.Context, request types.MediaDownloadRequest) ([]byte, error)
}

type mediaService struct {
	nodeRepo node.NodeRepo
}

func (s *mediaService) GetMedia(ctx context.Context) (result map[string][]mhTypes.MediaItem, err error) {
	log.Info().Msg("Getting all media from all nodes")
	result = map[string][]mhTypes.MediaItem{}

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

		result[node.Host()] = items
	}

	return
}

func (s *mediaService) StreamMedia(ctx context.Context, nodeId string, mediaId uuid.UUID) ([]byte, error) {
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

func (s *mediaService) DownloadMedia(ctx context.Context, request types.MediaDownloadRequest) ([]byte, error) {
	log.Debug().Msg("Start: download media")

	populatedItems := make([]populatedDownloadItem, 0)

	// build a simple cache to just fetch a node once
	nodeCache := make(map[string]node.NodeConfig)

	for _, item := range request {
		populatedItem := populatedDownloadItem{
			MediaId: item.MediaId,
		}

		if n, ok := nodeCache[item.NodeId]; ok {
			populatedItem.Node = n
		} else {
			node, err := s.nodeRepo.GetNode(ctx, item.NodeId)
			if err != nil {
				return nil, err
			}

			nodeCache[item.NodeId] = node
			populatedItem.Node = node
		}

		populatedItems = append(populatedItems, populatedItem)
	}

	downloader := mediaDownloader{}

	return downloader.Download(ctx, populatedItems)
}

func newMediaService() (mediaApi, error) {
	repo, err := node.NewNodeRepo()

	if err != nil {
		return nil, err
	}

	return &mediaService{
		nodeRepo: repo,
	}, nil
}
