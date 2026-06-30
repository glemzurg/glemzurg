package engine

import (
	"fmt"
	"math/rand"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/actions"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/evaluator"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/invariants"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/object"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/state"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/surface"
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
	Violations invariants.ViolationErrors

	// TerminationReason explains why the simulation stopped.
	// One of: "max_steps", "violation", "deadlock".
	TerminationReason string

	// FinalState is the simulation state when the run ended.
	FinalState *state.SimulationState

	// Catalog holds scoped class metadata for trace rendering (association-class endpoints).
	Catalog *ClassCatalog
}

// SimulationEngine drives the state machine simulation loop.
type SimulationEngine struct {
	config SimulationConfig

	// Core state
	simState        *state.SimulationState
	bindingsBuilder *state.BindingsBuilder

	// Components
	catalog             *ClassCatalog
	stepExecutor        *StepExecutor
	selector            *ActionSelector
	invariantChecker    *invariants.InvariantChecker
	dataTypeChecker     *invariants.DataTypeChecker
	livenessChecker     *LivenessChecker
	stateMachineChecker *StateMachineChecker
}

// NewSimulationEngine creates and wires up all simulation components.
// The model must have its ExpressionSpec.Expression fields already populated
// (e.g., via parse functions passed to ExpressionSpec constructors).
func NewSimulationEngine(model *core.Model, config SimulationConfig) (*SimulationEngine, error) {
	rng := rand.New(rand.NewSource(config.RandomSeed)) //nolint:gosec // simulation uses deterministic seeded RNG

	activeModel, err := resolveActiveModel(model, config)
	if err != nil {
		return nil, err
	}

	if err := validateSimulationModel(activeModel); err != nil {
		return nil, err
	}

	catalog := NewClassCatalog(activeModel)
	PopulateCallerDataFromModel(activeModel, catalog)

	simState, bindingsBuilder, err := setupState(activeModel, catalog)
	if err != nil {
		return nil, err
	}

	checkers, err := setupCheckers(activeModel)
	if err != nil {
		return nil, err
	}

	stepExecutor, selector, livenessChecker, err := setupExecutors(
		activeModel, bindingsBuilder, checkers, catalog, rng,
	)
	if err != nil {
		return nil, err
	}

	return &SimulationEngine{
		config:              config,
		simState:            simState,
		bindingsBuilder:     bindingsBuilder,
		catalog:             catalog,
		stepExecutor:        stepExecutor,
		selector:            selector,
		invariantChecker:    checkers.invariantChecker,
		dataTypeChecker:     checkers.dataTypeChecker,
		livenessChecker:     livenessChecker,
		stateMachineChecker: NewStateMachineChecker(catalog),
	}, nil
}

// resolveActiveModel applies surface area filtering if configured.
func resolveActiveModel(model *core.Model, config SimulationConfig) (*core.Model, error) {
	if config.Surface == nil || config.Surface.IsEmpty() {
		return model, nil
	}
	resolved, err := surface.Resolve(config.Surface, model)
	if err != nil {
		return nil, fmt.Errorf("surface area resolution: %w", err)
	}
	filtered, err := surface.BuildFilteredModel(model, resolved)
	if err != nil {
		return nil, fmt.Errorf("build filtered model: %w", err)
	}
	return filtered, nil
}

// setupState creates simulation state and bindings builder, registers associations,
// and sets up derived attribute evaluation.
func setupState(model *core.Model, catalog *ClassCatalog) (*state.SimulationState, *state.BindingsBuilder, error) {
	simState := state.NewSimulationState()
	bindingsBuilder := state.NewBindingsBuilder(simState)

	registerCatalogAssociations(catalog, bindingsBuilder)

	// Set up derived attribute evaluation (on-demand computation).
	derivedEval, err := NewDerivedAttributeEvaluator(model, simState, bindingsBuilder.RelationContext())
	if err != nil {
		return nil, nil, fmt.Errorf("derived attribute setup: %w", err)
	}
	if derivedEval.HasDerivedAttributes() {
		bindingsBuilder.SetDerivedResolver(derivedEval)
	}

	if err := bindingsBuilder.RegisterNamedSets(model); err != nil {
		return nil, nil, fmt.Errorf("named set setup: %w", err)
	}

	return simState, bindingsBuilder, nil
}

