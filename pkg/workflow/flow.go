package workflow

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

// TODO: Setup real documentation about each struct, usage etc ...

type HistoryRecord struct {
	Entity string         `json:"entity" bson:"entity"`
	Status ProgressStatus `json:"status" bson:"status"`
}

// ProgressDetail is the definition of details about the step and his processing
type ProgressDetail struct {
	Name        string         `json:"name" bson:"name"`
	CreateDate  time.Time      `json:"create_date" bson:"create_date"`
	StartDate   time.Time      `json:"start_date" bson:"start_date,omitempty"`
	EndDate     time.Time      `json:"end_date" bson:"end_date,omitempty"`
	Status      ProgressStatus `json:"status" bson:"status"`
	ErrorCode   ProgressError  `json:"error_code" bson:"error_code"`
	EntityRef   string         `json:"entity_ref" bson:"entity_ref"`
	Duration    time.Duration  `json:"duration" bson:"duration,omitempty"`
	DurationStr string         `json:"duration_str" bson:"duration_str,omitempty"`
	RoutingKey  string         `json:"routing_key" bson:"routing_key"`
}

type StepIn struct {
	Pattern PatternIn `json:"pattern" bson:"pattern"`
	Depth   int       `json:"depth" bson:"depth"`
}

type StepOut struct {
	Pattern  PatternOut    `json:"pattern" bson:"pattern"`
	NextStep int           `json:"next_step" bson:"next_step"`
	Record   HistoryRecord `json:"record" bson:"record"`
	Enrich   []string      `json:"enrich" bson:"enrich"`
}

// Step is the structure to define a step in the workflow
type Step struct {
	Name          string    `json:"name" bson:"name"`
	AutoPropagate bool      `json:"auto_propagate" bson:"auto_propagate"`
	In            []StepIn  `json:"in" bson:"in"`
	Out           []StepOut `json:"out" bson:"out"`
	Fail          []StepOut `json:"fail" bson:"fail"`
}

// GetOutKey is used to ask the routing key to send to RabbitMQ to trigger end of the step in workflow.
func (s *Step) GetOutKey(flowName string) string {
	var out_key bytes.Buffer

	out_key.WriteString("flow.")
	out_key.WriteString(flowName)
	out_key.WriteString(".")
	out_key.WriteString(s.Name)
	out_key.WriteString(".end")

	return out_key.String()
}

// GetInKey is used to ask the routing key to send to RabbitMQ to trigger start of the step in workflow.
func (s *Step) GetInKey(flowName string) string {
	var out_key bytes.Buffer

	out_key.WriteString("flow.")
	out_key.WriteString(flowName)
	out_key.WriteString(".")
	out_key.WriteString(s.Name)
	out_key.WriteString(".start")

	return out_key.String()
}

// Marshal is a shortcut on step to Marshal or panic if err
func (s *Step) Marshal() []byte {
	ret, err := json.Marshal(s)
	FailOnError(err, "Fail to unmarshal Step")
	return ret
}

// Flow is the base level of a flow definition
type Flow struct {
	FlowId           uuid.UUID      `json:"w_id" bson:"w_id" validate:"required"`
	BranchId         uuid.UUID      `json:"b_id" bson:"b_id" validate:"required"`
	StepId           uuid.UUID      `json:"s_id" bson:"s_id" validate:"required"`
	CorrelationChain []string       `json:"c_chain" bson:"c_chain"`
	Name             string         `json:"name" bson:"name" validate:"required"`
	Steps            []Step         `json:"steps" bson:"steps" validate:"required"`
	CurrentStep      int            `json:"current_step" bson:"current_step"`
	JoinResult       JoinResult     `json:"join_result" bson:"join_result"`
	Metadata         Table          `json:"metadata" bson:"metadata"`
	Progress         ProgressDetail `json:"progress" bson:"progress"`
}

// NewFlow initialize a new flow
func NewFlow(name string, steps []Step) *Flow {
	return &Flow{
		Name:             name,
		Steps:            steps,
		CurrentStep:      0,
		FlowId:           uuid.New(),
		BranchId:         uuid.New(),
		StepId:           uuid.New(),
		CorrelationChain: []string{},
		Metadata:         nil,
		Progress: ProgressDetail{
			Name:       name,
			CreateDate: time.Now(),
			Status:     StepPending,
		},
	}
}

// Validate Flow json struct
func (f *Flow) Validate() error {
	v := validator.New()
	if err := v.Struct(f); err != nil {
		return err
	}
	return nil
}

// FromBroker is used to UnMarshal a []byte message received from broker
func FromBroker(body []byte) (*Flow, error) {
	var flow Flow
	if err := json.Unmarshal(body, &flow); err != nil {
		return nil, err
	}
	if err := flow.Validate(); err != nil {
		return nil, err
	}

	return &flow, nil
}

// GetCurrentStep return current step of the flow
func (f *Flow) GetCurrentStep() *Step {
	return &f.Steps[f.CurrentStep]
}

// GetCurrentInKey return the current step in routing key for brokers
func (f *Flow) GetCurrentInKey() string {
	step := f.GetCurrentStep()
	return step.GetInKey(f.Name)
}

// GetCurrentOutKey return the current step out routing key for brokers
func (f *Flow) GetCurrentOutKey() string {
	step := f.GetCurrentStep()
	return step.GetOutKey(f.Name)
}

// AssignFlowId is used to setup a new UUID for current flow
func (f *Flow) AssignFlowId() {
	f.FlowId = uuid.New()
}

// AssignBranchId is used to setup a new UUID for current branch
func (f *Flow) AssignBranchId() {
	f.BranchId = uuid.New()
}

// AssignStepId is used to setup a new UUID for current step
func (f *Flow) AssignStepId() {
	f.StepId = uuid.New()
}

// Marshal is a shortcut on flow to Marshal or panic if err
func (f *Flow) Marshal() []byte {
	ret, err := json.Marshal(f)
	FailOnError(err, "Fail to unmarshal Flow")
	return ret
}

func (f *Flow) DeepCopy() *Flow {
	// FIXME: Setup K8S GenGo package to generate efficient DeepCopy function instead of copier package
	var buffer bytes.Buffer
	var dst Flow

	err_encode := gob.NewEncoder(&buffer).Encode(f)
	FailOnError(err_encode, "Fail to encode flow")

	err_decode := gob.NewDecoder(&buffer).Decode(&dst)
	FailOnError(err_decode, "Fail to decode encoded flow")

	return &dst
}

// SetMetadata is used to update the f.Metadata Table if so or set it using metadata arg
func (f *Flow) SetMetadata(metadata Table) error {
	if err := metadata.Validate(); err != nil {
		return err
	}

	if f.Metadata != nil {
		for key, val := range metadata {
			f.Metadata[key] = val
		}
	} else {
		f.Metadata = metadata
	}
	return nil
}
