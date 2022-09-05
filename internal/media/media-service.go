package media

import (
	"context"

	"github.com/egfanboy/mediapire-manager/internal/node"

	"github.com/egfanboy/mediapire-media-host/pkg/api"
	mhApi "github.com/egfanboy/mediapire-media-host/pkg/api"
	"github.com/egfanboy/mediapire-media-host/pkg/types"
)

type mediaApi interface {
	GetMedia(ctx context.Context) (map[string][]types.MediaItem, error)
	StreamMedia(ctx context.Context, nodeId string, filePath string) ([]byte, error)
}

type mediaService struct {
	nodeRepo node.NodeRepo
}

func (s *mediaService) GetMedia(ctx context.Context) (result map[string][]types.MediaItem, err error) {
	result = map[string][]types.MediaItem{}

	nodes, err := s.nodeRepo.GetAllNodes(ctx)
	if err != nil {
		return
	}

	for _, node := range nodes {

		items, _, err2 := api.NewClient(ctx).GetMedia(node)
		if err2 != nil {
			err = err2
			return
		}

		result[node.Host()] = items
	}

	return
}

func (s *mediaService) StreamMedia(ctx context.Context, nodeId string, filePath string) ([]byte, error) {
	node, err := s.nodeRepo.GetNode(ctx, nodeId)

	if err != nil {
		return nil, err
	}

	client := mhApi.NewClient(ctx)

	b, _, err := client.StreamMedia(node, filePath)

	return b, err
}

func newMediaService() mediaApi {
	return &mediaService{
		node.NewNodeRepo(),
	}
}
