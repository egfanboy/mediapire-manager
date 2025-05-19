package media

import (
	"context"
	"encoding/json"
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
	SortBy     *string
	OrderBy    *string
	Ids        []string
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

	if f.SortBy != nil && f.OrderBy != nil {
		return false
	}

	if len(f.Ids) > 0 {
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

	mediaItems []byte
}

var inMemoryRepoInst = &inMemoryRepo{}

func (r *inMemoryRepo) GetMedia(ctx context.Context, filter getMediaFilter) ([]types.MediaItem, error) {
	result := make([]map[string]any, 0)

	r.mu.RLock()
	defer r.mu.RUnlock()

	if filter.IsEmpty() {
		var result []types.MediaItem
		err := json.Unmarshal(r.mediaItems, &result)
		return result, err
	}

	var items []map[string]interface{}

	err := json.Unmarshal(r.mediaItems, &items)
	if err != nil {
		return nil, err
	}

	for _, item := range items {
		if filter.Id != nil && item["id"] == *filter.Id {
			result = append(result, item)

			return utils.ConvertStruct[[]map[string]any, []types.MediaItem](result)
		}

		matchesFilters := make([]bool, 0)

		if len(filter.NodeIds) > 0 {
			matchesAny := false
			for _, nodeId := range filter.NodeIds {
				if nodeId == item["nodeId"] {
					matchesAny = true
					break
				}
			}

			matchesFilters = append(matchesFilters, matchesAny)
		}

		if len(filter.MediaTypes) > 0 {
			matchesAny := false
			for _, mediaType := range filter.MediaTypes {
				if mediaType == item["extension"] {
					matchesAny = true
					break
				}
			}

			matchesFilters = append(matchesFilters, matchesAny)
		}

		if len(filter.Ids) > 0 {
			matchesAny := false

			for _, id := range filter.Ids {
				if id == item["id"] {
					matchesAny = true
					break
				}
			}

			matchesFilters = append(matchesFilters, matchesAny)
		}

		if filter.Exclude != nil {
			matchesValue := false
			for _, excludedValue := range filter.Exclude.Values {
				if item[filter.Exclude.FieldName] == excludedValue {
					matchesValue = true
					break
				}

			}

			matchesFilters = append(matchesFilters, !matchesValue)

		}

		shouldBeAdded := true
		for _, matched := range matchesFilters {
			// did not match one filter, do not include it
			if !matched {
				shouldBeAdded = false
				break
			}
		}

		if shouldBeAdded {
			result = append(result, item)
		}

	}

	if filter.SortBy != nil {
		err := sortMedia(result, *filter.SortBy, *filter.OrderBy)
		if err != nil {
			return nil, err
		}
	}

	return utils.ConvertStruct[[]map[string]any, []types.MediaItem](result)
}

func (r *inMemoryRepo) SaveItems(ctx context.Context, mediaItems []types.MediaItem) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	items, err := json.Marshal(mediaItems)
	if err != nil {
		return err
	}

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

	var items []types.MediaItem

	err := json.Unmarshal(r.mediaItems, &items)
	if err != nil {
		return err
	}

	for _, item := range items {
		if filter.NodeId != nil && item.NodeId != *filter.NodeId {
			result = append(result, item)
		}
	}

	itemsBytes, err := json.Marshal(result)
	if err != nil {
		return err
	}

	r.mediaItems = itemsBytes

	return nil
}

func newMediaRepo(ctx context.Context) (mediaRepo, error) {
	return inMemoryRepoInst, nil
}
