package media

import (
	"context"

	"github.com/egfanboy/mediapire-manager/internal/node"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"

	"github.com/egfanboy/mediapire-media-host/pkg/api"
	mhApi "github.com/egfanboy/mediapire-media-host/pkg/api"
	"github.com/egfanboy/mediapire-media-host/pkg/types"
)

type mediaApi interface {
	GetMedia(ctx context.Context) (map[string][]types.MediaItem, error)
	StreamMedia(ctx context.Context, nodeId string, mediaId uuid.UUID) ([]byte, error)
}

type mediaService struct {
	nodeRepo node.NodeRepo
}

func (s *mediaService) GetMedia(ctx context.Context) (result map[string][]types.MediaItem, err error) {
	log.Info().Msg("Getting all media from all nodes")
	result = map[string][]types.MediaItem{}

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

		items, _, errMedia := api.NewClient(ctx).GetMedia(node)
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

func newMediaService() (mediaApi, error) {
	repo, err := node.NewNodeRepo()

	if err != nil {
		return nil, err
	}

	return &mediaService{
		nodeRepo: repo,
	}, nil
}
