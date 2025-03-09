package media

import (
	"context"
	"sort"

	"github.com/egfanboy/mediapire-manager/internal/node"
	"github.com/egfanboy/mediapire-manager/pkg/types"
	"github.com/rs/zerolog/log"
)

type MediaSync interface {
	HandleNewNode(ctx context.Context, nodeId string) error
	SyncFromAllNodes(ctx context.Context) error
	HandleRemovedNode(ctx context.Context, nodeId string) error
}

type syncService struct {
	mediaService MediaApi
	repo         mediaRepo
	nodeService  node.NodeApi
}

func (s *syncService) HandleNewNode(ctx context.Context, nodeId string) error {
	media, err := s.mediaService.GetMediaByNodeId(ctx, []string{}, nodeId)
	if err != nil {
		return err
	}

	otherMedia, err := s.repo.GetMedia(ctx, getMediaFilter{Exclude: &excludeFilter{FieldName: "nodeId", Values: []interface{}{nodeId}}})
	if err != nil {
		return err
	}

	media = append(media, otherMedia...)

	sortedItems := sortMediaByExtension(media)

	err = s.repo.SaveItems(ctx, sortedItems)
	if err != nil {
		return err
	}

	return nil
}

func (s *syncService) SyncFromAllNodes(ctx context.Context) error {
	nodes, err := s.nodeService.GetAllNodes(ctx)
	if err != nil {
		return err
	}

	nodesToFetch := make([]string, 0)

	for _, node := range nodes {
		if node.IsUp {
			nodesToFetch = append(nodesToFetch, node.Id)
		} else {
			log.Debug().Msgf("node %s is not up and will be skipped when fetching media", node.Id)
		}
	}

	media, err := s.mediaService.InternalGetAllMediaFromNodes(ctx, nodesToFetch)
	if err != nil {
		return err
	}

	sortedItems := sortMediaByExtension(media)

	err = s.repo.SaveItems(ctx, sortedItems)
	if err != nil {
		return err
	}

	return nil
}

func (s *syncService) HandleRemovedNode(ctx context.Context, nodeId string) error {
	return s.repo.DeleteMany(ctx, deleteManyFilter{NodeId: &nodeId})
}

func sortMediaByExtension(media []types.MediaItem) []types.MediaItem {
	mediaByExtension := make(map[string][]types.MediaItem)

	for _, m := range media {
		mediaByExtension[m.Extension] = append(mediaByExtension[m.Extension], m)
	}

	sortedItems := make([]types.MediaItem, 0)

	for _, v := range mediaByExtension {
		sort.SliceStable(v, func(i, j int) bool {
			return v[i].Name < v[j].Name
		})

		sortedItems = append(sortedItems, v...)
	}

	return sortedItems
}

func NewMediaSyncService(ctx context.Context) (MediaSync, error) {
	mediaService, err := NewMediaService()
	if err != nil {
		return nil, err
	}

	repo, err := newMediaRepo(ctx)
	if err != nil {
		return nil, err
	}

	nodeService, err := node.NewNodeService()
	if err != nil {
		return nil, err
	}

	return &syncService{mediaService: mediaService, repo: repo, nodeService: nodeService}, nil
}
