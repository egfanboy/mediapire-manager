package transfer

import (
	"time"

	"github.com/egfanboy/mediapire-common/types"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type TransferStatus string

const (
	StatusInProgress      TransferStatus = "in_progress"
	StatusPending         TransferStatus = "pending"
	StatusProcessComplete TransferStatus = "processing_complete"
	StatusComplete        TransferStatus = "complete"
	StatusFailed          TransferStatus = "failed"
	StatusExpired         TransferStatus = "expired"
)

type Transfer struct {
	Id       primitive.ObjectID     `json:"id" bson:"_id,omitempty"`
	TargetId uuid.UUID              `json:"targetId" bson:"target_id"`
	Inputs   map[uuid.UUID][]string `json:"inputs" bson:"inputs"`
	// tracks which input has responded
	Outputs       map[uuid.UUID]bool `json:"-" bson:"outputs"`
	Status        TransferStatus     `json:"status" bson:"status"`
	FailureReason string             `json:"failureReason" bson:"failure_reason"`
	Expiry        time.Time          `json:"expiry" bson:"expiry"`
}

func (t *Transfer) ToApiResponse() types.Transfer {
	return types.Transfer{
		Id:            t.Id.Hex(),
		Status:        string(t.Status),
		FailureReason: &t.FailureReason,
		Expiry:        t.Expiry,
	}
}

func (t *Transfer) DidFail() bool {
	return t.Status == StatusFailed
}

func (t *Transfer) SetFailed(failureReason string) {
	t.Status = StatusFailed
	t.FailureReason = failureReason
}

func (t *Transfer) AllNodesHandled() bool {
	var allNodesHandled = true

	for _, v := range t.Outputs {
		if !v {
			allNodesHandled = false

			return allNodesHandled
		}
	}

	// at this point we looped over every node and it was true
	return allNodesHandled
}

// for now don't expose a function that allows to set the expiry
func newTransferModel(targetId uuid.UUID, inputs map[uuid.UUID][]string, expiry *time.Time) *Transfer {
	outputs := make(map[uuid.UUID]bool)

	for k := range inputs {
		outputs[k] = false
	}

	t := &Transfer{
		Id:       primitive.NewObjectID(),
		TargetId: targetId,
		Inputs:   inputs,
		Status:   StatusPending,
		Outputs:  outputs,
	}

	if expiry != nil {
		t.Expiry = *expiry
	} else {
		t.Expiry = time.Now().Add(time.Hour * 24)
	}

	return t
}

func NewTransferModel(targetId uuid.UUID, inputs map[uuid.UUID][]string) *Transfer {
	return newTransferModel(targetId, inputs, nil)
}
