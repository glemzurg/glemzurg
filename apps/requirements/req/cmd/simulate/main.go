package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/notation/tla_plus/convert"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/parser_human"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/engine"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/report"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/surface"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/trace"
)

type cliOptions struct {
	maxSteps              int
	seed                  int64
	stopOnViolation       bool
	output                string
	showTrace             bool
	quiet                 bool
	rootSource            string
	modelName             string
	includeSubdomainPaths []string
	includeClassNames     []string
}

func main() {
	opts := parseCLIOptions()
	hasViolations, err := runSimulation(opts)
	if err != nil {
		log.Print(err)
		os.Exit(1)
	}
	if hasViolations {
		os.Exit(1)
	}
}

func parseCLIOptions() cliOptions {
	maxSteps := flag.Int("max-steps", 100, "Maximum simulation steps")
	seed := flag.Int64("seed", 0, "Random seed (0 = current time)")
	stopOnViolation := flag.Bool("stop-on-violation", true, "Stop at first violation")
	continueOnViolation := flag.Bool("continue-on-violation", false, "Keep simulating after violations (overrides -stop-on-violation)")
	output := flag.String("output", "text", "Output format: text or json")
	showTrace := flag.Bool("trace", false, "Include full step trace in output (also shown by default when no violations are found)")
	quiet := flag.Bool("quiet", false, "Only output violations")
	rootSource := flag.String("rootsource", "", "Human model root source directory (e.g. data_sandbox/model)")
	modelName := flag.String("model", "", "Model name when using -rootsource (e.g. evenplay)")
	includeSubdomains := flag.String("include-subdomain", "", "Comma-separated subdomains to simulate: subdomain or domain/subdomain")
	includeClasses := flag.String("include-class", "", "Comma-separated classes to simulate: name, subdomain/class, or domain/subdomain/class")
	flag.Parse()

	stop := *stopOnViolation
	if *continueOnViolation {
		stop = false
	}

	return cliOptions{
		maxSteps:              *maxSteps,
		seed:                  *seed,
		stopOnViolation:       stop,
		output:                *output,
		showTrace:             *showTrace,
		quiet:                 *quiet,
		rootSource:            *rootSource,
		modelName:             *modelName,
		includeSubdomainPaths: parseCommaSeparatedFlag(*includeSubdomains),
		includeClassNames:     parseCommaSeparatedFlag(*includeClasses),
	}
}

func runSimulation(opts cliOptions) (hasViolations bool, err error) {
	model, err := loadModel(opts.rootSource, opts.modelName, opts.includeSubdomainPaths, opts.includeClassNames)
	if err != nil {
		return false, fmt.Errorf("loading model: %w", err)
	}

	actualSeed := opts.seed
	if actualSeed == 0 {
		actualSeed = time.Now().UnixNano()
	}

	surfaceSpec, err := buildSurfaceSpec(model, opts.includeSubdomainPaths, opts.includeClassNames)
	if err != nil {
		return false, fmt.Errorf("building surface specification: %w", err)
	}

	eng, err := engine.NewSimulationEngine(model, engine.SimulationConfig{
		MaxSteps:        opts.maxSteps,
		RandomSeed:      actualSeed,
		StopOnViolation: opts.stopOnViolation,
		Surface:         surfaceSpec,
	})
	if err != nil {
		return false, fmt.Errorf("creating simulation engine: %w", err)
	}

	result, err := eng.Run()
	if err != nil {
		return false, fmt.Errorf("simulation error: %w", err)
	}

	simTrace := trace.FromResult(result)
	violationReport := report.FromViolations(result.Violations)

	switch opts.output {
	case "json":
		outputJSON(simTrace, violationReport, opts.showTrace, opts.quiet)
	default:
		outputText(simTrace, violationReport, opts.showTrace, opts.quiet, actualSeed)
	}

	return violationReport.HasViolations(), nil
}

func loadModel(rootSource, modelName string, includeSubdomainPaths, includeClassNames []string) (*core.Model, error) {
	if rootSource == "" {
		return nil, fmt.Errorf("-rootsource and -model are required")
	}
	if modelName == "" {
		return nil, fmt.Errorf("model name is required with -rootsource")
	}

	modelPath := filepath.Join(rootSource, modelName)
	parsed, failures, err := parser_human.Parse(modelPath)
	if err != nil {
		return nil, err
	}
	if len(failures) > 0 {
		return nil, fmt.Errorf("model has %d parse failure(s); fix before simulating", len(failures))
	}

	active := &parsed
	if hasSurfaceScope(includeSubdomainPaths, includeClassNames) {
		surfaceSpec, specErr := buildSurfaceSpec(active, includeSubdomainPaths, includeClassNames)
		if specErr != nil {
			return nil, specErr
		}
		active, err = applySurfaceFilter(active, surfaceSpec)
		if err != nil {
			return nil, err
		}
	}
	if err := convert.LowerModel(active); err != nil {
		return nil, err
	}
	return active, nil
}

func applySurfaceFilter(model *core.Model, surfaceSpec *surface.SurfaceSpecification) (*core.Model, error) {
	resolved, err := surface.Resolve(surfaceSpec, model)
	if err != nil {
		return nil, err
	}
	return surface.BuildFilteredModel(model, resolved)
}

func parseCommaSeparatedFlag(flagValue string) []string {
	if strings.TrimSpace(flagValue) == "" {
		return nil
	}
	values := strings.Split(flagValue, ",")
	var trimmed []string
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value != "" {
			trimmed = append(trimmed, value)
		}
	}
	return trimmed
}

func hasSurfaceScope(includeSubdomainPaths, includeClassNames []string) bool {
	return len(includeSubdomainPaths) > 0 || len(includeClassNames) > 0
}

func buildSurfaceSpec(model *core.Model, includeSubdomainPaths, includeClassNames []string) (*surface.SurfaceSpecification, error) {
	if !hasSurfaceScope(includeSubdomainPaths, includeClassNames) {
		return &surface.SurfaceSpecification{}, nil
	}

	spec := &surface.SurfaceSpecification{}
	if len(includeSubdomainPaths) > 0 {
		keys, err := surface.ResolveSubdomainKeysByPath(model, includeSubdomainPaths)
		if err != nil {
			return nil, err
		}
		spec.IncludeSubdomains = keys
	}
	if len(includeClassNames) > 0 {
		keys, err := surface.ResolveClassKeysByName(model, includeClassNames)
		if err != nil {
			return nil, err
		}
		spec.IncludeClasses = keys
	}
	return spec, nil
}

// shouldShowStepTrace reports whether the full step trace belongs in CLI output.
// Clean runs always include steps so incremental surface growth is visible without -trace.
func shouldShowStepTrace(showTrace, quiet bool, hasViolations bool) bool {
	if quiet {
		return false
	}
	return showTrace || !hasViolations
}

func outputText(simTrace *trace.SimulationTrace, violationReport *report.ViolationReport, showTrace, quiet bool, seed int64) {
	if !quiet {
		log.Printf("Simulation completed: %d steps, terminated: %s (seed: %d)\n",
			simTrace.StepsTaken, simTrace.TerminationReason, seed)
	}

	if shouldShowStepTrace(showTrace, quiet, violationReport.HasViolations()) {
		log.Print(simTrace.FormatText())
		log.Println()
	}

	log.Print(violationReport.FormatText())
}

func outputJSON(simTrace *trace.SimulationTrace, violationReport *report.ViolationReport, showTrace, quiet bool) {
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
		log.Printf("Error marshaling output: %v", err)
		os.Exit(1)
	}
	os.Stdout.Write(data)
	os.Stdout.Write([]byte("\n"))
}
