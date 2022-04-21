package workflow_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Flow - Panic checks", func() {
	Describe("set UUID on workflow", func() {
		It("AssignFlowId", func() {
			flow := getSimpleWorkflow()

			previousId := flow.FlowId
			flow.AssignFlowId()
			Expect(flow.FlowId).NotTo(Equal(previousId))
		})
		It("AssignBranchId", func() {
			flow := getSimpleWorkflow()

			previousId := flow.BranchId
			flow.AssignBranchId()
			Expect(flow.BranchId).NotTo(Equal(previousId))
		})
		It("AssignStepId", func() {
			flow := getSimpleWorkflow()

			previousId := flow.StepId
			flow.AssignStepId()
			Expect(flow.StepId).NotTo(Equal(previousId))
		})
	})
})
