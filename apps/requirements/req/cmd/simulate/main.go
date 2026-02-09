package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/glemzurg/go-tlaplus/internal/simulator/engine"
	"github.com/glemzurg/go-tlaplus/internal/simulator/loader"
	"github.com/glemzurg/go-tlaplus/internal/simulator/report"
	"github.com/glemzurg/go-tlaplus/internal/simulator/trace"
)

func main() {
	maxSteps := flag.Int("max-steps", 100, "Maximum simulation steps")
	seed := flag.Int64("seed", 0, "Random seed (0 = current time)")
	stopOnViolation := flag.Bool("stop-on-violation", false, "Stop at first violation")
	output := flag.String("output", "text", "Output format: text or json")
	showTrace := flag.Bool("trace", false, "Include full step trace in output")
	quiet := flag.Bool("quiet", false, "Only output violations")
	flag.Parse()

	if flag.NArg() < 1 {
		fmt.Fprintf(os.Stderr, "Usage: simulate [flags] <model-path>\n\nFlags:\n")
		flag.PrintDefaults()
		os.Exit(2)
	}
	modelPath := flag.Arg(0)

	// Determine seed.
	actualSeed := *seed
	if actualSeed == 0 {
		actualSeed = time.Now().UnixNano()
	}

	// Load model.
	model, err := loader.LoadModel(modelPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading model: %v\n", err)
		os.Exit(1)
	}

	// Configure and run simulation.
	config := engine.SimulationConfig{
		MaxSteps:        *maxSteps,
		RandomSeed:      actualSeed,
		StopOnViolation: *stopOnViolation,
	}

	eng, err := engine.NewSimulationEngine(model, config)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating simulation engine: %v\n", err)
		os.Exit(1)
	}

	result, err := eng.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Simulation error: %v\n", err)
		os.Exit(1)
	}

	// Build reports.
	simTrace := trace.FromResult(result)
	violationReport := report.FromViolations(result.Violations)

	// Output.
	switch *output {
	case "json":
		outputJSON(simTrace, violationReport, *showTrace, *quiet)
	default:
		outputText(simTrace, violationReport, *showTrace, *quiet, actualSeed)
	}

	// Exit code.
	if violationReport.HasViolations() {
		os.Exit(1)
	}
}

func outputText(simTrace *trace.SimulationTrace, violationReport *report.ViolationReport, showTrace, quiet bool, seed int64) {
	if !quiet {
		fmt.Printf("Simulation completed: %d steps, terminated: %s (seed: %d)\n\n",
			simTrace.StepsTaken, simTrace.TerminationReason, seed)
	}

	if showTrace && !quiet {
		fmt.Print(simTrace.FormatText())
		fmt.Println()
	}

	fmt.Print(violationReport.FormatText())
}

func outputJSON(simTrace *trace.SimulationTrace, violationReport *report.ViolationReport, showTrace, quiet bool) {
	output := make(map[string]interface{})

	if !quiet {
		output["summary"] = map[string]interface{}{
			"steps_taken":        simTrace.StepsTaken,
			"termination_reason": simTrace.TerminationReason,
		}
	}

	if showTrace && !quiet {
		output["trace"] = simTrace
	}

	output["violations"] = violationReport

	data, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error marshaling output: %v\n", err)
		os.Exit(1)
	}
	fmt.Println(string(data))
}
