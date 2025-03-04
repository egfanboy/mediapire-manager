package changeset

import (
	"time"

	"github.com/egfanboy/mediapire-manager/internal/utils"
	"github.com/egfanboy/mediapire-manager/pkg/types"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type changesetStatus string
type changesetType string

type input struct {
	MediaId string                 `bson:"mediaId"`
	Change  map[string]interface{} `bson:"change"`
}

const (
	StatusInProgress changesetStatus = "in_progress"
	StatusPending    changesetStatus = "pending"
	StatusComplete   changesetStatus = "complete"
	StatusFailed     changesetStatus = "failed"

	TypeUpdate changesetType = "update"
	TypeDelete changesetType = "delete"
)

type Changeset struct {
	Id     primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Inputs map[string][]input `json:"inputs" bson:"inputs"`
	// tracks which input has responded
	Outputs       map[string]bool `json:"-" bson:"outputs"`
	Type          changesetType   `json:"type" bson:"type"`
	Status        changesetStatus `json:"status" bson:"status"`
	FailureReason string          `json:"failureReason" bson:"failure_reason"`
	Expiry        *time.Time      `json:"expiry" bson:"expiry"`
}

func (c *Changeset) ToApiResponse() types.ChangesetItem {
	return types.ChangesetItem{
		Id:            c.Id.Hex(),
		Status:        string(c.Status),
		FailureReason: c.FailureReason,
		Expiry:        c.Expiry,
		Type:          string(c.Type),
	}
}

func (c *Changeset) IsDone() bool {
	var isDone bool
	for _, v := range c.Outputs {
		if v {
			isDone = true

			return isDone
		}

	}

	return isDone
}

func (c *Changeset) GetChanges() ([]types.Changeset, error) {
	result := make([]types.Changeset, 0)

	for nodeId, changes := range c.Inputs {
		for _, change := range changes {
			changeStruct, err := utils.ConvertStruct[map[string]interface{}, types.MediaItemChange](change.Change)
			if err != nil {
				return nil, err
			}

			result = append(result, types.Changeset{
				MediaItemMapping: types.MediaItemMapping{NodeId: nodeId, MediaId: change.MediaId},
				Change:           changeStruct,
			})
		}
	}

	return result, nil
}

func newChangesetFromRequest(r types.ChangesetCreateRequest) (*Changeset, error) {
	inputs := make(map[string][]input)
	outputs := make(map[string]bool)

	for _, item := range r.Changes {
		mapChange, err := utils.ConvertStruct[types.MediaItemChange, map[string]interface{}](item.Change)
		if err != nil {
			return nil, err
		}

		ip := input{
			MediaId: item.MediaId,
			Change:  mapChange,
		}
		inputs[item.NodeId] = append(inputs[item.NodeId], ip)
		outputs[item.NodeId] = false
	}

	return &Changeset{
		Id:      primitive.NewObjectID(),
		Status:  StatusPending,
		Type:    changesetType(r.Action),
		Inputs:  inputs,
		Outputs: outputs,
	}, nil
}
