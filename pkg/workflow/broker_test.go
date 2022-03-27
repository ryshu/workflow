package workflow_test

import (
	"errors"
	"log"
	"testing"

	"bou.ke/monkey"
	"git.spikeelabs.com/workflow/v1/pkg/workflow"
	"github.com/stretchr/testify/assert"
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

func TestNewBroker(t *testing.T) {
	assert := assert.New(t)

	mock := newBrokerMock()

	broker := workflow.NewBroker(mock)

	assert.False(mock.closeCalled)
	assert.False(mock.consumeCalled)
	assert.False(mock.unsafePushCalled)
	assert.False(mock.pushCalled)

	assert.IsType(workflow.Broker{}, *broker)
}

func TestBrokerPush(t *testing.T) {
	assert := assert.New(t)

	mock := newBrokerMock()

	broker := workflow.NewBroker(mock)
	broker.Push("test", []byte("string"))

	assert.False(mock.closeCalled)
	assert.False(mock.consumeCalled)
	assert.False(mock.unsafePushCalled)
	assert.True(mock.pushCalled)

	assert.IsType(workflow.Broker{}, *broker)
}

func TestBrokerUnsafePush(t *testing.T) {
	assert := assert.New(t)

	mock := newBrokerMock()

	broker := workflow.NewBroker(mock)
	broker.UnsafePush("test", []byte("string"))

	assert.False(mock.closeCalled)
	assert.False(mock.consumeCalled)
	assert.True(mock.unsafePushCalled)
	assert.False(mock.pushCalled)

	assert.IsType(workflow.Broker{}, *broker)
}

func TestBrokerClose(t *testing.T) {
	assert := assert.New(t)

	mock := newBrokerMock()

	broker := workflow.NewBroker(mock)
	broker.Close()

	assert.True(mock.closeCalled)
	assert.False(mock.consumeCalled)
	assert.False(mock.unsafePushCalled)
	assert.False(mock.pushCalled)

	assert.IsType(workflow.Broker{}, *broker)
}

func TestBrokerConsume(t *testing.T) {
	assert := assert.New(t)
	mock := newBrokerMock()
	broker := workflow.NewBroker(mock)
	store := newStorageMock()

	// Setup mocks
	callPatchPrintf := 0
	sentry := NewSentryMock()
	PatchPrintf := monkey.Patch(log.Printf, func(format string, v ...interface{}) { callPatchPrintf += 1 })

	// Basic testing Consume
	broker.Consume("test", []string{"sample"}, workflow.NewStorage(store), getMockedHandler(nil, nil))
	assert.False(mock.closeCalled)
	assert.True(mock.consumeCalled)
	assert.False(mock.unsafePushCalled)
	assert.False(mock.pushCalled)
	assert.IsType(workflow.Broker{}, *broker)

	// Testing ConfigureScope is called during handling of valid message
	sentry.PatchConfigureScope.Restore()
	mock.handler(getSimpleWorkflow().Marshal())
	assert.Equal(1, sentry.callPatchConfigureScope)
	sentry.Reset()

	// Testing ConfigureScope is called during handling of valid message
	sentry.PatchSetTag.Restore()
	mock.handler(getSimpleWorkflow().Marshal())
	assert.Equal(4, sentry.callPatchSetTag)
	sentry.Reset()

	// Testing Error while Parsing invalid json body
	callPatchPrintf = 0
	sentry.PatchCaptureException.Restore()
	mock.handler([]byte("not_json"))
	assert.Equal(1, sentry.callPatchCaptureException)
	assert.Equal(1, callPatchPrintf)
	sentry.Reset()

	// Testing Error while Validating body
	callPatchPrintf = 0
	sentry.PatchCaptureException.Restore()
	mock.handler([]byte("{\"test\": \"test\"}"))
	assert.Equal(1, sentry.callPatchCaptureException)
	assert.Equal(1, callPatchPrintf)
	sentry.Reset()

	// Testing Error on CreateStateLogIfNotExist
	store.CreateStateLogIfNotExistReturn = errors.New("Sample")
	sentry.PatchConfigureScope.Restore()
	sentry.PatchCaptureException.Restore()
	mock.handler(getSimpleWorkflow().Marshal())
	assert.Equal(1, sentry.callPatchCaptureException)
	assert.Equal(1, sentry.callPatchConfigureScope)
	sentry.Reset()
	store.CreateStateLogIfNotExistReturn = nil

	// Testing Error on IsFlowShutdown
	callPatchPrintf = 0
	store.IsFlowShutdownErrorReturn = errors.New("Sample")
	sentry.PatchConfigureScope.Restore()
	sentry.PatchCaptureException.Restore()
	mock.handler(getSimpleWorkflow().Marshal())
	assert.Equal(1, sentry.callPatchCaptureException)
	assert.Equal(1, sentry.callPatchConfigureScope)
	assert.Equal(2, callPatchPrintf) // Display processing flow + Fail to correlate flow
	sentry.Reset()
	store.IsFlowShutdownErrorReturn = nil

	// Testing IsFlowShutdown return True (flow is shutdown manually)
	callPatchPrintf = 0
	store.IsFlowShutdownBoolReturn = true
	sentry.PatchConfigureScope.Restore()
	mock.handler(getSimpleWorkflow().Marshal())
	assert.Equal(1, sentry.callPatchConfigureScope)
	assert.Equal(1, callPatchPrintf) // Display processing flow
	sentry.Reset()
	store.IsFlowShutdownBoolReturn = false

	// Testing SetMetaData for propagate without errors
	// TODO:

	// Testing SetMetaData with errors
	// TODO:

	// Testing Error return by handler
	// TODO:

	// Testing Error during Propagate
	// TODO:

	// Unpatch mocks
	PatchPrintf.Unpatch()
}
