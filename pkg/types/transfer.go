package types

type TransferCreateRequest struct {
	TargetId *string            `json:"targetId"`
	Inputs   []MediaItemMapping `json:"inputs"`
}
