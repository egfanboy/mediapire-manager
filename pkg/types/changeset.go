package types

import "time"

type MediaItemChange struct {
	Title       string `json:"title"`
	Artist      string `json:"artist"`
	Album       string `json:"album"`
	Comment     string `json:"comment"`
	Genre       string `json:"genre"`
	TrackNumber int    `json:"trackNumber"`
	// Art string `json:""`
}

type Changeset struct {
	MediaItemMapping
	Change MediaItemChange `json:"change"`
}

type ChangesetCreateRequest struct {
	Action  string      `json:"action"`
	Changes []Changeset `json:"changes"`
}

type ChangesetItem struct {
	Id string `json:"id"`
	// enum driven by model, simply set to string for representation
	Status        string     `json:"status"`
	FailureReason string     `json:"failureReason" `
	Expiry        *time.Time `json:"expiry" `
	Type          string     `json:"type"`
}
