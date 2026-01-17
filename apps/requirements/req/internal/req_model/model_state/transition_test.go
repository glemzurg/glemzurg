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

// TestValidate tests all validation rules for Transition.
func (suite *TransitionSuite) TestValidate() {
	domainKey := helper.Must(identity.NewDomainKey("domain1"))
	subdomainKey := helper.Must(identity.NewSubdomainKey(domainKey, "subdomain1"))
	classKey := helper.Must(identity.NewClassKey(subdomainKey, "class1"))
	fromStateKey := helper.Must(identity.NewStateKey(classKey, "state1"))
	toStateKey := helper.Must(identity.NewStateKey(classKey, "state2"))
	eventKey := helper.Must(identity.NewEventKey(classKey, "event1"))
	guardKey := helper.Must(identity.NewGuardKey(classKey, "guard1"))
	actionKey := helper.Must(identity.NewActionKey(classKey, "action1"))
	validKey := helper.Must(identity.NewTransitionKey(classKey, "state1", "event1", "guard1", "action1", "state2"))

	tests := []struct {
		testName   string
		transition Transition
		errstr     string
	}{
		{
			testName: "valid transition with all fields",
			transition: Transition{
				Key:          validKey,
				FromStateKey: &fromStateKey,
				EventKey:     eventKey,
				GuardKey:     &guardKey,
				ActionKey:    &actionKey,
				ToStateKey:   &toStateKey,
			},
		},
		{
			testName: "valid transition with only from state",
			transition: Transition{
				Key:          validKey,
				FromStateKey: &fromStateKey,
				EventKey:     eventKey,
				ToStateKey:   nil,
			},
		},
		{
			testName: "valid transition with only to state",
			transition: Transition{
				Key:          validKey,
				FromStateKey: nil,
				EventKey:     eventKey,
				ToStateKey:   &toStateKey,
			},
		},
		{
			testName: "error empty key",
			transition: Transition{
				Key:          identity.Key{},
				FromStateKey: &fromStateKey,
				EventKey:     eventKey,
				ToStateKey:   &toStateKey,
			},
			errstr: "keyType: cannot be blank",
		},
		{
			testName: "error wrong key type",
			transition: Transition{
				Key:          domainKey,
				FromStateKey: &fromStateKey,
				EventKey:     eventKey,
				ToStateKey:   &toStateKey,
			},
			errstr: "Key: invalid key type 'domain' for transition",
		},
		{
			testName: "error empty event key",
			transition: Transition{
				Key:          validKey,
				FromStateKey: &fromStateKey,
				EventKey:     identity.Key{},
				ToStateKey:   &toStateKey,
			},
			errstr: "keyType: cannot be blank",
		},
		{
			testName: "error wrong event key type",
			transition: Transition{
				Key:          validKey,
				FromStateKey: &fromStateKey,
				EventKey:     domainKey,
				ToStateKey:   &toStateKey,
			},
			errstr: "EventKey: invalid key type 'domain' for event",
		},
		{
			testName: "error both state keys nil",
			transition: Transition{
				Key:          validKey,
				FromStateKey: nil,
				EventKey:     eventKey,
				ToStateKey:   nil,
			},
			errstr: "FromStateKey, ToStateKey: cannot both be blank",
		},
		{
			testName: "error wrong from state key type",
			transition: Transition{
				Key:          validKey,
				FromStateKey: &domainKey,
				EventKey:     eventKey,
				ToStateKey:   &toStateKey,
			},
			errstr: "FromStateKey: invalid key type 'domain' for from state",
		},
		{
			testName: "error wrong to state key type",
			transition: Transition{
				Key:          validKey,
				FromStateKey: &fromStateKey,
				EventKey:     eventKey,
				ToStateKey:   &domainKey,
			},
			errstr: "ToStateKey: invalid key type 'domain' for to state",
		},
		{
			testName: "error wrong guard key type",
			transition: Transition{
				Key:          validKey,
				FromStateKey: &fromStateKey,
				EventKey:     eventKey,
				GuardKey:     &domainKey,
				ToStateKey:   &toStateKey,
			},
			errstr: "GuardKey: invalid key type 'domain' for guard",
		},
		{
			testName: "error wrong action key type",
			transition: Transition{
				Key:          validKey,
				FromStateKey: &fromStateKey,
				EventKey:     eventKey,
				ActionKey:    &domainKey,
				ToStateKey:   &toStateKey,
			},
			errstr: "ActionKey: invalid key type 'domain' for action",
		},
	}
	for _, tt := range tests {
		suite.T().Run(tt.testName, func(t *testing.T) {
			err := tt.transition.Validate()
			if tt.errstr == "" {
				assert.NoError(t, err)
			} else {
				assert.ErrorContains(t, err, tt.errstr)
			}
		})
	}
}

