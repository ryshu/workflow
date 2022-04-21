package workflow

import (
	"fmt"
	"time"
)

type JoinResult string

const (
	Joined    JoinResult = "joined"
	NotJoined JoinResult = "not-joined"
)

type FlowState string

const (
	StateStarted    FlowState = "started"
	StateShutdowned FlowState = "shutdowned"
)

type ProgressStatus string

const (
	StepSuccess ProgressStatus = "success"
	StepFailure ProgressStatus = "failure"
	StepPending ProgressStatus = "pending"
	StepRunning ProgressStatus = "running"
	StepDropped ProgressStatus = "dropped"
)

type ProgressError string

const (
	StepErrUnknown ProgressError = "err-unknown"
)

type PatternIn string

const (
	AggregatePattern     PatternIn = "aggregate-pattern-in"
	PassthroughPatternIn PatternIn = "passthrough-pattern-in"
)

type PatternOut string

const (
	SplitPattern          PatternOut = "split-pattern-out"
	PassthroughPatternOut PatternOut = "passthrough-pattern-out"
)

// Table stores user supplied fields of the following types:
//
//   bool
//   byte
//   float32
//   float64
//   int
//   int16
//   int32
//   int64
//   nil
//   string
//   time.Time
//   amqp.Decimal
//   amqp.Table
//   []byte
//
// Functions taking a table will immediately fail when the table contains a
// value of an unsupported type.
//
// The caller must be specific in which precision of integer it wishes to
// encode.
//
// Use a type assertion when reading values from a table for type conversion.
//
type Table map[string]interface{}

func validateField(f interface{}) error {
	switch f.(type) {
	case nil, bool, byte, int, int16, int32, int64, float32, float64, string, []byte, time.Time:
		return nil
	}

	return fmt.Errorf("value %T not supported", f)
}

// Validate returns and error if any Go types in the table are incompatible with Table allowed types.
func (t Table) Validate() error {
	for _, val := range t {
		ret := validateField(val)
		if ret != nil {
			return ret
		}
	}
	return nil
}
