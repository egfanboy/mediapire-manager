package types

import "github.com/google/uuid"

type MediaDownloadRequestItem struct {
	NodeId  uuid.UUID `json:"nodeId"`
	MediaId uuid.UUID `json:"mediaId"`
}

type MediaDownloadRequest []MediaDownloadRequestItem
