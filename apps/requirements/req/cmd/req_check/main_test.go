package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/coreerr"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/parser_ai"
	"github.com/stretchr/testify/suite"
)

type CLISuite struct {
	suite.Suite
}

func TestCLISuite(t *testing.T) {
	suite.Run(t, new(CLISuite))
}

// --- flattenErrors tests ---

func (s *CLISuite) TestFlattenErrors_Nil() {
	result := flattenErrors(nil)
	s.Nil(result)
}

func (s *CLISuite) TestFlattenErrors_SingleError() {
	err := fmt.Errorf("single error")
	result := flattenErrors(err)
	s.Require().Len(result, 1)
	s.Equal("single error", result[0].Error())
}

func (s *CLISuite) TestFlattenErrors_JoinedErrors() {
	err := errors.Join(
		fmt.Errorf("error 1"),
		fmt.Errorf("error 2"),
		fmt.Errorf("error 3"),
	)
	result := flattenErrors(err)
	s.Require().Len(result, 3)
	s.Equal("error 1", result[0].Error())
	s.Equal("error 2", result[1].Error())
	s.Equal("error 3", result[2].Error())
}

func (s *CLISuite) TestFlattenErrors_NestedJoinedErrors() {
	inner := errors.Join(
		fmt.Errorf("inner 1"),
		fmt.Errorf("inner 2"),
	)
	outer := errors.Join(
		fmt.Errorf("outer 1"),
		inner,
	)
	result := flattenErrors(outer)
	s.Require().Len(result, 3)
	s.Equal("outer 1", result[0].Error())
	s.Equal("inner 1", result[1].Error())
	s.Equal("inner 2", result[2].Error())
}

// --- outputJSON tests ---

func (s *CLISuite) TestOutputJSON_ParseError() {
	pe := parser_ai.NewParseError(parser_ai.ErrModelNameRequired, "model name is required", "model.json").
		WithField("name").
		WithHint("add a name field")
	errs := []error{pe}

	// Capture stdout.
	var buf bytes.Buffer
	outputJSONTo(&buf, errs)

	var items []map[string]any
	err := json.Unmarshal(buf.Bytes(), &items)
	s.Require().NoError(err)
	s.Require().Len(items, 1)

	item := items[0]
	s.Equal("parse", item["type"])
	s.Equal("E1001", item["code"])
	s.Equal("model name is required", item["message"])
	s.Equal("model.json", item["file"])
	s.Equal("name", item["field"])
	s.Equal("add a name field | run: req_check --explain E1001", item["hint"])
}

func (s *CLISuite) TestOutputJSON_ValidationError() {
	ctx := coreerr.NewContext("test", "")
	ve := coreerr.NewWithValues(ctx, "TEST_CODE", "test message", "field1", "bad_value", "good_value")
	errs := []error{ve}

	var buf bytes.Buffer
	outputJSONTo(&buf, errs)

	var items []map[string]any
	err := json.Unmarshal(buf.Bytes(), &items)
	s.Require().NoError(err)
	s.Require().Len(items, 1)

	item := items[0]
	s.Equal("validation", item["type"])
	s.Equal("TEST_CODE", item["code"])
	s.Equal("test message", item["message"])
	s.Equal("field1", item["field"])
	s.Equal("bad_value", item["got"])
	s.Equal("good_value", item["want"])
}

func (s *CLISuite) TestOutputJSON_GenericError() {
	errs := []error{fmt.Errorf("unknown error")}

	var buf bytes.Buffer
	outputJSONTo(&buf, errs)

	var items []map[string]any
	err := json.Unmarshal(buf.Bytes(), &items)
	s.Require().NoError(err)
	s.Require().Len(items, 1)

	item := items[0]
	s.Equal("error", item["type"])
	s.Equal("unknown error", item["message"])
}

func (s *CLISuite) TestOutputJSON_MixedErrors() {
	pe := parser_ai.NewParseError(parser_ai.ErrModelNameRequired, "parse error", "model.json")
	ctx := coreerr.NewContext("test", "")
	ve := coreerr.New(ctx, "VAL_CODE", "validation error", "field")
	ge := fmt.Errorf("generic error")
	errs := []error{pe, ve, ge}

	var buf bytes.Buffer
	outputJSONTo(&buf, errs)

	var items []map[string]any
	err := json.Unmarshal(buf.Bytes(), &items)
	s.Require().NoError(err)
	s.Require().Len(items, 3)

	s.Equal("parse", items[0]["type"])
	s.Equal("validation", items[1]["type"])
	s.Equal("error", items[2]["type"])
}

// --- runExplain tests ---

func (s *CLISuite) TestRunExplain_ValidCode() {
	// E1001 has a doc file.
	var buf bytes.Buffer
	err := runExplainTo(&buf, "E1001")
	s.Require().NoError(err)
	s.NotEmpty(buf.String())
	s.Contains(buf.String(), "1001")
}

func (s *CLISuite) TestRunExplain_ValidCodeWithoutPrefix() {
	var buf bytes.Buffer
	err := runExplainTo(&buf, "1001")
	s.Require().NoError(err)
	s.NotEmpty(buf.String())
}

func (s *CLISuite) TestRunExplain_InvalidCode() {
	var buf bytes.Buffer
	err := runExplainTo(&buf, "not_a_number")
	s.Error(err)
}

func (s *CLISuite) TestRunExplain_UnknownCode() {
	var buf bytes.Buffer
	err := runExplainTo(&buf, "99999")
	s.Error(err)
}

// --- runSchema tests ---

func (s *CLISuite) TestRunSchema_ValidEntity() {
	var buf bytes.Buffer
	err := runSchemaTo(&buf, "model")
	s.Require().NoError(err)
	s.NotEmpty(buf.String())
	// Should be valid JSON.
	s.True(json.Valid(buf.Bytes()), "schema output should be valid JSON")
}

func (s *CLISuite) TestRunSchema_CaseInsensitive() {
	var buf bytes.Buffer
	err := runSchemaTo(&buf, "Model")
	s.Require().NoError(err)
	s.NotEmpty(buf.String())
}

func (s *CLISuite) TestRunSchema_UnknownEntity() {
	var buf bytes.Buffer
	err := runSchemaTo(&buf, "nonexistent")
	s.Error(err)
}
