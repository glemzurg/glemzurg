package main

import (
	"bytes"
	"encoding/json"
	"log"
	"os"
	"strings"
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/report"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/trace"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type OutputSuite struct {
	suite.Suite
}

func TestOutputSuite(t *testing.T) {
	suite.Run(t, new(OutputSuite))
}

func (s *OutputSuite) TestShouldShowStepTrace() {
	cases := []struct {
		name          string
		showTrace     bool
		quiet         bool
		hasViolations bool
		want          bool
	}{
		{name: "clean run shows steps by default", hasViolations: false, want: true},
		{name: "violations hide steps unless trace flag", hasViolations: true, want: false},
		{name: "trace flag forces steps with violations", showTrace: true, hasViolations: true, want: true},
		{name: "quiet suppresses clean-run steps", quiet: true, hasViolations: false, want: false},
		{name: "quiet suppresses trace flag", showTrace: true, quiet: true, hasViolations: true, want: false},
	}

	for _, tc := range cases {
		s.Run(tc.name, func() {
			got := shouldShowStepTrace(tc.showTrace, tc.quiet, tc.hasViolations)
			s.Equal(tc.want, got)
		})
	}
}

func (s *OutputSuite) TestOutputTextCleanRunIncludesSteps() {
	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer log.SetOutput(os.Stderr)

	simTrace := &trace.SimulationTrace{
		StepsTaken:        1,
		TerminationReason: "max_steps",
		Steps: []trace.TraceStep{
			{
				StepNumber: 1,
				Kind:       "transition",
				ClassName:  "Partner",
				ClassKey:   "domain/finance/subdomain/wallet/class/partner",
				InstanceID: 1,
				FromState:  "active",
				ToState:    "active",
				EventName:  "update",
			},
		},
	}
	violationReport := report.FromViolations(nil)

	outputText(simTrace, violationReport, false, false, 42)

	text := buf.String()
	s.Contains(text, "Simulation completed: 1 steps")
	s.Contains(text, "[1] Partner#1: active -> active")
	s.Contains(text, "No violations found.")
}

func (s *OutputSuite) TestOutputTextViolationsHideStepsWithoutTrace() {
	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer log.SetOutput(os.Stderr)

	simTrace := &trace.SimulationTrace{
		StepsTaken:        1,
		TerminationReason: "violation",
		Steps: []trace.TraceStep{
			{
				StepNumber: 1,
				Kind:       "creation",
				ClassName:  "Partner",
				ClassKey:   "domain/finance/subdomain/wallet/class/partner",
				InstanceID: 1,
				ToState:    "active",
			},
		},
	}
	violationReport := &report.ViolationReport{
		TotalCount: 1,
		Summary:    "1 violations found: 1 other",
		Categories: []report.ViolationCategory{
			{
				Name:  "Other Violations",
				Count: 1,
				Violations: []report.ViolationEntry{
					{Type: "state_machine_incomplete", Message: "example violation"},
				},
			},
		},
	}

	outputText(simTrace, violationReport, false, false, 42)

	text := buf.String()
	s.NotContains(text, "[1] CREATE Partner#1")
	s.Contains(text, "1 violations found")
}

func (s *OutputSuite) TestOutputJSONCleanRunIncludesTrace() {
	simTrace := &trace.SimulationTrace{
		StepsTaken:        2,
		TerminationReason: "max_steps",
		Steps: []trace.TraceStep{
			{StepNumber: 1, Kind: "creation", ClassName: "Partner", ClassKey: "k", InstanceID: 1, ToState: "active"},
		},
	}
	violationReport := report.FromViolations(nil)

	var buf bytes.Buffer
	outputJSONTo(&buf, simTrace, violationReport, false, false)

	var payload map[string]any
	s.Require().NoError(json.Unmarshal(buf.Bytes(), &payload))
	_, ok := payload["trace"]
	s.Require().True(ok)
}

func (s *OutputSuite) TestOutputJSONViolationsOmitTraceWithoutFlag() {
	simTrace := &trace.SimulationTrace{
		StepsTaken:        1,
		TerminationReason: "violation",
		Steps: []trace.TraceStep{
			{StepNumber: 1, Kind: "creation", ClassName: "Partner", ClassKey: "k", InstanceID: 1, ToState: "active"},
		},
	}
	violationReport := &report.ViolationReport{TotalCount: 1, Summary: "1 violations found: 1 other"}

	var buf bytes.Buffer
	outputJSONTo(&buf, simTrace, violationReport, false, false)

	var payload map[string]any
	s.Require().NoError(json.Unmarshal(buf.Bytes(), &payload))
	_, ok := payload["trace"]
	s.Require().False(ok)
}

func outputJSONTo(buf *bytes.Buffer, simTrace *trace.SimulationTrace, violationReport *report.ViolationReport, showTrace, quiet bool) {
	output := make(map[string]any)

	if !quiet {
		output["summary"] = map[string]any{
			"steps_taken":        simTrace.StepsTaken,
			"termination_reason": simTrace.TerminationReason,
		}
	}

	if shouldShowStepTrace(showTrace, quiet, violationReport.HasViolations()) {
		output["trace"] = simTrace
	}

	output["violations"] = violationReport

	data, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		panic(err)
	}
	buf.Write(data)
	buf.WriteString("\n")
}

func TestOutputJSONToProducesValidJSON(t *testing.T) {
	var buf bytes.Buffer
	outputJSONTo(&buf, &trace.SimulationTrace{StepsTaken: 0, TerminationReason: "max_steps"}, report.FromViolations(nil), false, false)
	require.True(t, strings.HasSuffix(strings.TrimSpace(buf.String()), "}"))
}
