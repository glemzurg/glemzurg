package coreerr

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type ErrorSuite struct {
	suite.Suite
}

func TestErrorSuite(t *testing.T) {
	suite.Run(t, new(ErrorSuite))
}

func (suite *ErrorSuite) TestValidationErrorMessage() {
	tests := []struct {
		testName string
		err      *ValidationError
		expected string
	}{
		{
			testName: "minimal error",
			err: &ValidationError{
				Code:    "TEST_CODE",
				Message: "something went wrong",
			},
			expected: `[TEST_CODE] something went wrong`,
		},
		{
			testName: "error with field",
			err: &ValidationError{
				Code:    "TEST_CODE",
				Message: "something went wrong",
				Field:   "Name",
			},
			expected: `[TEST_CODE] something went wrong (field: Name)`,
		},
		{
			testName: "error with field got and want",
			err: &ValidationError{
				Code:    "TEST_CODE",
				Message: "something went wrong",
				Field:   "Type",
				Got:     "invalid",
				Want:    "one of: person, system",
			},
			expected: `[TEST_CODE] something went wrong (field: Type, got: "invalid", want: one of: person, system)`,
		},
		{
			testName: "error with path",
			err: &ValidationError{
				Code:    "TEST_CODE",
				Message: "something went wrong",
				Path: []PathSegment{
					{Entity: "model", Key: ""},
					{Entity: "domains", Key: "domain1"},
					{Entity: "classes", Key: "order"},
				},
			},
			expected: `[TEST_CODE] something went wrong at model.domains[domain1].classes[order]`,
		},
		{
			testName: "full error",
			err: &ValidationError{
				Code:    "CLASS_NAME_REQUIRED",
				Message: "class name is required",
				Path: []PathSegment{
					{Entity: "model", Key: ""},
					{Entity: "domains", Key: "d1"},
				},
				Field: "Name",
				Got:   "",
				Want:  "non-empty string",
			},
			expected: `[CLASS_NAME_REQUIRED] class name is required (field: Name, want: non-empty string) at model.domains[d1]`,
		},
	}

	for _, test := range tests {
		suite.Run(test.testName, func() {
			suite.Equal(test.expected, test.err.Error())
		})
	}
}

func (suite *ErrorSuite) TestValidationErrorIs() {
	err1 := &ValidationError{Code: "TEST_CODE", Message: "msg1"}
	err2 := &ValidationError{Code: "TEST_CODE", Message: "msg2"}
	err3 := &ValidationError{Code: "OTHER_CODE", Message: "msg1"}

	// Same code matches.
	suite.Require().ErrorIs(err1, err2)
	// Different code does not match.
	suite.NotErrorIs(err1, err3)
}

func (suite *ErrorSuite) TestValidationErrorAs() {
	err := &ValidationError{Code: "TEST_CODE", Message: "msg"}

	var ve *ValidationError
	suite.Require().ErrorAs(err, &ve)
	suite.Equal(Code("TEST_CODE"), ve.Code)
}

func (suite *ErrorSuite) TestFormatPath() {
	tests := []struct {
		testName string
		path     []PathSegment
		expected string
	}{
		{
			testName: "empty path",
			path:     nil,
			expected: "",
		},
		{
			testName: "single segment without key",
			path:     []PathSegment{{Entity: "model"}},
			expected: "model",
		},
		{
			testName: "single segment with key",
			path:     []PathSegment{{Entity: "domains", Key: "domain1"}},
			expected: "domains[domain1]",
		},
		{
			testName: "multi segment",
			path: []PathSegment{
				{Entity: "model"},
				{Entity: "domains", Key: "d1"},
				{Entity: "subdomains", Key: "default"},
				{Entity: "classes", Key: "order"},
			},
			expected: "model.domains[d1].subdomains[default].classes[order]",
		},
	}

	for _, test := range tests {
		suite.Run(test.testName, func() {
			suite.Equal(test.expected, FormatPath(test.path))
		})
	}
}

func (suite *ErrorSuite) TestValidationContext() {
	ctx := NewContext("model", "")
	suite.Len(ctx.Path(), 1)
	suite.Equal("model", ctx.Path()[0].Entity)

	child := ctx.Child("domains", "d1")
	suite.Len(child.Path(), 2)
	suite.Equal("domains", child.Path()[1].Entity)
	suite.Equal("d1", child.Path()[1].Key)

	// Original context is unchanged.
	suite.Len(ctx.Path(), 1)
}

func (suite *ErrorSuite) TestValidationContextErr() {
	ctx := NewContext("model", "").Child("classes", "order")
	err := ctx.Err("CLASS_NAME_REQUIRED", "Name", "", "non-empty string", "class name is required")

	suite.Equal(Code("CLASS_NAME_REQUIRED"), err.Code)
	suite.Equal("class name is required", err.Message)
	suite.Equal("Name", err.Field)
	suite.Empty(err.Got)
	suite.Equal("non-empty string", err.Want)
	suite.Len(err.Path, 2)
}

func (suite *ErrorSuite) TestEnsureContext() {
	// Non-nil context is returned as-is.
	existing := NewContext("model", "")
	result := EnsureContext(existing, "other", "key")
	suite.Same(existing, result)

	// Nil context creates a new one.
	result = EnsureContext(nil, "action", "place_order")
	suite.NotNil(result)
	suite.Equal("action", result.Path()[0].Entity)
	suite.Equal("place_order", result.Path()[0].Key)
}
