package mongo

import (
	"git.spikeelabs.com/workflow/v1/pkg/workflow"
	"github.com/google/uuid"
)

type HistoryEntry struct {
	Entity   string                  `json:"entity" bson:"entity"`
	Status   workflow.ProgressStatus `json:"status" bson:"status"`
	State    workflow.FlowState      `json:"state" bson:"state"`
	FlowId   uuid.UUID               `json:"w_id" bson:"w_id"`
	BranchId uuid.UUID               `json:"b_id" bson:"b_id"`
	StepId   uuid.UUID               `json:"s_id" bson:"s_id"`
	Name     string                  `json:"name" bson:"name"`
}
