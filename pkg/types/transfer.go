package types

import "github.com/google/uuid"

type TransferCreateRequest struct {
	TargetId *uuid.UUID         `json:"targetId"`
	Inputs   []MediaItemMapping `json:"inputs"`
}
