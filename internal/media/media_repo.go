package media

import (
	"context"
	"fmt"
	"sync"

	"github.com/egfanboy/mediapire-manager/internal/utils"
	"github.com/egfanboy/mediapire-manager/pkg/types"
)

type mediaRepo interface {
	GetMedia(ctx context.Context, filter getMediaFilter) ([]types.MediaItem, error)
	SaveItems(ctx context.Context, items []types.MediaItem) error
	DeleteMany(ctx context.Context, filter deleteManyFilter) error
}

type excludeFilter struct {
	FieldName string
	Values    []any
}

func newExcludeFilter[T comparable](fieldName string, typedValues []T) *excludeFilter {
	values := make([]any, len(typedValues))

	for i, v := range typedValues {
		values[i] = v
	}

	return &excludeFilter{FieldName: fieldName, Values: values}
}

type getMediaFilter struct {
	Id         *string
	NodeIds    []string
	MediaTypes []string
	Exclude    *excludeFilter
}

func (f getMediaFilter) IsEmpty() bool {
	if f.Id != nil {
		return false
	}

	if len(f.NodeIds) > 0 {
		return false
	}

	if len(f.MediaTypes) > 0 {
		return false
	}

	if f.Exclude != nil {
		return false
	}

	return true
}

type deleteManyFilter struct {
	NodeId *string
}

func (f deleteManyFilter) IsEmpty() bool {
	if f.NodeId != nil {
		return false
	}

	return true
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

	if filter.IsEmpty() {
		return r.mediaItems, nil
	}

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

			matchesValue := false
			for _, excludedValue := range filter.Exclude.Values {
				if jsonItem[filter.Exclude.FieldName] == excludedValue {
					matchesValue = true

				}

			}

			if !matchesValue {
				result = append(result, item)
			}

		}
	}

	return result, nil
}

func (r *inMemoryRepo) SaveItems(ctx context.Context, items []types.MediaItem) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.mediaItems = items

	return nil
}

func (r *inMemoryRepo) DeleteMany(ctx context.Context, filter deleteManyFilter) error {
	if filter.IsEmpty() {
		return fmt.Errorf("cannot delete many without a filter")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	result := make([]types.MediaItem, 0)

	for _, item := range r.mediaItems {
		if filter.NodeId != nil && item.NodeId != *filter.NodeId {
			result = append(result, item)
		}
	}

	r.mediaItems = result

	return nil
}

func newMediaRepo(ctx context.Context) (mediaRepo, error) {
	return inMemoryRepoInst, nil
}
