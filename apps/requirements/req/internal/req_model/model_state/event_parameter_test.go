package model_state

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

func TestEventParameterSuite(t *testing.T) {
	suite.Run(t, new(EventParameterSuite))
}

type EventParameterSuite struct {
	suite.Suite
}

// TestValidate tests all validation rules for EventParameter.
func (suite *EventParameterSuite) TestValidate() {
	tests := []struct {
		testName string
		param    EventParameter
		errstr   string
	}{
		{
			testName: "valid event parameter",
			param: EventParameter{
				Name:   "Name",
				Source: "Source",
			},
		},
		{
			testName: "error blank name",
			param: EventParameter{
				Name:   "",
				Source: "Source",
			},
			errstr: "Name: cannot be blank",
		},
		{
			testName: "error blank source",
			param: EventParameter{
				Name:   "Name",
				Source: "",
			},
			errstr: "Source: cannot be blank",
		},
	}
	for _, tt := range tests {
		suite.T().Run(tt.testName, func(t *testing.T) {
			err := tt.param.Validate()
			if tt.errstr == "" {
				assert.NoError(t, err)
			} else {
				assert.ErrorContains(t, err, tt.errstr)
			}
		})
	}
}

// TestNew tests that NewEventParameter maps parameters correctly and calls Validate.
func (suite *EventParameterSuite) TestNew() {
	// Test parameters are mapped correctly.
	param, err := NewEventParameter("Name", "Source")
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), EventParameter{
		Name:   "Name",
		Source: "Source",
	}, param)

	// Test that Validate is called (invalid data should fail).
	_, err = NewEventParameter("", "Source")
	assert.ErrorContains(suite.T(), err, "Name: cannot be blank")
}

// TestValidateWithParent tests that ValidateWithParent calls Validate.
func (suite *EventParameterSuite) TestValidateWithParent() {
	// Test that Validate is called.
	param := EventParameter{
		Name:   "",
		Source: "Source",
	}
	err := param.ValidateWithParent()
	assert.ErrorContains(suite.T(), err, "Name: cannot be blank", "ValidateWithParent should call Validate()")

	// Test valid case.
	param = EventParameter{
		Name:   "Name",
		Source: "Source",
	}
	err = param.ValidateWithParent()
	assert.NoError(suite.T(), err)
}
