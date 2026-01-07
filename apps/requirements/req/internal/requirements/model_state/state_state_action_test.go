package model_state

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

func TestStateActionSuite(t *testing.T) {
	suite.Run(t, new(StateActionSuite))
}

type StateActionSuite struct {
	suite.Suite
}

func (suite *StateActionSuite) TestNew() {
	tests := []struct {
		key       string
		actionKey string
		when      string
		obj       StateAction
		errstr    string
	}{
		// OK.
		{
			key:       "Key",
			actionKey: "ActionKey",
			when:      "entry",
			obj: StateAction{
				Key:       "Key",
				ActionKey: "ActionKey",
				When:      "entry",
			},
		},
		{
			key:       "Key",
			actionKey: "ActionKey",
			when:      "exit",
			obj: StateAction{
				Key:       "Key",
				ActionKey: "ActionKey",
				When:      "exit",
			},
		},
		{
			key:       "Key",
			actionKey: "ActionKey",
			when:      "do",
			obj: StateAction{
				Key:       "Key",
				ActionKey: "ActionKey",
				When:      "do",
			},
		},

		// Error states.
		{
			key:       "",
			actionKey: "ActionKey",
			when:      "entry",
			errstr:    `Key: cannot be blank`,
		},
		{
			key:       "Key",
			actionKey: "",
			when:      "entry",
			errstr:    `ActionKey: cannot be blank`,
		},
		{
			key:       "Key",
			actionKey: "ActionKey",
			when:      "",
			errstr:    `When: cannot be blank`,
		},
		{
			key:       "Key",
			actionKey: "ActionKey",
			when:      "unknown",
			errstr:    `When: must be a valid value`,
		},
	}
	for i, test := range tests {
		testName := fmt.Sprintf("Case %d: %+v", i, test)
		obj, err := NewStateAction(test.key, test.actionKey, test.when)
		if test.errstr == "" {
			assert.Nil(suite.T(), err, testName)
			assert.Equal(suite.T(), test.obj, obj, testName)
		} else {
			assert.ErrorContains(suite.T(), err, test.errstr, testName)
			assert.Empty(suite.T(), obj, testName)
		}
	}
}