// TestNew tests that NewTransition maps parameters correctly and calls Validate.
func (suite *TransitionSuite) TestNew() {
	domainKey := helper.Must(identity.NewDomainKey("domain1"))
	subdomainKey := helper.Must(identity.NewSubdomainKey(domainKey, "subdomain1"))
	classKey := helper.Must(identity.NewClassKey(subdomainKey, "class1"))
	fromStateKey := helper.Must(identity.NewStateKey(classKey, "state1"))
	toStateKey := helper.Must(identity.NewStateKey(classKey, "state2"))
	eventKey := helper.Must(identity.NewEventKey(classKey, "event1"))
	guardKey := helper.Must(identity.NewGuardKey(classKey, "guard1"))
	actionKey := helper.Must(identity.NewActionKey(classKey, "action1"))
	key := helper.Must(identity.NewTransitionKey(classKey, "state1", "event1", "guard1", "action1", "state2"))

	// Test parameters are mapped correctly.
	transition, err := NewTransition(key, &fromStateKey, eventKey, &guardKey, &actionKey, &toStateKey, "UmlComment")
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), Transition{
		Key:          key,
		FromStateKey: &fromStateKey,
		EventKey:     eventKey,
		GuardKey:     &guardKey,
		ActionKey:    &actionKey,
		ToStateKey:   &toStateKey,
		UmlComment:   "UmlComment",
	}, transition)

	// Test that Validate is called (invalid data should fail).
	_, err = NewTransition(key, nil, eventKey, nil, nil, nil, "UmlComment")
	assert.ErrorContains(suite.T(), err, "FromStateKey, ToStateKey: cannot both be blank")
}

// TestValidateWithParent tests that ValidateWithParent calls Validate and ValidateParent.
func (suite *TransitionSuite) TestValidateWithParent() {
	domainKey := helper.Must(identity.NewDomainKey("domain1"))
	subdomainKey := helper.Must(identity.NewSubdomainKey(domainKey, "subdomain1"))
	classKey := helper.Must(identity.NewClassKey(subdomainKey, "class1"))
	fromStateKey := helper.Must(identity.NewStateKey(classKey, "state1"))
	toStateKey := helper.Must(identity.NewStateKey(classKey, "state2"))
	eventKey := helper.Must(identity.NewEventKey(classKey, "event1"))
	validKey := helper.Must(identity.NewTransitionKey(classKey, "state1", "event1", "", "", "state2"))
	otherClassKey := helper.Must(identity.NewClassKey(subdomainKey, "other_class"))

	// Test that Validate is called.
	transition := Transition{
		Key:          validKey,
		FromStateKey: nil, // Invalid - both nil
		EventKey:     eventKey,
		ToStateKey:   nil, // Invalid - both nil
	}
	err := transition.ValidateWithParent(&classKey)
	assert.ErrorContains(suite.T(), err, "FromStateKey, ToStateKey: cannot both be blank", "ValidateWithParent should call Validate()")

	// Test that ValidateParent is called - transition key has class1 as parent, but we pass other_class.
	transition = Transition{
		Key:          validKey,
		FromStateKey: &fromStateKey,
		EventKey:     eventKey,
		ToStateKey:   &toStateKey,
	}
	err = transition.ValidateWithParent(&otherClassKey)
	assert.ErrorContains(suite.T(), err, "does not match expected parent", "ValidateWithParent should call ValidateParent()")

	// Test valid case.
	err = transition.ValidateWithParent(&classKey)
	assert.NoError(suite.T(), err)
}
