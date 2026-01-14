package model_state

import (
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/helper"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

func TestTransitionSuite(t *testing.T) {
	suite.Run(t, new(TransitionSuite))
}

type TransitionSuite struct {
	suite.Suite
}

func (suite *TransitionSuite) TestNew() {

	domainKey := helper.Must(identity.NewDomainKey("domain1"))
	subdomainKey := helper.Must(identity.NewSubdomainKey(domainKey, "subdomain1"))
	classKey := helper.Must(identity.NewClassKey(subdomainKey, "class1"))
	fromStateKey := helper.Must(identity.NewStateKey(classKey, "state1"))
	toStateKey := helper.Must(identity.NewStateKey(classKey, "state2"))
	eventKey := helper.Must(identity.NewEventKey(classKey, "event1"))
	guardKey := helper.Must(identity.NewGuardKey(classKey, "guard1"))
	actionKey := helper.Must(identity.NewActionKey(classKey, "action1"))

	tests := []struct {
		testName     string
		key          identity.Key
		fromStateKey *identity.Key
		eventKey     identity.Key
		guardKey     *identity.Key
		actionKey    *identity.Key
		toStateKey   *identity.Key
		umlComment   string
		obj          Transition
		errstr       string
	}{
		// OK.
		{
			testName:     "ok with all fields",
			key:          helper.Must(identity.NewTransitionKey(classKey, "transition1")),
			fromStateKey: &fromStateKey,
			eventKey:     eventKey,
			guardKey:     &guardKey,
			actionKey:    &actionKey,
			toStateKey:   &toStateKey,
			umlComment:   "UmlComment",
			obj: Transition{
				Key:          helper.Must(identity.NewTransitionKey(classKey, "transition1")),
				FromStateKey: &fromStateKey,
				EventKey:     eventKey,
				GuardKey:     &guardKey,
				ActionKey:    &actionKey,
				ToStateKey:   &toStateKey,
				UmlComment:   "UmlComment",
			},
		},
		{
			testName:     "ok with minimal fields",
			key:          helper.Must(identity.NewTransitionKey(classKey, "transition2")),
			fromStateKey: &fromStateKey,
			eventKey:     eventKey,
			guardKey:     nil,
			actionKey:    nil,
			toStateKey:   &toStateKey,
			umlComment:   "",
			obj: Transition{
				Key:          helper.Must(identity.NewTransitionKey(classKey, "transition2")),
				FromStateKey: &fromStateKey,
				EventKey:     eventKey,
				GuardKey:     nil,
				ActionKey:    nil,
				ToStateKey:   &toStateKey,
				UmlComment:   "",
			},
		},
		{
			testName:     "ok with only from state",
			key:          helper.Must(identity.NewTransitionKey(classKey, "transition3")),
			fromStateKey: &fromStateKey,
			eventKey:     eventKey,
			guardKey:     nil,
			actionKey:    nil,
			toStateKey:   nil,
			umlComment:   "",
			obj: Transition{
				Key:          helper.Must(identity.NewTransitionKey(classKey, "transition3")),
				FromStateKey: &fromStateKey,
				EventKey:     eventKey,
				GuardKey:     nil,
				ActionKey:    nil,
				ToStateKey:   nil,
				UmlComment:   "",
			},
		},
		{
			testName:     "ok with only to state",
			key:          helper.Must(identity.NewTransitionKey(classKey, "transition4")),
			fromStateKey: nil,
			eventKey:     eventKey,
			guardKey:     nil,
			actionKey:    nil,
			toStateKey:   &toStateKey,
			umlComment:   "",
			obj: Transition{
				Key:          helper.Must(identity.NewTransitionKey(classKey, "transition4")),
				FromStateKey: nil,
				EventKey:     eventKey,
				GuardKey:     nil,
				ActionKey:    nil,
				ToStateKey:   &toStateKey,
				UmlComment:   "",
			},
		},

		// Error states.
		{
			testName:     "error empty key",
			key:          identity.Key{},
			fromStateKey: &fromStateKey,
			eventKey:     eventKey,
			guardKey:     &guardKey,
			actionKey:    &actionKey,
			toStateKey:   &toStateKey,
			umlComment:   "UmlComment",
			errstr:       "keyType: cannot be blank",
		},
		{
			testName:     "error wrong key type",
			key:          helper.Must(identity.NewDomainKey("domain1")),
			fromStateKey: &fromStateKey,
			eventKey:     eventKey,
			guardKey:     &guardKey,
			actionKey:    &actionKey,
			toStateKey:   &toStateKey,
			umlComment:   "UmlComment",
			errstr:       "Key: invalid key type 'domain' for transition",
		},
		{
			testName:     "error empty event key",
			key:          helper.Must(identity.NewTransitionKey(classKey, "transition5")),
			fromStateKey: &fromStateKey,
			eventKey:     identity.Key{},
			guardKey:     &guardKey,
			actionKey:    &actionKey,
			toStateKey:   &toStateKey,
			umlComment:   "UmlComment",
			errstr:       "keyType: cannot be blank",
		},
		{
			testName:     "error wrong event key type",
			key:          helper.Must(identity.NewTransitionKey(classKey, "transition6")),
			fromStateKey: &fromStateKey,
			eventKey:     helper.Must(identity.NewDomainKey("domain1")),
			guardKey:     &guardKey,
			actionKey:    &actionKey,
			toStateKey:   &toStateKey,
			umlComment:   "UmlComment",
			errstr:       "EventKey: invalid key type 'domain' for event",
		},
		{
			testName:     "error with both state keys nil",
			key:          helper.Must(identity.NewTransitionKey(classKey, "transition7")),
			fromStateKey: nil,
			eventKey:     eventKey,
			guardKey:     &guardKey,
			actionKey:    &actionKey,
			toStateKey:   nil,
			umlComment:   "UmlComment",
			errstr:       `FromStateKey, ToStateKey: cannot both be blank`,
		},
	}
	for _, tt := range tests {
		_ = suite.T().Run(tt.testName, func(t *testing.T) {
			obj, err := NewTransition(tt.key, tt.fromStateKey, tt.eventKey, tt.guardKey, tt.actionKey, tt.toStateKey, tt.umlComment)
			if tt.errstr == "" {
				assert.NoError(t, err)
				assert.Equal(t, tt.obj, obj)
			} else {
				assert.ErrorContains(t, err, tt.errstr)
				assert.Empty(t, obj)
			}
		})
	}
}
