package workflow_test

import (
	"errors"

	"git.spikeelabs.com/workflow/v1/pkg/workflow"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

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

func getSimpleWorkflowWithMetadata() *workflow.Flow {
	flow := getSimpleWorkflow()
	flow.Metadata = workflow.Table{"test": "test"}
	return flow
}

var _ = Describe("Utils", func() {
	It("FailOnError panic", func() {
		Expect(func() { workflow.FailOnError(errors.New("Sample"), "panic") }).To(Panic())
	})
})