// simulationCheckers groups all invariant/constraint checkers.
type simulationCheckers struct {
	invariantChecker         *invariants.InvariantChecker
	dataTypeChecker          *invariants.DataTypeChecker
	indexChecker             *invariants.IndexUniquenessChecker
	multChecker              *invariants.MultiplicityChecker
	assocInstancePairChecker *invariants.AssociationInstancePairChecker
	assocUniquenessChecker   *invariants.AssociationUniquenessChecker
	associationInvChecker    *invariants.AssociationInvariantChecker
}

// setupCheckers creates all invariant and constraint checkers.
func setupCheckers(model *core.Model) (*simulationCheckers, error) {
	invariantChecker, err := invariants.NewInvariantChecker(model)
	if err != nil {
		return nil, fmt.Errorf("invariant checker setup: %w", err)
	}

	dataTypeChecker, _ := invariants.NewDataTypeChecker(model)

	indexChecker := invariants.NewIndexUniquenessChecker(model)
	multChecker := invariants.NewMultiplicityChecker(model)
	assocInstancePairChecker := invariants.NewAssociationInstancePairChecker(model)
	assocUniquenessChecker := invariants.NewAssociationUniquenessChecker(model)
	associationInvChecker, err := invariants.NewAssociationInvariantChecker(model)
	if err != nil {
		return nil, fmt.Errorf("association invariant checker setup: %w", err)
	}

	return &simulationCheckers{
		invariantChecker:         invariantChecker,
		dataTypeChecker:          dataTypeChecker,
		indexChecker:             indexChecker,
		multChecker:              multChecker,
		assocInstancePairChecker: assocInstancePairChecker,
		assocUniquenessChecker:   assocUniquenessChecker,
		associationInvChecker:    associationInvChecker,
	}, nil
}

func registerCatalogAssociations(catalog *ClassCatalog, bindingsBuilder *state.BindingsBuilder) {
	for _, ai := range catalog.AllAssociations() {
		assoc := ai.Association
		fromMult := evaluator.Multiplicity{
			LowerBound:  assoc.FromMultiplicity.LowerBound,
			HigherBound: assoc.FromMultiplicity.HigherBound,
		}
		toMult := evaluator.Multiplicity{
			LowerBound:  assoc.ToMultiplicity.LowerBound,
			HigherBound: assoc.ToMultiplicity.HigherBound,
		}
		if assoc.AssociationClassKey != nil {
			linkClassName := ""
			if linkInfo := catalog.GetClassInfo(*assoc.AssociationClassKey); linkInfo != nil {
				linkClassName = linkInfo.Class.Name
			}
			bindingsBuilder.AddAssociationClassHost(
				assoc.Key,
				assoc.Name,
				evaluator.AssociationHostEndpoints{
					FromClassKey: assoc.FromClassKey.String(),
					ToClassKey:   assoc.ToClassKey.String(),
				},
				linkClassName,
				evaluator.AssociationHostMultiplicities{From: fromMult, To: toMult},
			)
			continue
		}
		bindingsBuilder.AddAssociation(
			assoc.Key,
			assoc.Name,
			assoc.FromClassKey,
			assoc.ToClassKey,
			fromMult,
			toMult,
		)
	}
}

// setupExecutors creates step executor, action selector, and liveness checker.
func setupExecutors(
	_ *core.Model,
	bindingsBuilder *state.BindingsBuilder,
	checkers *simulationCheckers,
	catalog *ClassCatalog,
	rng *rand.Rand,
) (*StepExecutor, *ActionSelector, *LivenessChecker, error) {
	actionExecutor := buildActionExecutor(bindingsBuilder, checkers, catalog, rng)

	if len(catalog.AllEventBearingClasses()) == 0 {
		return nil, nil, nil, fmt.Errorf("no event-bearing simulatable classes found in model")
	}

	stepExecutor, selector, livenessChecker := buildStepExecutor(actionExecutor, bindingsBuilder, catalog, rng)
	return stepExecutor, selector, livenessChecker, nil
}

