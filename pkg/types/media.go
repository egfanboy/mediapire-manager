package types

import "github.com/google/uuid"

type MediaItemMapping struct {
	NodeId  uuid.UUID `json:"nodeId"`
	MediaId string    `json:"mediaId"`
}

type MediaDownloadRequest []MediaItemMapping
type MediaDeleteRequest []MediaItemMapping
