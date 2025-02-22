package types

type MediaItemMapping struct {
	NodeId  string `json:"nodeId"`
	MediaId string `json:"mediaId"`
}

type MediaDownloadRequest []MediaItemMapping
type MediaDeleteRequest []MediaItemMapping
