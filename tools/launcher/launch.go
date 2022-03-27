package main

import (
	"context"
	"fmt"
	"time"

	"git.spikeelabs.com/workflow/v1/pkg/workflow"
	"git.spikeelabs.com/workflow/v1/pkg/workflow/amqp"
)

func main() {
	// Setup broker
	amqpInstance := amqp.New("amq.topic", "amqp://workflow:workflow@localhost:5672/")
	broker := workflow.NewBroker(amqpInstance)

	// Setup sample msg
	steps := []workflow.Step{
		{
			Name: "test_step_1",
			In:   []workflow.StepIn{{Pattern: workflow.PassthroughPatternIn}},
			Out: []workflow.StepOut{{
				Pattern:  workflow.PassthroughPatternOut,
				NextStep: 1,
				Record:   workflow.HistoryRecord{Entity: "test_entity", Status: workflow.StepRunning},
			}},
			Fail: []workflow.StepOut{{
				Pattern:  workflow.PassthroughPatternOut,
				NextStep: 1,
				Record:   workflow.HistoryRecord{Entity: "test_entity", Status: workflow.StepFailure},
			}},
		},
		{
			Name: "test_step_2",
			In:   []workflow.StepIn{{Pattern: workflow.PassthroughPatternIn}},
			Out: []workflow.StepOut{{
				Pattern: workflow.PassthroughPatternOut,
				Record:  workflow.HistoryRecord{Entity: "test_entity", Status: workflow.StepSuccess},
			}},
			Fail: []workflow.StepOut{{
				Pattern: workflow.PassthroughPatternOut,
				Record:  workflow.HistoryRecord{Entity: "test_entity", Status: workflow.StepFailure},
			}},
		},
	}

	flow := workflow.NewFlow("test", steps)

	// Wait until broker is ready or timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	for {
		if amqpInstance.IsReady || ctx.Err() != nil {
			break
		}
	}

	// Publish message
	if err := broker.Push(flow.GetCurrentInKey(), flow.Marshal()); err != nil {
		fmt.Printf("Push failed: %s\n", err)
	} else {
		fmt.Printf("Push succeeded ! %s\n", flow.GetCurrentInKey())
	}
}
