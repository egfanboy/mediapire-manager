package types

import "time"

type RepeatMode string

const (
	RepeatModeOff RepeatMode = "off"
	RepeatModeOne RepeatMode = "one"
	RepeatModeAll RepeatMode = "all"
)

type PlaybackSessionState struct {
	Queue                 []MediaItemMapping `json:"queue"`
	PlayOrder             []int              `json:"playOrder"`
	CurrentPlayOrderIndex int                `json:"currentPlayOrderIndex"`
	CurrentItem           *MediaItemMapping  `json:"currentItem,omitempty"`
	CurrentQueueIndex     *int               `json:"currentQueueIndex,omitempty"`
	CurrentMedia          *MediaItem         `json:"currentMedia,omitempty"`
	ShuffleEnabled        bool               `json:"shuffleEnabled"`
	RepeatMode            string             `json:"repeatMode"`
	UpdatedAt             time.Time          `json:"updatedAt"`
}

type PlaybackSetQueueRequest struct {
	Items      []MediaItemMapping `json:"items"`
	StartIndex *int               `json:"startIndex"`
}

type PlaybackStartRequest struct {
	MediaType      *string `json:"mediaType,omitempty"`
	MediaIds       *string `json:"mediaIds,omitempty"`
	SortBy         *string `json:"sortBy,omitempty"`
	StartIndex     *int    `json:"startIndex,omitempty"`
	ShuffleEnabled *bool   `json:"shuffleEnabled,omitempty"`
	RepeatMode     *string `json:"repeatMode,omitempty"`
}

type PlaybackCommandPayload struct {
	Mode    *string `json:"mode,omitempty"`
	Enabled *bool   `json:"enabled,omitempty"`
	Index   *int    `json:"index,omitempty"`
}

type PlaybackCommandRequest struct {
	Command string                 `json:"command"`
	Payload PlaybackCommandPayload `json:"payload"`
}
