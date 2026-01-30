package parser_ai

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
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
	assert.Nil(suite.T(), err)

	for _, testData := range testDataFiles {
		testName := testData.Filename
		pass := suite.T().Run(testName, func(t *testing.T) {
			var expected inputStateMachine

			actual, err := parseStateMachine([]byte(testData.InputJSON), testData.Filename)
			assert.Nil(t, err, testName)

			err = json.Unmarshal([]byte(testData.ExpectedJSON), &expected)
			assert.Nil(t, err, testName)

			// Compare states
			assert.Equal(t, len(expected.States), len(actual.States), testName+" states count")
			for key, expectedState := range expected.States {
				actualState, exists := actual.States[key]
				assert.True(t, exists, testName+" state '"+key+"' should exist")
				if exists {
					assert.Equal(t, expectedState.Name, actualState.Name, testName+" state '"+key+"' name")
					assert.Equal(t, expectedState.Details, actualState.Details, testName+" state '"+key+"' details")
					assert.Equal(t, expectedState.UMLComment, actualState.UMLComment, testName+" state '"+key+"' uml_comment")
					assert.Equal(t, len(expectedState.Actions), len(actualState.Actions), testName+" state '"+key+"' actions count")
					for i, expectedAction := range expectedState.Actions {
						assert.Equal(t, expectedAction.ActionKey, actualState.Actions[i].ActionKey, testName+" state '"+key+"' action["+string(rune('0'+i))+"] action_key")
						assert.Equal(t, expectedAction.When, actualState.Actions[i].When, testName+" state '"+key+"' action["+string(rune('0'+i))+"] when")
					}
				}
			}

			// Compare events
			assert.Equal(t, len(expected.Events), len(actual.Events), testName+" events count")
			for key, expectedEvent := range expected.Events {
				actualEvent, exists := actual.Events[key]
				assert.True(t, exists, testName+" event '"+key+"' should exist")
				if exists {
					assert.Equal(t, expectedEvent.Name, actualEvent.Name, testName+" event '"+key+"' name")
					assert.Equal(t, expectedEvent.Details, actualEvent.Details, testName+" event '"+key+"' details")
					assert.Equal(t, len(expectedEvent.Parameters), len(actualEvent.Parameters), testName+" event '"+key+"' parameters count")
					for i, expectedParam := range expectedEvent.Parameters {
						assert.Equal(t, expectedParam.Name, actualEvent.Parameters[i].Name, testName+" event '"+key+"' param["+string(rune('0'+i))+"] name")
						assert.Equal(t, expectedParam.Source, actualEvent.Parameters[i].Source, testName+" event '"+key+"' param["+string(rune('0'+i))+"] source")
					}
				}
			}

			// Compare guards
			assert.Equal(t, len(expected.Guards), len(actual.Guards), testName+" guards count")
			for key, expectedGuard := range expected.Guards {
				actualGuard, exists := actual.Guards[key]
				assert.True(t, exists, testName+" guard '"+key+"' should exist")
				if exists {
					assert.Equal(t, expectedGuard.Name, actualGuard.Name, testName+" guard '"+key+"' name")
					assert.Equal(t, expectedGuard.Details, actualGuard.Details, testName+" guard '"+key+"' details")
				}
			}

			// Compare transitions
			assert.Equal(t, len(expected.Transitions), len(actual.Transitions), testName+" transitions count")
			for i, expectedTrans := range expected.Transitions {
				actualTrans := actual.Transitions[i]
				assert.Equal(t, expectedTrans.FromStateKey, actualTrans.FromStateKey, testName+" transition["+string(rune('0'+i))+"] from_state_key")
				assert.Equal(t, expectedTrans.ToStateKey, actualTrans.ToStateKey, testName+" transition["+string(rune('0'+i))+"] to_state_key")
				assert.Equal(t, expectedTrans.EventKey, actualTrans.EventKey, testName+" transition["+string(rune('0'+i))+"] event_key")
				assert.Equal(t, expectedTrans.GuardKey, actualTrans.GuardKey, testName+" transition["+string(rune('0'+i))+"] guard_key")
				assert.Equal(t, expectedTrans.ActionKey, actualTrans.ActionKey, testName+" transition["+string(rune('0'+i))+"] action_key")
				assert.Equal(t, expectedTrans.UMLComment, actualTrans.UMLComment, testName+" transition["+string(rune('0'+i))+"] uml_comment")
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
		suite.T().Run(testName, func(t *testing.T) {
			_, err := parseStateMachine([]byte(testData.InputJSON), testData.Filename)
			assert.NotNil(t, err, testName+" should return an error")

			parseErr, ok := err.(*ParseError)
			assert.True(t, ok, testName+" should return a ParseError")
			if !ok {
				return
			}

			expected := testData.ExpectedError
			assert.Equal(t, expected.Code, parseErr.Code, testName+" error code")
			assert.Equal(t, expected.ErrorFile, parseErr.ErrorFile, testName+" error file")

			if expected.Message != "" {
				assert.Equal(t, expected.Message, parseErr.Message, testName+" error message")
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
			assert.Equal(t, testData.Filename, parseErr.File, testName+" error file path")

			if expected.Field != "" {
				assert.Equal(t, expected.Field, parseErr.Field, testName+" error field")
			}
		})
	}
}
