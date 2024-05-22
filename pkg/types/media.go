package types

import "github.com/google/uuid"

type MediaItemMapping struct {
	NodeId  uuid.UUID `json:"nodeId"`
	MediaId uuid.UUID `json:"mediaId"`
}

type MediaDownloadRequest []MediaItemMapping
type MediaDeleteRequest []MediaItemMapping
