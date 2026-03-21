package playback

import (
	"time"

	"github.com/egfanboy/mediapire-manager/pkg/types"
)

type repeatMode string

const (
	repeatModeOff repeatMode = "off"
	repeatModeOne repeatMode = "one"
	repeatModeAll repeatMode = "all"
)

type playbackSession struct {
	Id                    string                   `bson:"_id"`
	Queue                 []types.MediaItemMapping `bson:"queue"`
	PlayOrder             []int                    `bson:"playOrder"`
	CurrentPlayOrderIndex int                      `bson:"currentPlayOrderIndex"`
	ShuffleEnabled        bool                     `bson:"shuffleEnabled"`
	ShuffleSeed           int64                    `bson:"shuffleSeed"`
	RepeatMode            repeatMode               `bson:"repeatMode"`
	CreatedAt             time.Time                `bson:"createdAt"`
	UpdatedAt             time.Time                `bson:"updatedAt"`
}

func newPlaybackSession(id string, now time.Time) *playbackSession {
	return &playbackSession{
		Id:                    id,
		Queue:                 make([]types.MediaItemMapping, 0),
		PlayOrder:             make([]int, 0),
		CurrentPlayOrderIndex: -1,
		ShuffleEnabled:        false,
		ShuffleSeed:           0,
		RepeatMode:            repeatModeOff,
		CreatedAt:             now,
		UpdatedAt:             now,
	}
}

func (s *playbackSession) toApiState() types.PlaybackSessionState {
	state := types.PlaybackSessionState{
		Queue:                 s.Queue,
		PlayOrder:             s.PlayOrder,
		CurrentPlayOrderIndex: s.CurrentPlayOrderIndex,
		ShuffleEnabled:        s.ShuffleEnabled,
		RepeatMode:            string(s.RepeatMode),
		UpdatedAt:             s.UpdatedAt,
	}

	if s.CurrentPlayOrderIndex >= 0 && s.CurrentPlayOrderIndex < len(s.PlayOrder) {
		queueIndex := s.PlayOrder[s.CurrentPlayOrderIndex]
		if queueIndex >= 0 && queueIndex < len(s.Queue) {
			current := s.Queue[queueIndex]
			state.CurrentItem = &current
			state.CurrentQueueIndex = &queueIndex
		}
	}

	return state
}
