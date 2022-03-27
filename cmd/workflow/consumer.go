package main

import (
	"context"
	"time"

	"git.spikeelabs.com/workflow/v1/pkg/workflow"
	"git.spikeelabs.com/workflow/v1/pkg/workflow/amqp"
	"git.spikeelabs.com/workflow/v1/pkg/workflow/mongo"
)

func main() {
	// Setup broker
	amqpInstance := amqp.New("amq.topic", "amqp://workflow:workflow@localhost:5672/")
	broker := workflow.NewBroker(amqpInstance)

	// Setup MongoDB connection
	mongoInstance := mongo.New("workflow_db", "mongodb://workflow:workflow@localhost:27017", "flow", "history")
	storage := workflow.NewStorage(mongoInstance)

	// Consume on callback
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	for {
		if amqpInstance.IsReady || ctx.Err() != nil {
			break
		}
	}
	broker.Consume("test_queue", []string{"flow.test.test_step_1.start", "flow.test.test_step_2.start"}, storage, handler)
}

func handler(flow *workflow.Flow) (workflow.Table, error) {
	return nil, nil
}
