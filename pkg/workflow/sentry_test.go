package workflow_test

import (
	"reflect"

	"bou.ke/monkey"
	"github.com/getsentry/sentry-go"
)

type SentryMock struct {
	callPatchConfigureScope   int
	callPatchCaptureException int
	callPatchSetTag           int

	PatchCaptureException *monkey.PatchGuard
	PatchConfigureScope   *monkey.PatchGuard
	PatchSetTag           *monkey.PatchGuard
}

func NewSentryMock() *SentryMock {
	sentryMock := SentryMock{
		callPatchConfigureScope:   0,
		callPatchCaptureException: 0,
		callPatchSetTag:           0,
	}

	sentryMock.PatchCaptureException = monkey.Patch(
		sentry.CaptureException,
		func(exception error) *sentry.EventID { sentryMock.callPatchCaptureException += 1; return nil },
	)
	sentryMock.PatchSetTag = monkey.PatchInstanceMethod(
		reflect.TypeOf(sentry.CurrentHub().Scope()), "SetTag", func(*sentry.Scope, string, string) { sentryMock.callPatchSetTag += 1 },
	)
	sentryMock.PatchConfigureScope = monkey.Patch(
		sentry.ConfigureScope,
		func(callback func(scope *sentry.Scope)) { sentryMock.callPatchConfigureScope += 1 },
	)

	sentryMock.PatchCaptureException.Unpatch()
	sentryMock.PatchSetTag.Unpatch()
	sentryMock.PatchConfigureScope.Unpatch()

	return &sentryMock
}

func (s *SentryMock) Reset() {
	s.callPatchCaptureException = 0
	s.callPatchConfigureScope = 0
	s.callPatchSetTag = 0

	s.PatchCaptureException.Unpatch()
	s.PatchSetTag.Unpatch()
	s.PatchConfigureScope.Unpatch()
}
