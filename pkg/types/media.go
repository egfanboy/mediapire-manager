package types

import "github.com/google/uuid"

type MediaDownloadRequestItem struct {
	NodeId  string    `json:"nodeId"`
	MediaId uuid.UUID `json:"mediaId"`
}

type MediaDownloadRequest []MediaDownloadRequestItem
