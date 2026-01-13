package model_state

import (
	"testing"

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
	tests := []struct {
		testName     string
		key          string
		fromStateKey string
		eventKey     string
		guardKey     string
		actionKey    string
		toStateKey   string
		umlComment   string
		obj          Transition
		errstr       string
	}{
		// OK.
		{
			testName:     "ok with all fields",
			key:          "Key",
			fromStateKey: "FromStateKey",
			eventKey:     "EventKey",
			guardKey:     "GuardKey",
			actionKey:    "ActionKey",
			toStateKey:   "ToStateKey",
			umlComment:   "UmlComment",
			obj: Transition{
				Key:          "Key",
				FromStateKey: "FromStateKey",
				EventKey:     "EventKey",
				GuardKey:     "GuardKey",
				ActionKey:    "ActionKey",
				ToStateKey:   "ToStateKey",
				UmlComment:   "UmlComment",
			},
		},
		{
			testName:     "ok with minimal fields",
			key:          "Key",
			fromStateKey: "FromStateKey",
			eventKey:     "EventKey",
			guardKey:     "",
			actionKey:    "",
			toStateKey:   "ToStateKey",
			umlComment:   "",
			obj: Transition{
				Key:          "Key",
				FromStateKey: "FromStateKey",
				EventKey:     "EventKey",
				GuardKey:     "",
				ActionKey:    "",
				ToStateKey:   "ToStateKey",
				UmlComment:   "",
			},
		},

		// Error states.
		{
			testName:     "error with blank key",
			key:          "",
			fromStateKey: "FromStateKey",
			eventKey:     "EventKey",
			guardKey:     "GuardKey",
			actionKey:    "ActionKey",
			toStateKey:   "ToStateKey",
			umlComment:   "UmlComment",
			errstr:       `Key: cannot be blank`,
		},
		{
			testName:     "error with blank event key",
			key:          "Key",
			fromStateKey: "FromStateKey",
			eventKey:     "",
			guardKey:     "GuardKey",
			actionKey:    "ActionKey",
			toStateKey:   "ToStateKey",
			umlComment:   "UmlComment",
			errstr:       `EventKey: cannot be blank`,
		},
		{
			testName:     "error with both state keys blank",
			key:          "Key",
			fromStateKey: "",
			eventKey:     "EventKey",
			guardKey:     "GuardKey",
			actionKey:    "ActionKey",
			toStateKey:   "",
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
