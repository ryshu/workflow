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
	callPatchSetContext       int

	PatchCaptureException *monkey.PatchGuard
	PatchConfigureScope   *monkey.PatchGuard
	PatchSetTag           *monkey.PatchGuard
	PatchSetContext       *monkey.PatchGuard
}

func NewSentryMock() *SentryMock {
	sentryMock := SentryMock{
		callPatchConfigureScope:   0,
		callPatchCaptureException: 0,
		callPatchSetTag:           0,
		callPatchSetContext:       0,
	}

	sentryMock.PatchCaptureException = monkey.PatchInstanceMethod(
		reflect.TypeOf(sentry.CurrentHub()), "CaptureException", func(*sentry.Hub, error) *sentry.EventID {
			sentryMock.callPatchCaptureException += 1
			return &sentry.NewEvent().EventID
		},
	)
	sentryMock.PatchSetTag = monkey.PatchInstanceMethod(
		reflect.TypeOf(sentry.CurrentHub().Scope()), "SetTag", func(*sentry.Scope, string, string) { sentryMock.callPatchSetTag += 1 },
	)
	sentryMock.PatchConfigureScope = monkey.PatchInstanceMethod(
		reflect.TypeOf(sentry.CurrentHub()), "ConfigureScope", func(*sentry.Hub, func(*sentry.Scope)) { sentryMock.callPatchConfigureScope += 1 },
	)
	sentryMock.PatchSetContext = monkey.PatchInstanceMethod(
		reflect.TypeOf(sentry.CurrentHub().Scope()), "SetContext", func(*sentry.Scope, string, interface{}) { sentryMock.callPatchSetContext += 1 },
	)

	sentryMock.PatchCaptureException.Unpatch()
	sentryMock.PatchSetTag.Unpatch()
	sentryMock.PatchConfigureScope.Unpatch()
	sentryMock.PatchSetContext.Unpatch()

	return &sentryMock
}

func (s *SentryMock) Reset() {
	s.callPatchCaptureException = 0
	s.callPatchConfigureScope = 0
	s.callPatchSetTag = 0
	s.callPatchSetContext = 0

	s.PatchCaptureException.Unpatch()
	s.PatchSetTag.Unpatch()
	s.PatchConfigureScope.Unpatch()
	s.PatchSetContext.Unpatch()
}
