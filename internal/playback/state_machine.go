package playback

import (
	"errors"
	"fmt"
	"math/rand"

	"github.com/egfanboy/mediapire-manager/pkg/types"
)

func buildPlayOrder(queueSize int, shuffleEnabled bool, seed int64) []int {
	order := make([]int, queueSize)
	for i := 0; i < queueSize; i++ {
		order[i] = i
	}

	if !shuffleEnabled || queueSize <= 1 {
		return order
	}

	r := rand.New(rand.NewSource(seed))
	for i := queueSize - 1; i > 0; i-- {
		j := r.Intn(i + 1)
		order[i], order[j] = order[j], order[i]
	}

	return order
}

func getOrderPosByQueueIndex(playOrder []int, queueIndex int) (int, error) {
	for i, idx := range playOrder {
		if idx == queueIndex {
			return i, nil
		}
	}

	return -1, errors.New("queue index not found in play order")
}

func setQueue(session *playbackSession, items []types.MediaItemMapping, startIndex *int, shuffleSeed int64) error {
	session.Queue = items

	if len(items) == 0 {
		session.PlayOrder = make([]int, 0)
		session.CurrentPlayOrderIndex = -1
		return nil
	}

	if startIndex != nil && (*startIndex < 0 || *startIndex >= len(items)) {
		return fmt.Errorf("startIndex %d is out of bounds", *startIndex)
	}

	if session.ShuffleEnabled {
		session.ShuffleSeed = shuffleSeed
	}

	session.PlayOrder = buildPlayOrder(len(items), session.ShuffleEnabled, session.ShuffleSeed)

	if startIndex == nil && session.ShuffleEnabled {
		session.CurrentPlayOrderIndex = 0
		return nil
	}

	indexToUse := 0
	if startIndex != nil {
		indexToUse = *startIndex
	}

	orderPos, err := getOrderPosByQueueIndex(session.PlayOrder, indexToUse)
	if err != nil {
		return err
	}

	session.CurrentPlayOrderIndex = orderPos
	return nil
}

func setCurrentIndex(session *playbackSession, index int) error {
	if len(session.Queue) == 0 {
		return errors.New("queue is empty")
	}

	if index < 0 || index >= len(session.Queue) {
		return fmt.Errorf("index %d is out of bounds", index)
	}

	orderPos, err := getOrderPosByQueueIndex(session.PlayOrder, index)
	if err != nil {
		return err
	}

	session.CurrentPlayOrderIndex = orderPos
	return nil
}

func next(session *playbackSession) {
	if len(session.PlayOrder) == 0 {
		session.CurrentPlayOrderIndex = -1
		return
	}

	if session.RepeatMode == repeatModeOne {
		if session.CurrentPlayOrderIndex < 0 {
			session.CurrentPlayOrderIndex = 0
		}
		return
	}

	lastIndex := len(session.PlayOrder) - 1
	if session.CurrentPlayOrderIndex < 0 {
		session.CurrentPlayOrderIndex = 0
		return
	}

	switch session.RepeatMode {
	case repeatModeAll:
		if session.CurrentPlayOrderIndex >= lastIndex {
			session.CurrentPlayOrderIndex = 0
			return
		}
		session.CurrentPlayOrderIndex++
	default:
		if session.CurrentPlayOrderIndex < lastIndex {
			session.CurrentPlayOrderIndex++
		}
	}
}

func previous(session *playbackSession) {
	if len(session.PlayOrder) == 0 {
		session.CurrentPlayOrderIndex = -1
		return
	}

	if session.RepeatMode == repeatModeOne {
		if session.CurrentPlayOrderIndex < 0 {
			session.CurrentPlayOrderIndex = 0
		}
		return
	}

	lastIndex := len(session.PlayOrder) - 1
	if session.CurrentPlayOrderIndex < 0 {
		session.CurrentPlayOrderIndex = lastIndex
		return
	}

	switch session.RepeatMode {
	case repeatModeAll:
		if session.CurrentPlayOrderIndex == 0 {
			session.CurrentPlayOrderIndex = lastIndex
			return
		}
		session.CurrentPlayOrderIndex--
	default:
		if session.CurrentPlayOrderIndex > 0 {
			session.CurrentPlayOrderIndex--
		}
	}
}
