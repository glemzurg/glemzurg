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
	ctx := NewContext("model", "")

	tests := []struct {
		testName string
		err      *ValidationError
		expected string
	}{
		{
			testName: "minimal error",
			err:      New(ctx, "TEST_CODE", "something went wrong", ""),
			expected: `[TEST_CODE] something went wrong at model`,
		},
		{
			testName: "error with field",
			err:      New(ctx, "TEST_CODE", "something went wrong", "Name"),
			expected: `[TEST_CODE] something went wrong (field: Name) at model`,
		},
		{
			testName: "error with field got and want",
			err:      NewWithValues(ctx, "TEST_CODE", "something went wrong", "Type", "invalid", "one of: person, system"),
			expected: `[TEST_CODE] something went wrong (field: Type, got: "invalid", want: one of: person, system) at model`,
		},
		{
			testName: "error with deep path",
			err: New(
				ctx.Child("domains", "domain1").Child("classes", "order"),
				"TEST_CODE", "something went wrong", "",
			),
			expected: `[TEST_CODE] something went wrong at model.domains[domain1].classes[order]`,
		},
		{
			testName: "full error",
			err: NewWithValues(
				ctx.Child("domains", "d1"),
				"CLASS_NAME_REQUIRED", "class name is required", "Name", "", "non-empty string",
			),
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
	ctx := NewContext("model", "")
	err1 := New(ctx, "TEST_CODE", "msg1", "")
	err2 := New(ctx, "TEST_CODE", "msg2", "")
	err3 := New(ctx, "OTHER_CODE", "msg1", "")

	// Same code matches.
	suite.Require().ErrorIs(err1, err2)
	// Different code does not match.
	suite.NotErrorIs(err1, err3)
}

func (suite *ErrorSuite) TestValidationErrorAs() {
	ctx := NewContext("model", "")
	err := New(ctx, "TEST_CODE", "msg", "")

	var ve *ValidationError
	suite.Require().ErrorAs(err, &ve)
	suite.Equal(Code("TEST_CODE"), ve.Code())
}

func (suite *ErrorSuite) TestGetters() {
	ctx := NewContext("model", "")
	err := NewWithValues(ctx, "TEST_CODE", "msg", "Field1", "bad", "good")
	suite.Equal(Code("TEST_CODE"), err.Code())
	suite.Equal("msg", err.Message())
	suite.Equal("Field1", err.Field())
	suite.Equal("bad", err.Got())
	suite.Equal("good", err.Want())
	suite.Len(err.Path(), 1) // Has the context path.
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
	suite.Len(ctx.ContextPath(), 1)
	suite.Equal("model", ctx.ContextPath()[0].Entity)

	child := ctx.Child("domains", "d1")
	suite.Len(child.ContextPath(), 2)
	suite.Equal("domains", child.ContextPath()[1].Entity)
	suite.Equal("d1", child.ContextPath()[1].Key)

	// Original context is unchanged.
	suite.Len(ctx.ContextPath(), 1)
}

func (suite *ErrorSuite) TestNewPanicsOnNilContext() {
	suite.Panics(func() { _ = New(nil, "CODE", "msg", "Field") })
}

func (suite *ErrorSuite) TestNewPanicsOnEmptyCode() {
	ctx := NewContext("model", "")
	suite.Panics(func() { _ = New(ctx, "", "msg", "Field") })
}

func (suite *ErrorSuite) TestNewPanicsOnEmptyMessage() {
	ctx := NewContext("model", "")
	suite.Panics(func() { _ = New(ctx, "CODE", "", "Field") })
}

func (suite *ErrorSuite) TestNewWithValuesPanicsOnNilContext() {
	suite.Panics(func() { _ = NewWithValues(nil, "CODE", "msg", "Field", "got", "want") })
}

func (suite *ErrorSuite) TestNewWithValuesPanicsOnEmptyCode() {
	ctx := NewContext("model", "")
	suite.Panics(func() { _ = NewWithValues(ctx, "", "msg", "Field", "got", "want") })
}

func (suite *ErrorSuite) TestNewWithValuesPanicsOnEmptyMessage() {
	ctx := NewContext("model", "")
	suite.Panics(func() { _ = NewWithValues(ctx, "CODE", "", "Field", "got", "want") })
}

func (suite *ErrorSuite) TestPathIsPopulated() {
	ctx := NewContext("model", "").Child("domains", "d1").Child("classes", "order")
	err := New(ctx, "TEST_CODE", "msg", "Field")

	suite.Len(err.Path(), 3)
	suite.Equal("model", err.Path()[0].Entity)
	suite.Equal("domains", err.Path()[1].Entity)
	suite.Equal("d1", err.Path()[1].Key)
	suite.Equal("classes", err.Path()[2].Entity)
	suite.Equal("order", err.Path()[2].Key)
}
