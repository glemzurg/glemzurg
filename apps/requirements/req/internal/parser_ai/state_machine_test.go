package parser_ai

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

const (
	t_STATE_MACHINE_PATH_OK  = "test_files/state_machine"
	t_STATE_MACHINE_PATH_ERR = t_STATE_MACHINE_PATH_OK + "/err"
)

func TestStateMachineSuite(t *testing.T) {
	suite.Run(t, new(StateMachineSuite))
}

type StateMachineSuite struct {
	suite.Suite
}

func (suite *StateMachineSuite) TestParseStateMachineFiles() {
	testDataFiles, err := t_ContentsForAllJSONFiles(t_STATE_MACHINE_PATH_OK)
	suite.Require().NoError(err)

	for _, testData := range testDataFiles {
		testName := testData.Filename
		pass := suite.Run(testName, func() {
			t := suite.T()
			var expected inputStateMachine

			actual, err := parseStateMachine([]byte(testData.InputJSON), testData.Filename)
			require.NoError(t, err, testName)

			err = json.Unmarshal([]byte(testData.ExpectedJSON), &expected)
			require.NoError(t, err, testName)

			// Compare states
			assert.Len(t, actual.States, len(expected.States), testName+" states count")
			for key, expectedState := range expected.States {
				actualState, exists := actual.States[key]
				assert.True(t, exists, testName+" state '"+key+"' should exist")
				if exists {
					suite.Equal(expectedState.Name, actualState.Name, testName+" state '"+key+"' name")
					suite.Equal(expectedState.Details, actualState.Details, testName+" state '"+key+"' details")
					suite.Equal(expectedState.UMLComment, actualState.UMLComment, testName+" state '"+key+"' uml_comment")
					assert.Len(t, actualState.Actions, len(expectedState.Actions), testName+" state '"+key+"' actions count")
					for i, expectedAction := range expectedState.Actions {
						suite.Equal(expectedAction.ActionKey, actualState.Actions[i].ActionKey, testName+" state '"+key+"' action["+string(rune('0'+i))+"] action_key")
						suite.Equal(expectedAction.When, actualState.Actions[i].When, testName+" state '"+key+"' action["+string(rune('0'+i))+"] when")
					}
				}
			}

			// Compare events
			assert.Len(t, actual.Events, len(expected.Events), testName+" events count")
			for key, expectedEvent := range expected.Events {
				actualEvent, exists := actual.Events[key]
				assert.True(t, exists, testName+" event '"+key+"' should exist")
				if exists {
					suite.Equal(expectedEvent.Name, actualEvent.Name, testName+" event '"+key+"' name")
					suite.Equal(expectedEvent.Details, actualEvent.Details, testName+" event '"+key+"' details")
					assert.Len(t, actualEvent.Parameters, len(expectedEvent.Parameters), testName+" event '"+key+"' parameters count")
					for i, expectedParam := range expectedEvent.Parameters {
						suite.Equal(expectedParam.Name, actualEvent.Parameters[i].Name, testName+" event '"+key+"' param["+string(rune('0'+i))+"] name")
						suite.Equal(expectedParam.DataTypeRules, actualEvent.Parameters[i].DataTypeRules, testName+" event '"+key+"' param["+string(rune('0'+i))+"] data_type_rules")
					}
				}
			}

			// Compare guards
			assert.Len(t, actual.Guards, len(expected.Guards), testName+" guards count")
			for key, expectedGuard := range expected.Guards {
				actualGuard, exists := actual.Guards[key]
				assert.True(t, exists, testName+" guard '"+key+"' should exist")
				if exists {
					suite.Equal(expectedGuard.Name, actualGuard.Name, testName+" guard '"+key+"' name")
					suite.Equal(expectedGuard.Logic.Description, actualGuard.Logic.Description, testName+" guard '"+key+"' logic description")
				}
			}

			// Compare transitions
			assert.Len(t, actual.Transitions, len(expected.Transitions), testName+" transitions count")
			for i, expectedTrans := range expected.Transitions {
				actualTrans := actual.Transitions[i]
				suite.Equal(expectedTrans.FromStateKey, actualTrans.FromStateKey, testName+" transition["+string(rune('0'+i))+"] from_state_key")
				suite.Equal(expectedTrans.ToStateKey, actualTrans.ToStateKey, testName+" transition["+string(rune('0'+i))+"] to_state_key")
				suite.Equal(expectedTrans.EventKey, actualTrans.EventKey, testName+" transition["+string(rune('0'+i))+"] event_key")
				suite.Equal(expectedTrans.GuardKey, actualTrans.GuardKey, testName+" transition["+string(rune('0'+i))+"] guard_key")
				suite.Equal(expectedTrans.ActionKey, actualTrans.ActionKey, testName+" transition["+string(rune('0'+i))+"] action_key")
				suite.Equal(expectedTrans.UMLComment, actualTrans.UMLComment, testName+" transition["+string(rune('0'+i))+"] uml_comment")
			}
		})
		if !pass {
			break
		}
	}
}

func (suite *StateMachineSuite) TestParseStateMachineErrors() {
	testDataFiles, err := t_ContentsForAllErrorJSONFiles(t_STATE_MACHINE_PATH_ERR)
	if err != nil {
		suite.T().Fatalf("Failed to read error test files: %v", err)
	}

	if len(testDataFiles) == 0 {
		return
	}

	for _, testData := range testDataFiles {
		testName := testData.Filename
		suite.Run(testName, func() {
			t := suite.T()
			_, err := parseStateMachine([]byte(testData.InputJSON), testData.Filename)
			require.Error(t, err, testName+" should return an error")

			var parseErr *ParseError
			ok := errors.As(err, &parseErr)
			assert.True(t, ok, testName+" should return a ParseError")
			if !ok {
				return
			}

			expected := testData.ExpectedError
			suite.Equal(expected.Code, parseErr.Code, testName+" error code")
			suite.Equal(expected.ErrorFile, parseErr.ErrorFile, testName+" error file")

			if expected.Message != "" {
				suite.Equal(expected.Message, parseErr.Message, testName+" error message")
			} else if expected.MessagePrefix != "" {
				assert.True(t, len(parseErr.Message) >= len(expected.MessagePrefix) &&
					parseErr.Message[:len(expected.MessagePrefix)] == expected.MessagePrefix,
					testName+" error message should start with '"+expected.MessagePrefix+"', got '"+parseErr.Message+"'")
			}

			if expected.HasSchema {
				assert.NotEmpty(t, parseErr.Schema, testName+" should have schema content")
			} else {
				assert.Empty(t, parseErr.Schema, testName+" should not have schema content")
			}

			assert.NotEmpty(t, parseErr.Docs, testName+" should have docs content")
			suite.Equal(testData.Filename, parseErr.File, testName+" error file path")

			if expected.Field != "" {
				suite.Equal(expected.Field, parseErr.Field, testName+" error field")
			}
		})
	}
}
