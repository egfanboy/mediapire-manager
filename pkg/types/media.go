package types

type MediaItemMapping struct {
	NodeId  string `json:"nodeId"`
	MediaId string `json:"mediaId"`
}

type MediaItem struct {
	NodeId    string      `json:"nodeId"`
	Name      string      `json:"name"`
	Extension string      `json:"extension"`
	Id        string      `json:"id"`
	Metadata  interface{} `json:"metadata"`
}

type MediaDownloadRequest []MediaItemMapping
type MediaDeleteRequest []MediaItemMapping

type MediaResponse struct {
	Result []MediaItem `json:"result"`
}
