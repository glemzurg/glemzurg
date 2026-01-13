package model_state

import (
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
		testName  string
		key       string
		actionKey string
		when      string
		obj       StateAction
		errstr    string
	}{
		// OK.
		{
			testName:  "ok with entry",
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
			testName:  "ok with exit",
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
			testName:  "ok with do",
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
			testName:  "error with blank key",
			key:       "",
			actionKey: "ActionKey",
			when:      "entry",
			errstr:    `Key: cannot be blank`,
		},
		{
			testName:  "error with blank action key",
			key:       "Key",
			actionKey: "",
			when:      "entry",
			errstr:    `ActionKey: cannot be blank`,
		},
		{
			testName:  "error with blank when",
			key:       "Key",
			actionKey: "ActionKey",
			when:      "",
			errstr:    `When: cannot be blank`,
		},
		{
			testName:  "error with unknown when",
			key:       "Key",
			actionKey: "ActionKey",
			when:      "unknown",
			errstr:    `When: must be a valid value`,
		},
	}
	for _, tt := range tests {
		_ = suite.T().Run(tt.testName, func(t *testing.T) {
			obj, err := NewStateAction(tt.key, tt.actionKey, tt.when)
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
