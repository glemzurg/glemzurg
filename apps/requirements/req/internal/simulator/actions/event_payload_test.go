package actions

import (
	"math/rand"
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_state"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/helper"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/evaluator"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/object"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

func TestEventPayloadSuite(t *testing.T) {
	suite.Run(t, new(EventPayloadSuite))
}

type EventPayloadSuite struct {
	suite.Suite
	classKey  identity.Key
	actionKey identity.Key
}

func (s *EventPayloadSuite) SetupSuite() {
	s.classKey = helper.Must(identity.NewClassKey(
		helper.Must(identity.NewSubdomainKey(helper.Must(identity.NewDomainKey("d")), "sd")),
		"order",
	))
	s.actionKey = helper.Must(identity.NewActionKey(s.classKey, "process"))
}

func (s *EventPayloadSuite) TestSampleEventPayloadIncludesEventOnlyNames() {
	eventKey := helper.Must(identity.NewEventKey(s.classKey, "submit"))
	event := model_state.NewEvent(eventKey, "Submit", "", []string{
		"quantity",
		"extra_telemetry",
	})

	quantity := helper.Must(model_state.NewParameter(s.actionKey, "quantity", "[1 .. 10] at 1 unit", false))
	action := model_state.NewAction(s.actionKey, "Process", "", nil, nil, nil, []model_state.Parameter{quantity})

	binder := NewParameterBinder()
	sampler := NewParameterSampler(binder, nil)
	rng := rand.New(rand.NewSource(42)) //nolint:gosec // deterministic test seed

	result, err := SampleEventPayload(event, &action, binder, sampler, rng)
	s.Require().NoError(err)

	s.Contains(result, "quantity")
	s.Contains(result, "extra_telemetry")
	s.NotNil(result["quantity"])
	s.NotNil(result["extra_telemetry"])
	s.True(isEventOnlyValue(result["extra_telemetry"]))
}

func (s *EventPayloadSuite) TestSampleEventPayloadEventOnlyNamesWithoutAction() {
	eventKey := helper.Must(identity.NewEventKey(s.classKey, "notify"))
	event := model_state.NewEvent(eventKey, "Notify", "", []string{"extra_telemetry"})

	binder := NewParameterBinder()
	rng := rand.New(rand.NewSource(7)) //nolint:gosec // deterministic test seed

	result, err := SampleEventPayload(event, nil, binder, nil, rng)
	s.Require().NoError(err)
	s.Require().Len(result, 1)
	s.Contains(result, "extra_telemetry")
	s.True(isEventOnlyValue(result["extra_telemetry"]))
}

func (s *EventPayloadSuite) TestGenerateEventOnlyParameterValueAlternates() {
	binder := NewParameterBinder()
	rng := rand.New(rand.NewSource(99)) //nolint:gosec // deterministic test seed

	var sawString bool
	var sawEmpty bool
	for range 50 {
		val := binder.GenerateEventOnlyParameterValue(rng)
		switch val.(type) {
		case *object.String:
			sawString = true
		default:
			if val == evaluator.EMPTY_SET {
				sawEmpty = true
			}
		}
	}
	s.True(sawString)
	s.True(sawEmpty)
}

func isEventOnlyValue(val object.Object) bool {
	if val == nil {
		return false
	}
	if _, ok := val.(*object.String); ok {
		return true
	}
	return val == evaluator.EMPTY_SET
}

func TestMatchActionParametersByEventNames(t *testing.T) {
	classKey := helper.Must(identity.NewClassKey(
		helper.Must(identity.NewSubdomainKey(helper.Must(identity.NewDomainKey("d")), "sd")),
		"order",
	))
	actionKey := helper.Must(identity.NewActionKey(classKey, "process"))

	quantity := helper.Must(model_state.NewParameter(actionKey, "quantity", "Nat", false))
	priority := helper.Must(model_state.NewParameter(actionKey, "priority", "Nat", false))
	action := model_state.NewAction(actionKey, "Process", "", nil, nil, nil, []model_state.Parameter{quantity, priority})

	matched := matchActionParametersByEventNames([]string{"quantity", "extra_telemetry"}, &action)
	require.Len(t, matched, 1)
	require.Equal(t, "quantity", matched[0].Name)
}
