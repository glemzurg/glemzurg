package engine

import (
	"fmt"
	"math/rand"

	"github.com/glemzurg/go-tlaplus/internal/req_model"
	"github.com/glemzurg/go-tlaplus/internal/simulator/actions"
	"github.com/glemzurg/go-tlaplus/internal/simulator/evaluator"
	"github.com/glemzurg/go-tlaplus/internal/simulator/invariants"
	"github.com/glemzurg/go-tlaplus/internal/simulator/state"
	"github.com/glemzurg/go-tlaplus/internal/simulator/surface"
)

// SimulationConfig controls how a simulation run behaves.
type SimulationConfig struct {
	// MaxSteps is the maximum number of simulation steps to run.
	MaxSteps int

	// RandomSeed controls the random number generator for reproducibility.
	RandomSeed int64

	// StopOnViolation stops the simulation at the first violation if true.
	StopOnViolation bool

	// Surface specifies which classes participate in the simulation.
	// nil or empty means "simulate everything" (backward compatible).
	Surface *surface.SurfaceSpecification
}

// SimulationResult captures the outcome of a simulation run.
type SimulationResult struct {
	// Steps holds all simulation steps that were executed.
	Steps []*SimulationStep

	// StepsTaken is the number of steps actually executed.
	StepsTaken int

	// Violations is the combined list of all violations from all steps.
	Violations invariants.ViolationList

	// TerminationReason explains why the simulation stopped.
	// One of: "max_steps", "violation", "deadlock".
	TerminationReason string

	// FinalState is the simulation state when the run ended.
	FinalState *state.SimulationState
}

// SimulationEngine drives the state machine simulation loop.
type SimulationEngine struct {
	config SimulationConfig

	// Core state
	simState        *state.SimulationState
	bindingsBuilder *state.BindingsBuilder

	// Components
	stepExecutor     *StepExecutor
	selector         *ActionSelector
	invariantChecker *invariants.InvariantChecker
	livenessChecker  *LivenessChecker
}

// NewSimulationEngine creates and wires up all simulation components.
func NewSimulationEngine(model *req_model.Model, config SimulationConfig) (*SimulationEngine, error) {
	rng := rand.New(rand.NewSource(config.RandomSeed))

	// Apply surface area filtering.
	activeModel := model
	if config.Surface != nil && !config.Surface.IsEmpty() {
		resolved, err := surface.Resolve(config.Surface, model)
		if err != nil {
			return nil, fmt.Errorf("surface area resolution: %w", err)
		}
		activeModel = surface.BuildFilteredModel(model, resolved)
	}

	// Create simulation state.
	simState := state.NewSimulationState()
	bindingsBuilder := state.NewBindingsBuilder(simState)

	// Register all associations with bindings builder so the evaluator can traverse them.
	for _, assoc := range activeModel.GetClassAssociations() {
		bindingsBuilder.AddAssociation(
			assoc.Key,
			assoc.Name,
			assoc.FromClassKey,
			assoc.ToClassKey,
			evaluator.Multiplicity{
				LowerBound:  assoc.FromMultiplicity.LowerBound,
				HigherBound: assoc.FromMultiplicity.HigherBound,
			},
			evaluator.Multiplicity{
				LowerBound:  assoc.ToMultiplicity.LowerBound,
				HigherBound: assoc.ToMultiplicity.HigherBound,
			},
		)
	}

	// Set up derived attribute evaluation (on-demand computation).
	derivedEval, err := NewDerivedAttributeEvaluator(activeModel, simState, bindingsBuilder.RelationContext())
	if err != nil {
		return nil, fmt.Errorf("derived attribute setup: %w", err)
	}
	if derivedEval.HasDerivedAttributes() {
		bindingsBuilder.SetDerivedResolver(derivedEval)
	}

	// Create invariant checkers.
	invariantChecker, err := invariants.NewInvariantChecker(activeModel)
	if err != nil {
		return nil, fmt.Errorf("invariant checker setup: %w", err)
	}

	dataTypeChecker, dtWarnings := invariants.NewDataTypeChecker(activeModel)
	_ = dtWarnings // Warnings about unparsed data types are informational.

	indexChecker := invariants.NewIndexUniquenessChecker(activeModel)

	// Create action execution components.
	guardEvaluator := actions.NewGuardEvaluator(bindingsBuilder)
	actionExecutor := actions.NewActionExecutor(
		bindingsBuilder, invariantChecker, dataTypeChecker, indexChecker, guardEvaluator, rng,
	)

	paramBinder := actions.NewParameterBinder()

	// Build the class catalog.
	catalog := NewClassCatalog(activeModel)

	if len(catalog.AllSimulatableClasses()) == 0 {
		return nil, fmt.Errorf("no simulatable classes found in model (classes must have states)")
	}

	// Create engine components.
	stateActionExec := NewStateActionExecutor(actionExecutor)
	chainHandler := NewCreationChainHandler(catalog, actionExecutor, stateActionExec, paramBinder, rng)
	multChecker := NewMultiplicityChecker(catalog)
	selector := NewActionSelector(catalog, rng)
	livenessChecker := NewLivenessChecker(catalog)

	stepExecutor := NewStepExecutor(
		actionExecutor, stateActionExec, chainHandler, multChecker, paramBinder, catalog, rng,
	)

	return &SimulationEngine{
		config:           config,
		simState:         simState,
		bindingsBuilder:  bindingsBuilder,
		stepExecutor:     stepExecutor,
		selector:         selector,
		invariantChecker: invariantChecker,
		livenessChecker:  livenessChecker,
	}, nil
}

// Run executes the simulation loop and returns the result.
func (e *SimulationEngine) Run() (*SimulationResult, error) {
	result := &SimulationResult{}

	for step := 0; step < e.config.MaxSteps; step++ {
		// Pick the next action.
		pending, err := e.selector.SelectAction(e.simState)
		if err != nil {
			result.TerminationReason = "deadlock"
			break
		}

		// Execute the step.
		stepResult, err := e.stepExecutor.Execute(pending, e.simState, step+1)
		if err != nil {
			return nil, fmt.Errorf("step %d execution error: %w", step+1, err)
		}

		// Run model-level invariant check after each step.
		modelViolations := e.invariantChecker.CheckModelInvariants(e.simState, e.bindingsBuilder)
		stepResult.Violations = append(stepResult.Violations, modelViolations...)

		result.Steps = append(result.Steps, stepResult)
		result.StepsTaken++
		result.Violations = append(result.Violations, stepResult.Violations...)

		if e.config.StopOnViolation && result.Violations.HasViolations() {
			result.TerminationReason = "violation"
			break
		}
	}

	if result.TerminationReason == "" {
		result.TerminationReason = "max_steps"
	}

	result.FinalState = e.simState

	// Run liveness checks after simulation completes.
	livenessViolations := e.livenessChecker.Check(result)
	result.Violations = append(result.Violations, livenessViolations...)

	return result, nil
}

// State returns the current simulation state (useful for testing).
func (e *SimulationEngine) State() *state.SimulationState {
	return e.simState
}
