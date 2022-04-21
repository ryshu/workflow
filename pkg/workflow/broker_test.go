package workflow_test

import (
	"errors"
	"log"

	"bou.ke/monkey"
	"git.spikeelabs.com/workflow/v1/pkg/workflow"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

type BrokerMock struct {
	pushCalled       bool
	unsafePushCalled bool
	consumeCalled    bool
	closeCalled      bool

	handler func(body []byte)
}

func (b *BrokerMock) Push(routingKey string, data []byte) error {
	b.pushCalled = true
	return nil
}

func (b *BrokerMock) UnsafePush(routingKey string, data []byte) error {
	b.unsafePushCalled = true
	return nil
}

func (b *BrokerMock) Consume(name string, keys []string, handler func(body []byte)) {
	b.handler = handler
	b.consumeCalled = true
}

func (b *BrokerMock) Close() error {
	b.closeCalled = true
	return nil
}

func newBrokerMock() *BrokerMock {
	return &BrokerMock{
		pushCalled:       false,
		unsafePushCalled: false,
		consumeCalled:    false,
		closeCalled:      false,
	}
}

func getMockedHandler(table workflow.Table, err error) func(flow *workflow.Flow) (workflow.Table, error) {
	return func(flow *workflow.Flow) (workflow.Table, error) {
		return table, err
	}
}

var _ = Describe("Broker", func() {
	Describe("Default connector utils", func() {
		It("Check new broker creation", func() {
			mock := newBrokerMock()

			broker := workflow.NewBroker(mock)

			Expect(mock.closeCalled).To(BeFalse())
			Expect(mock.consumeCalled).To(BeFalse())
			Expect(mock.unsafePushCalled).To(BeFalse())
			Expect(mock.pushCalled).To(BeFalse())
			Expect(*broker).To(BeAssignableToTypeOf(workflow.Broker{}))
		})

		It("Check that we can push to broker (safe push)", func() {
			mock := newBrokerMock()

			broker := workflow.NewBroker(mock)
			broker.Push("test", []byte("string"))

			Expect(mock.closeCalled).To(BeFalse())
			Expect(mock.consumeCalled).To(BeFalse())
			Expect(mock.unsafePushCalled).To(BeFalse())
			Expect(mock.pushCalled).To(BeTrue())
			Expect(*broker).To(BeAssignableToTypeOf(workflow.Broker{}))
		})

		It("Check that we can push to broker (unsafe push)", func() {
			mock := newBrokerMock()

			broker := workflow.NewBroker(mock)
			broker.UnsafePush("test", []byte("string"))

			Expect(mock.closeCalled).To(BeFalse())
			Expect(mock.consumeCalled).To(BeFalse())
			Expect(mock.unsafePushCalled).To(BeTrue())
			Expect(mock.pushCalled).To(BeFalse())
			Expect(*broker).To(BeAssignableToTypeOf(workflow.Broker{}))
		})

		It("Check that we can close the broker", func() {
			mock := newBrokerMock()

			broker := workflow.NewBroker(mock)
			broker.Close()

			Expect(mock.closeCalled).To(BeTrue())
			Expect(mock.consumeCalled).To(BeFalse())
			Expect(mock.unsafePushCalled).To(BeFalse())
			Expect(mock.pushCalled).To(BeFalse())
			Expect(*broker).To(BeAssignableToTypeOf(workflow.Broker{}))
		})
	})

	Describe("Consume handler", func() {
		var mock *BrokerMock
		var broker *workflow.Broker
		var store *StorageMock
		var callPatchPrintf int
		var sentryMock *SentryMock
		var PatchPrintf *monkey.PatchGuard

		BeforeEach(func() {
			mock = newBrokerMock()
			broker = workflow.NewBroker(mock)
			store = newStorageMock()

			// Setup mocks
			callPatchPrintf = 0
			sentryMock = NewSentryMock()
			PatchPrintf = monkey.Patch(log.Printf, func(format string, v ...interface{}) { callPatchPrintf += 1 })
		})

		Context("without metadata or errors in handler return", func() {
			BeforeEach(func() {
				broker.Consume("test", []string{"sample"}, workflow.NewStorage(store), getMockedHandler(nil, nil))
			})

			It("Testing ConfigureScope is called during handling of valid message", func() {
				sentryMock.PatchConfigureScope.Restore()
				mock.handler(getSimpleWorkflow().Marshal())
				Expect(sentryMock.callPatchConfigureScope).To(Equal(1))
			})

			It("Testing ConfigureScope is called during handling of valid message", func() {
				sentryMock.PatchSetTag.Restore()
				mock.handler(getSimpleWorkflow().Marshal())
				Expect(sentryMock.callPatchSetTag).To(Equal(4))
			})

			It("Testing Error while Parsing invalid json body", func() {
				callPatchPrintf = 0
				sentryMock.PatchCaptureException.Restore()
				mock.handler([]byte("not_json"))
				Expect(sentryMock.callPatchCaptureException).To(Equal(1))
				Expect(callPatchPrintf).To(Equal(1))
			})

			It("Testing Error while Validating body", func() {
				callPatchPrintf = 0
				sentryMock.PatchCaptureException.Restore()
				mock.handler([]byte("{\"test\": \"test\"}"))
				Expect(sentryMock.callPatchCaptureException).To(Equal(1))
				Expect(callPatchPrintf).To(Equal(1))
			})

			It("Testing Error on CreateStateLogIfNotExist", func() {
				store.CreateStateLogIfNotExistReturn = errors.New("Sample")
				sentryMock.PatchConfigureScope.Restore()
				sentryMock.PatchCaptureException.Restore()
				mock.handler(getSimpleWorkflow().Marshal())
				Expect(sentryMock.callPatchCaptureException).To(Equal(1))
				Expect(sentryMock.callPatchConfigureScope).To(Equal(1))
				store.CreateStateLogIfNotExistReturn = nil
			})

			It("Testing Error on IsFlowShutdown", func() {
				callPatchPrintf = 0
				store.IsFlowShutdownErrorReturn = errors.New("Sample")
				sentryMock.PatchConfigureScope.Restore()
				sentryMock.PatchCaptureException.Restore()
				mock.handler(getSimpleWorkflow().Marshal())
				Expect(sentryMock.callPatchCaptureException).To(Equal(1))
				Expect(sentryMock.callPatchConfigureScope).To(Equal(1))
				Expect(callPatchPrintf).To(Equal(2))
				store.IsFlowShutdownErrorReturn = nil
			})

			It("Testing IsFlowShutdown return True (flow is shutdown manually)", func() {
				callPatchPrintf = 0
				store.IsFlowShutdownBoolReturn = true
				sentryMock.PatchConfigureScope.Restore()
				mock.handler(getSimpleWorkflow().Marshal())
				Expect(sentryMock.callPatchConfigureScope).To(Equal(1))
				Expect(callPatchPrintf).To(Equal(1))
				store.IsFlowShutdownBoolReturn = false
			})

			It("Testing SetMetaData for propagate without errors", func() {
				broker.Consume("test", []string{"sample"}, workflow.NewStorage(store), getMockedHandler(workflow.Table{"test": "test"}, nil))
				callPatchPrintf = 0
				store.IsFlowShutdownBoolReturn = true
				sentryMock.PatchConfigureScope.Restore()
				mock.handler(getSimpleWorkflow().Marshal())
				Expect(sentryMock.callPatchConfigureScope).To(Equal(1))
				Expect(callPatchPrintf).To(Equal(1))
				store.IsFlowShutdownBoolReturn = false
			})

			It("Testing when propagate fail (Sentry Configure Scope)", func() {
				broker.Consume("test", []string{"sample"}, workflow.NewStorage(store), getMockedHandler(workflow.Table{"test": "test"}, nil))
				callPatchPrintf = 0
				store.CreateStateLogReturn = errors.New("Sample")
				sentryMock.PatchCaptureException.Restore()
				sentryMock.PatchConfigureScope.Restore()

				Expect(func() { mock.handler(getSimpleWorkflow().Marshal()) }).To(Panic())
				Expect(sentryMock.callPatchCaptureException).To(Equal(1))
				Expect(sentryMock.callPatchConfigureScope).To(Equal(2))
				Expect(callPatchPrintf).To(Equal(1))
				store.IsFlowShutdownBoolReturn = false
			})

			It("Testing when propagate fail (Sentry Set Context)", func() {
				broker.Consume("test", []string{"sample"}, workflow.NewStorage(store), getMockedHandler(workflow.Table{"test": "test"}, nil))
				callPatchPrintf = 0
				store.CreateStateLogReturn = errors.New("Sample")
				sentryMock.PatchCaptureException.Restore()
				sentryMock.PatchSetContext.Restore()

				Expect(func() { mock.handler(getSimpleWorkflow().Marshal()) }).To(Panic())
				Expect(sentryMock.callPatchCaptureException).To(Equal(1))
				Expect(sentryMock.callPatchSetContext).To(Equal(2))
				Expect(callPatchPrintf).To(Equal(1))
				store.IsFlowShutdownBoolReturn = false
			})
		})

		Context("with metadata in handler return", func() {
			BeforeEach(func() {
				broker.Consume("test", []string{"sample"}, workflow.NewStorage(store), getMockedHandler(workflow.Table{"test": "test"}, nil))
			})

			It("SetMetadata is called and success to set (no metadata before)", func() {
				// Testing ConfigureScope is called during handling of valid message
				callPatchPrintf = 0
				sentryMock.PatchConfigureScope.Restore()
				sentryMock.PatchCaptureException.Restore()
				mock.handler(getSimpleWorkflow().Marshal())
				Expect(sentryMock.callPatchConfigureScope).To(Equal(1))
				Expect(sentryMock.callPatchCaptureException).To(Equal(0))
				Expect(callPatchPrintf).To(Equal(1))
			})

			It("SetMetadata is called and success to set (metadata before)", func() {
				callPatchPrintf = 0
				sentryMock.PatchConfigureScope.Restore()
				sentryMock.PatchCaptureException.Restore()
				mock.handler(getSimpleWorkflowWithMetadata().Marshal())
				Expect(sentryMock.callPatchConfigureScope).To(Equal(1))
				Expect(sentryMock.callPatchCaptureException).To(Equal(0))
				Expect(callPatchPrintf).To(Equal(1))
			})
		})

		Context("with invalid metadata in handler return", func() {
			BeforeEach(func() {
				broker.Consume(
					"test",
					[]string{"sample"},
					workflow.NewStorage(store),
					getMockedHandler(workflow.Table{"test": workflow.Table{"test": "test"}}, nil),
				)
			})

			It("SetMetaData is called and fail to validate returned Table", func() {
				sentryMock.PatchConfigureScope.Restore()
				sentryMock.PatchCaptureException.Restore()
				mock.handler(getSimpleWorkflow().Marshal())
				Expect(sentryMock.callPatchConfigureScope).To(Equal(1))
				Expect(sentryMock.callPatchCaptureException).To(Equal(1))
				Expect(callPatchPrintf).To(Equal(2))
			})
		})

		Context("with error in handler return", func() {
			BeforeEach(func() {
				broker.Consume(
					"test",
					[]string{"sample"},
					workflow.NewStorage(store),
					getMockedHandler(nil, errors.New("Sample")),
				)
			})

			It("Sentry and Fail are called (check Configure Scope)", func() {
				sentryMock.PatchConfigureScope.Restore()
				sentryMock.PatchCaptureException.Restore()
				mock.handler(getSimpleWorkflow().Marshal())
				Expect(sentryMock.callPatchConfigureScope).To(Equal(2))
				Expect(sentryMock.callPatchCaptureException).To(Equal(1))
				Expect(callPatchPrintf).To(Equal(1))
			})

			It("Sentry and Fail are called (check Set Context)", func() {
				sentryMock.PatchSetContext.Restore()
				sentryMock.PatchCaptureException.Restore()
				mock.handler(getSimpleWorkflow().Marshal())
				Expect(sentryMock.callPatchSetContext).To(Equal(2))
				Expect(sentryMock.callPatchCaptureException).To(Equal(1))
				Expect(callPatchPrintf).To(Equal(1))
			})
		})

		AfterEach(func() {
			sentryMock.Reset()
			PatchPrintf.Unpatch()
		})
	})
})
