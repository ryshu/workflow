package workflow_test

import "git.spikeelabs.com/workflow/v1/pkg/workflow"

func getSimpleWorkflow() *workflow.Flow {
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

	return workflow.NewFlow("test", steps)
}

func getMockedBroker() *workflow.Broker {
	return workflow.NewBroker(newBrokerMock())
}

func getMockerStore() *workflow.Storage {
	return workflow.NewStorage(newStorageMock())
}
