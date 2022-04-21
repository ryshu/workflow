package workflow

import "time"

func propagateStore(flow *Flow, status ProgressStatus, store *Storage) error {
	flow.AssignBranchId()
	return store.CreateStateLog(flow, flow.GetCurrentOutKey(), status)
}

func handlePattern(flow *Flow, action *StepOut, store *Storage, broker *Broker) error {
	switch action.Pattern {
	case PassthroughPatternOut:
		currentInKey := flow.GetCurrentInKey()
		flow.AssignStepId()

		// Insert pending message statelog
		err := store.CreateStateLog(flow, currentInKey, StepPending)
		if err != nil {
			return err
		}

		// Handle case when we don't want to propagate msg to broker
		if broker != nil {
			return broker.Push(currentInKey, flow.Marshal())
		} else {
			return nil
		}
	case SplitPattern:
		// TODO: Implement expender hook and split pattern
		return nil
	default:
		return nil
	}
}

// handleOutActions is used to handle out actions of the workflow
func handleOutActions(flow *Flow, actions []StepOut, store *Storage, broker *Broker) error {
	if len(actions) > 0 {
		// FIXME: Is this function safe for multi-actions ?
		for _, action := range actions {
			// Change flow current step to next step
			clone := flow.DeepCopy()
			clone.Progress = ProgressDetail{
				Name:       clone.Name,
				CreateDate: time.Now(),
				Status:     StepPending,
			}
			clone.CurrentStep = action.NextStep

			// If record is defined on action, trigger it on storage
			if action.Record != (HistoryRecord{}) {
				err := store.StoreHistoryEntry(clone, &action.Record)
				if err != nil {
					return err
				}
			}

			// If enrich pattern might be called, call them on hooks
			if len(action.Enrich) > 0 {
				// TODO: Enrich pattern here using hook, if declared
			}

			if clone.CurrentStep != 0 {
				err := handlePattern(clone, &action, store, broker)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func Propagate(flow *Flow, status ProgressStatus, store *Storage, broker *Broker) error {
	// Propage to store end of workflow
	err := propagateStore(flow, status, store)
	if err != nil {
		return err
	}

	// Retrieve and handle out actions from workflow
	currentStep := flow.GetCurrentStep()
	switch status {
	case StepSuccess:
		return handleOutActions(flow, currentStep.Out, store, broker)
	case StepFailure:
		return handleOutActions(flow, currentStep.Fail, store, broker)
	default:
		return nil
	}
}

// ResumeStep is used to resume a step in case of failure (restart of broker for example)
func ResumeStep(flowId string, flowName string, stepId string, store *Storage, broker *Broker) (bool, error) {
	// Call to storage to retrieve resumable step from flow and step identifiers
	flow, err_storage := store.GetResumableStep(flowId, stepId)
	if err_storage != nil {
		return false, err_storage
	}

	// If no flow retrieved, return false for resume without error
	if flow == nil {
		return false, nil
	}

	// Push message to broker (don't register a new flow, flow is already registered)
	err_broker := broker.Push(flow.GetCurrentOutKey(), flow.Marshal())
	if err_broker != nil {
		return false, err_broker
	}

	// Add entry in replay log for history
	// FIXME: Setup this entry

	return true, nil
}

func Fail(flow *Flow, err error, store *Storage, broker *Broker) {
	// FIXME: Setup error in flow before propagate

	err_propagate := Propagate(flow, StepFailure, store, broker)
	if err_propagate != nil {
		panic(err_propagate)
	}
}

func Aggregate(flow *Flow, action *StepIn, store *Storage) (bool, error) {
	if flow.CorrelationChain != nil && len(flow.CorrelationChain) > 0 {
		// Truncate CorrelationChain to apply action.Depth as max length of chain
		Chain := flow.CorrelationChain[:min(len(flow.CorrelationChain), action.Depth)]

		check, err := store.IsChainsSuccessfull(flow, Chain)
		if err != nil {
			return false, err
		}

		if check {
			flow.JoinResult = Joined
			flow.CorrelationChain = flow.CorrelationChain[1:]
			return true, nil
		}

		flow.JoinResult = NotJoined
		return false, nil
	}
	flow.JoinResult = Joined
	return true, nil
}

func Correlate(flow *Flow, store *Storage, broker *Broker) (bool, error) {
	// Test if current flow is shutdown, and skip corralate if so
	result, err_is_shutdown := store.IsFlowShutdown(flow)
	if err_is_shutdown != nil {
		return false, err_is_shutdown
	}
	if result {
		return false, nil
	}

	// AutoPropagate loop to propagate all notification which need to be autopropagated without correlation
	// and handling by a callback.
	for {
		step := flow.GetCurrentStep()
		if step.AutoPropagate {
			err_propagate := Propagate(flow, StepSuccess, store, nil)
			if err_propagate != nil {
				return false, err_propagate
			}

		} else {
			break
		}
	}

	step := flow.GetCurrentStep()
	if step.In != nil && len(step.In) > 0 {
		// Validate each prerequisite from in array
		check := true
		for _, action := range step.In {
			switch action.Pattern {
			case PassthroughPatternIn:
				check = true
			case AggregatePattern:
				result, err := Aggregate(flow, &action, store)
				if err != nil {
					return false, err
				}
				if !result {
					check = false
				}
			default:
				check = false
			}
		}

		// If a check fail, drop message
		if !check {
			clone := flow.DeepCopy()
			clone.AssignBranchId()

			err := store.CreateStateLog(clone, flow.GetCurrentInKey(), StepDropped)
			if err != nil {
				return false, err
			}
			return false, nil
		}
	}
	return true, nil
}
