package workflow

import (
	"time"
)

// FIXME: Setup flush and rollback options for store to prevent write on errors

// StorageInterface interface, which need to be fully implemented to integrate a
// new way of storing workflows, their states as well as their progress histories.
type StorageInterface interface {
	GetResumableStep(flowId string, stepId string) (*Flow, error)
	CreateStateLogIfNotExist(flow *Flow, routingKey string, status ProgressStatus) error
	CreateStateLog(flow *Flow) error
	StoreHistoryEntry(flow *Flow, record *HistoryRecord) error
	IsChainsSuccessfull(flow *Flow, ChainIds []string) (bool, error)
	IsFlowShutdown(flow *Flow) (bool, error)
}

type Storage struct {
	store StorageInterface
}

// NewStorage Create new storage using given storage interface
func NewStorage(s StorageInterface) *Storage {
	return &Storage{store: s}
}

// GetResumableStep is used to get resumable step, using flow and step identifier
func (s *Storage) GetResumableStep(flowId string, stepId string) (*Flow, error) {
	return s.store.GetResumableStep(flowId, stepId)
}

// CreateStateLogIfNotExist is used to check if a statelog exist or create it.
func (s *Storage) CreateStateLogIfNotExist(flow *Flow, routingKey string, status ProgressStatus) error {
	return s.store.CreateStateLogIfNotExist(flow, routingKey, status)
}

// CreateStateLog create a statelog using flow, routingKey and progress status
func (s *Storage) CreateStateLog(flow *Flow, routingKey string, status ProgressStatus) error {
	// Set progress details for a end of workflow
	flow.Progress.Status = status
	flow.Progress.RoutingKey = routingKey
	if status != StepPending && status != StepRunning {
		flow.Progress.EndDate = time.Now()
		flow.Progress.Duration = flow.Progress.EndDate.Sub(flow.Progress.StartDate)
		flow.Progress.DurationStr = flow.Progress.Duration.String()
	}

	return s.store.CreateStateLog(flow)
}

// StoreHistoryEntry store a new history entry at creation and end of workflow. Used to flow workflow progress
func (s *Storage) StoreHistoryEntry(flow *Flow, record *HistoryRecord) error {
	return s.store.StoreHistoryEntry(flow, record)
}

// IsChainsSuccessfull check is the given chain of step is successfull
func (s *Storage) IsChainsSuccessfull(flow *Flow, ChainIds []string) (bool, error) {
	return s.store.IsChainsSuccessfull(flow, ChainIds)
}

func (s *Storage) IsFlowShutdown(flow *Flow) (bool, error) {
	return s.store.IsFlowShutdown(flow)
}
