package media

import (
	"context"
	"sync"

	"github.com/egfanboy/mediapire-manager/internal/utils"
	"github.com/egfanboy/mediapire-manager/pkg/types"
)

type excludeFilter struct {
	FieldName string
	Value     interface{}
}

type getMediaFilter struct {
	Id         *string
	NodeIds    []string
	MediaTypes []string
	Exclude    *excludeFilter
}

type mediaRepo interface {
	GetMedia(ctx context.Context, filter getMediaFilter) ([]types.MediaItem, error)
	SaveItems(ctx context.Context, items []types.MediaItem) error
}

type inMemoryRepo struct {
	mu sync.RWMutex

	mediaItems []types.MediaItem
}

var inMemoryRepoInst = &inMemoryRepo{}

func (r *inMemoryRepo) GetMedia(ctx context.Context, filter getMediaFilter) ([]types.MediaItem, error) {
	result := make([]types.MediaItem, 0)

	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, item := range r.mediaItems {
		if filter.Id != nil && item.Id == *filter.Id {
			result = append(result, item)

			return result, nil
		}

		if len(filter.NodeIds) > 0 {
			for _, nodeId := range filter.NodeIds {
				if nodeId == item.NodeId {
					result = append(result, item)
				}
			}
		}

		if len(filter.MediaTypes) > 0 {
			for _, mediaType := range filter.MediaTypes {
				if mediaType == item.Extension {
					result = append(result, item)
				}
			}
		}

		if filter.Exclude != nil {
			jsonItem, err := utils.ConvertStruct[types.MediaItem, map[string]interface{}](item)
			if err != nil {
				return nil, err
			}

			if jsonItem[filter.Exclude.FieldName] != filter.Exclude.Value {
				result = append(result, item)
			}

		}
	}

	return result, nil
}

func (r *inMemoryRepo) SaveItems(ctx context.Context, items []types.MediaItem) error {
	r.mu.RLock()
	defer r.mu.RUnlock()

	r.mediaItems = items

	return nil
}

func newMediaRepo(ctx context.Context) (mediaRepo, error) {
	return inMemoryRepoInst, nil
}