// buildActionExecutor creates the action executor with its dependencies.
func buildActionExecutor(
	bindingsBuilder *state.BindingsBuilder,
	checkers *simulationCheckers,
	catalog *ClassCatalog,
	rng *rand.Rand,
) *actions.ActionExecutor {
	guardEvaluator := actions.NewGuardEvaluator(bindingsBuilder)
	structuralCheckers := &invariants.StructuralInvariantCheckers{
		Index:                   checkers.indexChecker,
		Multiplicity:            checkers.multChecker,
		AssociationInstancePair: checkers.assocInstancePairChecker,
		AssociationUniqueness:   checkers.assocUniquenessChecker,
		AssociationInvariants:   checkers.associationInvChecker,
	}
	return actions.NewActionExecutor(
		bindingsBuilder,
		actions.InvariantRuntimeCheckers{Checker: checkers.invariantChecker, DataType: checkers.dataTypeChecker},
		structuralCheckers,
		guardEvaluator, catalog, rng,
	)
}

// buildStepParameterGenerator creates surface and nested parameter generators from model named sets.
func buildStepParameterGenerator(bindingsBuilder *state.BindingsBuilder) (*actions.ParameterBinder, *StepParameterGenerator) {
	paramBinder := actions.NewParameterBinder()
	paramSampler := actions.NewParameterSampler(paramBinder, bindingsBuilder.NamedSetValues())
	paramSampler.SetPeerFieldDistinctLookup(func(classKey identity.Key, fieldSubKey string) []object.Object {
		var values []object.Object
		excludeID := paramSampler.PeerFieldDistinctExcludeInstanceID()
		for _, inst := range bindingsBuilder.State().InstancesByClass(classKey) {
			if excludeID != 0 && inst.ID == excludeID {
				continue
			}
			values = append(values, inst.GetAttribute(fieldSubKey))
		}
		return values
	})
	return paramBinder, NewStepParameterGenerator(paramBinder, paramSampler)
}

// buildStepExecutor creates the step executor, action selector, and liveness checker.
func buildStepExecutor(
	actionExecutor *actions.ActionExecutor,
	bindingsBuilder *state.BindingsBuilder,
	catalog *ClassCatalog,
	rng *rand.Rand,
) (*StepExecutor, *ActionSelector, *LivenessChecker) {
	paramBinder, paramGen := buildStepParameterGenerator(bindingsBuilder)
	stateActionExec := NewStateActionExecutor(actionExecutor)
	chainHandler := NewCreationChainHandler(catalog, actionExecutor, stateActionExec, paramBinder, rng)
	stepExecutor := NewStepExecutor(
		actionExecutor, stateActionExec, chainHandler, paramGen, catalog, rng,
	)

	return stepExecutor, NewActionSelector(catalog, rng), NewLivenessChecker(catalog)
}

// Run executes the simulation loop and returns the result.
func (e *SimulationEngine) Run() (*SimulationResult, error) {
	result := &SimulationResult{}

	for step := range e.config.MaxSteps {
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

		// Run model-, class-, and attribute-level invariant checks after each step.
		modelViolations := e.invariantChecker.CheckModelInvariants(e.simState, e.bindingsBuilder)
		stepResult.Violations = append(stepResult.Violations, modelViolations...)
		classViolations := e.invariantChecker.CheckClassInvariants(e.simState, e.bindingsBuilder)
		stepResult.Violations = append(stepResult.Violations, classViolations...)
		attrViolations := e.invariantChecker.CheckAttributeInvariants(e.simState, e.bindingsBuilder)
		stepResult.Violations = append(stepResult.Violations, attrViolations...)

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
	result.Catalog = e.catalog

	if e.dataTypeChecker != nil {
		result.Violations = append(result.Violations, e.dataTypeChecker.UnparsedAttributeDefinitionViolations()...)
	}

	// Run post-simulation model checks.
	result.Violations = append(result.Violations, e.stateMachineChecker.Check()...)

	// Run liveness checks after simulation completes.
	livenessViolations := e.livenessChecker.Check(result)
	result.Violations = append(result.Violations, livenessViolations...)

	return result, nil
}

// State returns the current simulation state (useful for testing).
func (e *SimulationEngine) State() *state.SimulationState {
	return e.simState
}
