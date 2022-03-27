package mongo

import (
	"context"

	"git.spikeelabs.com/workflow/v1/pkg/workflow"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// StoreHistoryEntry add entry to the entity ref dict
func (session *Session) StoreHistoryEntry(flow *workflow.Flow, record *workflow.HistoryRecord) error {
	historyEntry := HistoryEntry{
		Entity:   record.Entity,
		Status:   record.Status,
		State:    workflow.StateStarted,
		FlowId:   flow.FlowId,
		StepId:   flow.StepId,
		BranchId: flow.BranchId,
		Name:     flow.Name,
	}

	client, err := session.GetMongoClient()
	if err != nil {
		return err
	}
	collection := client.Database(session.database).Collection(session.historyColl)

	_, err = collection.InsertOne(context.TODO(), historyEntry)
	if err != nil {
		return err
	}

	return nil
}

func (session *Session) IsFlowShutdown(flow *workflow.Flow) (bool, error) {
	collection, err := session.GetHistoryColl()
	if err != nil {
		return false, err
	}

	// Filter by Flow ID, order by desc on _id (retrieve last in first)
	filter := bson.D{primitive.E{Key: "w_id", Value: flow.FlowId}}
	options := options.FindOneOptions{Sort: bson.M{"ts": -1}, Projection: bson.M{"state": 1}}

	var result bson.M
	err = collection.FindOne(context.TODO(), filter, &options).Decode(&result)
	if err != nil {
		// ErrNoDocuments means the workflow isn't registered, so cannot be shutdowned.
		if err == mongo.ErrNoDocuments {
			return false, nil
		}
		return false, err
	}

	if val, ok := result["state"]; ok {
		return val == workflow.StateShutdowned, nil
	}
	return false, nil
}
