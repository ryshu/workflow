package workflow_test

import "git.spikeelabs.com/workflow/v1/pkg/workflow"

type StorageMock struct {
	getResumableStepCalled         bool
	createStateLogIfNotExistCalled bool
	createStateLogCalled           bool
	storeHistoryEntryCalled        bool
	isChainsSuccessfullCalled      bool
	isFlowShutdownCalled           bool

	CreateStateLogIfNotExistReturn error
	CreateStateLogReturn           error
	IsFlowShutdownErrorReturn      error
	IsFlowShutdownBoolReturn       bool
}

func (s *StorageMock) GetResumableStep(flowId string, stepId string) (*workflow.Flow, error) {
	s.getResumableStepCalled = true
	return nil, nil
}

func (s *StorageMock) CreateStateLogIfNotExist(flow *workflow.Flow, routingKey string, status workflow.ProgressStatus) error {
	s.createStateLogIfNotExistCalled = true
	return s.CreateStateLogIfNotExistReturn
}

func (s *StorageMock) CreateStateLog(flow *workflow.Flow) error {
	s.createStateLogCalled = true
	return s.CreateStateLogReturn
}

func (s *StorageMock) StoreHistoryEntry(flow *workflow.Flow, record *workflow.HistoryRecord) error {
	s.storeHistoryEntryCalled = true
	return nil
}

func (s *StorageMock) IsChainsSuccessfull(flow *workflow.Flow, ChainIds []string) (bool, error) {
	s.isChainsSuccessfullCalled = true
	return true, nil
}

func (s *StorageMock) IsFlowShutdown(flow *workflow.Flow) (bool, error) {
	return s.IsFlowShutdownBoolReturn, s.IsFlowShutdownErrorReturn
}

func newStorageMock() *StorageMock {
	return &StorageMock{
		getResumableStepCalled:         false,
		createStateLogIfNotExistCalled: false,
		createStateLogCalled:           false,
		storeHistoryEntryCalled:        false,
		isChainsSuccessfullCalled:      false,
		isFlowShutdownCalled:           false,

		CreateStateLogIfNotExistReturn: nil,
		IsFlowShutdownErrorReturn:      nil,
		IsFlowShutdownBoolReturn:       false,
	}
}
