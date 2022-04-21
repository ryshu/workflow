package workflow

import (
	"log"
	"time"

	"github.com/getsentry/sentry-go"
)

// BrokerInterface interface, which need to be fully implemented to integrate a
// new broker for workflows message producing and consuming.
type BrokerInterface interface {
	Push(routingKey string, data []byte) error
	UnsafePush(routingKey string, data []byte) error
	Consume(name string, keys []string, handler func(body []byte))
	Close() error

	// FIXME: Create method to test if ready for broker
	// FIXME: Create method to flush buffer when ready
	// FIXME: Create method to clear buffer
}

type Broker struct {
	broker BrokerInterface
}

// NewBroker Create new broker using given broker interface
func NewBroker(b BrokerInterface) *Broker {
	return &Broker{broker: b}
}

func (b *Broker) Push(routingKey string, data []byte) error {
	// FIXME: Create queue for push, buffer used to prevent push if not in auto-commit mode
	return b.broker.Push(routingKey, data)
}

func (b *Broker) UnsafePush(routingKey string, data []byte) error {
	return b.broker.UnsafePush(routingKey, data)
}

func (b *Broker) Consume(name string, keys []string, store *Storage, handler func(flow *Flow) (Table, error)) {
	sub_handler := func(body []byte) {
		// Parse body into flow
		flow, err := FromBroker(body)
		if err != nil {
			log.Printf("Fail to parse body into Flow, received body is %s", body)
			sentry.CaptureException(err)
			return
		}

		// Enrich Sentry context with tag, help developers for debug
		sentry.ConfigureScope(func(scope *sentry.Scope) {
			scope.SetTag("w_id", flow.FlowId.String())
			scope.SetTag("b_id", flow.BranchId.String())
			scope.SetTag("s_id", flow.StepId.String())
			scope.SetTag("step_name", flow.GetCurrentStep().Name)
		})

		flow.Progress.StartDate = time.Now()
		err = store.CreateStateLogIfNotExist(flow, flow.GetCurrentInKey(), StepPending)
		if err != nil {
			sentry.CaptureException(err)
			return
		}
		store.CreateStateLog(flow, flow.GetCurrentInKey(), StepRunning)

		log.Printf("Processing flow %s", flow.FlowId)
		res, err := Correlate(flow, store, b)
		if err != nil {
			log.Printf("Fail to correlate flow, %s", err.Error())
			sentry.CaptureException(err)
			return
		}
		if !res {
			// If correlate return false, skip handling and propagate
			return
		}

		table, err := handler(flow)
		if table != nil {
			err_set := flow.SetMetadata(table)
			if err_set != nil {
				log.Printf("Fail to update metadata of the flow, %s", err_set.Error())
				sentry.CaptureException(err_set)
				return
			}
		}
		if err != nil {
			// Handle handler error logging to Sentry
			sentry.ConfigureScope(func(scope *sentry.Scope) {
				scope.SetContext("current_step", flow.GetCurrentStep().Marshal())
				scope.SetContext("metadata", flow.Metadata)
			})
			sentry.CaptureException(err)
			Fail(flow, err, store, b)
			return
		}

		// Try to Propagate message at end
		if err := Propagate(flow, StepSuccess, store, b); err != nil {
			// If propagate fail, handle it as an handler error
			sentry.ConfigureScope(func(scope *sentry.Scope) {
				scope.SetContext("current_step", flow.GetCurrentStep().Marshal())
				scope.SetContext("metadata", flow.Metadata)
			})
			sentry.CaptureException(err)
			Fail(flow, err, store, b)
			return
		}
	}

	b.broker.Consume(name, keys, sub_handler)
}

func (b *Broker) Close() error {
	// FIXME: Flush before closing, if not force close
	return b.broker.Close()
}
