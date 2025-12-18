package requirements

import (
	"fmt"
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
	for i, test := range tests {
		testName := fmt.Sprintf("Case %d: %+v", i, test)
		obj, err := NewTransition(test.key, test.fromStateKey, test.eventKey, test.guardKey, test.actionKey, test.toStateKey, test.umlComment)
		if test.errstr == "" {
			assert.Nil(suite.T(), err, testName)
			assert.Equal(suite.T(), test.obj, obj, testName)
		} else {
			assert.ErrorContains(suite.T(), err, test.errstr, testName)
			assert.Empty(suite.T(), obj, testName)
		}
	}
}
