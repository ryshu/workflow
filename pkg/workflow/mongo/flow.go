package mongo

import (
	"context"

	"git.spikeelabs.com/workflow/v1/pkg/workflow"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func (session *Session) GetResumableStep(flowId string, stepId string) (*workflow.Flow, error) {
	result := workflow.Flow{}

	collection, err := session.GetHistoryColl()
	if err != nil {
		return nil, err
	}

	// Filter by Flow ID and Step ID, order by desc on TS (timestamp)
	filter := bson.D{primitive.E{Key: "w_id", Value: flowId}, primitive.E{Key: "s_id", Value: stepId}}
	options := options.FindOneOptions{Sort: bson.M{"ts": -1}}

	err = collection.FindOne(context.TODO(), filter, &options).Decode(&result)
	return &result, err
}

func (session *Session) CreateStateLogIfNotExist(flow *workflow.Flow, routingKey string, status workflow.ProgressStatus) error {
	collection, err := session.GetFlowColl()
	if err != nil {
		return err
	}

	// Filter by Flow ID, order by desc on _id (retrieve last in first)
	filter := bson.D{
		primitive.E{Key: "w_id", Value: flow.FlowId},
		primitive.E{Key: "s_id", Value: flow.StepId},
		primitive.E{Key: "progress.status", Value: status},
	}

	var result bson.M
	err = collection.FindOne(context.TODO(), filter, &options.FindOneOptions{}).Decode(&result)
	if err != nil {
		if err != mongo.ErrNoDocuments {
			return err
		} else {
			clone := flow.DeepCopy()
			clone.AssignBranchId()
			clone.Progress.Status = status
			session.CreateStateLog(clone)
		}
	}
	return nil
}

func (session *Session) CreateStateLog(flow *workflow.Flow) error {
	collection, err := session.GetFlowColl()
	if err != nil {
		return err
	}

	_, err = collection.InsertOne(context.TODO(), flow)
	return err
}

func (session *Session) IsChainsSuccessfull(flow *workflow.Flow, ChainIds []string) (bool, error) {
	coll, err := session.GetFlowColl()
	if err != nil {
		return false, err
	}

	ret := false

	for _, ChainId := range ChainIds {
		// FIXME: Find a way to do it in One pipeline, not X
		// PERFS: Check performances here on heavy workflows
		pipeline := mongo.Pipeline{
			bson.D{
				primitive.E{Key: "$match", Value: bson.M{
					"w_id": flow.FlowId,
					"s_id": bson.M{"$neq": flow.StepId},
					"$expr": bson.M{
						"$in": bson.A{ChainId, "$c_chain"},
					},
				}},
				primitive.E{Key: "$sort", Value: bson.M{"ts": -1}},
				primitive.E{Key: "$group", Value: bson.M{
					"_id":    "s_id",
					"status": bson.M{"$first": "status"},
				}},
			},
		}

		cursor, err := coll.Aggregate(context.TODO(), pipeline)
		if err != nil {
			return false, err
		}

		var results []bson.M
		if err = cursor.All(context.TODO(), &results); err != nil {
			return false, err
		}

		// FIXME: Find a way to get result from pipeline (computed during pipeline, not after)
		// FIXME: Maybe log problem also for DEBUG
		for _, result := range results {
			if val, ok := result["status"]; ok {
				if val == workflow.StepDropped || val == workflow.StepSuccess {
					ret = true
				} else {
					return false, nil
				}
			}
		}
		if ret == false {
			return false, nil
		}
	}

	return ret, nil
}
