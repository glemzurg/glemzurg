package engine

import (
	"fmt"
	"math/rand"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_class"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/actions"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/evaluator"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/instance"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/invariants"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/object"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/schema"
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
	FinalState *instance.State

	// Catalog holds scoped class metadata for trace rendering (association-class endpoints).
	Catalog *schema.Catalog

	// SimulationCoverage records parameter simulation specs that produced values during the run.
	SimulationCoverage *SimulationCoverageTracker
}

// SimulationEngine drives the state machine simulation loop.
type SimulationEngine struct {
	config SimulationConfig

	// Core state
	simState        *instance.State
	bindingsBuilder *state.BindingsBuilder

	// Components
	catalog             *schema.Catalog
	stepExecutor        *StepExecutor
	selector            *ActionSelector
	invariantChecker    *invariants.InvariantChecker
	dataTypeChecker     *invariants.DataTypeChecker
	livenessChecker     *LivenessChecker
	stateMachineChecker *StateMachineChecker
	simulationCoverage  *SimulationCoverageTracker

	// scopeEntries summarize which classes/subdomains participate (include-list scope).
	scopeEntries []surface.ScopeEntry
}

// NewSimulationEngine creates and wires up all simulation components.
// The model must have its ExpressionSpec.Expression fields already populated
// (e.g., via parse functions passed to ExpressionSpec constructors).
//
// Data-flow gate: model + surface resolve into schema.New(fullModel, RunScope).
// After that, the run is driven from *schema.Schema only (no dual model authority).
func NewSimulationEngine(model *core.Model, config SimulationConfig) (*SimulationEngine, error) {
	rng := newSimulationRNG(config.RandomSeed)

	sch, unavailable, scopeEntries, err := buildRunSchema(model, config)
	if err != nil {
		return nil, err
	}

	catalog := setupSimulationCatalog(sch)
	catalog.SetSurfaceUnavailableMembers(unavailable)

	core, err := wireSimulationCore(sch, catalog, rng)
	if err != nil {
		return nil, err
	}
	includeOutOfScopeExtents(core, catalog)

	return newWiredSimulationEngine(config, catalog, core, scopeEntries), nil
}

// buildRunSchema resolves surface scope, builds schema from the full model + RunScope,
// and overlays surface-scoped class bodies / model invariants when a surface is set.
func buildRunSchema(
	model *core.Model,
	config SimulationConfig,
) (*schema.Schema, []surface.UnavailableMember, []surface.ScopeEntry, error) {
	if err := validateSimulationModel(model); err != nil {
		return nil, nil, nil, err
	}
	if config.Surface == nil || config.Surface.IsEmpty() {
		scopeEntries := surface.BuildScopeEntries(model, surface.AllNonRealizedClasses(model))
		return schema.New(model, schema.RunScopeAll()), nil, scopeEntries, nil
	}
	return buildRunSchemaWithSurface(model, config.Surface)
}

func buildRunSchemaWithSurface(
	model *core.Model,
	spec *surface.SurfaceSpecification,
) (*schema.Schema, []surface.UnavailableMember, []surface.ScopeEntry, error) {
	resolved, err := surface.Resolve(spec, model)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("surface area resolution: %w", err)
	}
	scopeEntries := surface.BuildScopeEntries(model, resolved.Classes)
	sch := schema.New(model, schema.NewRunScope(resolvedClassKeys(resolved)))
	if err := applySurfaceClassOverlays(sch, model, resolved); err != nil {
		return nil, nil, nil, err
	}
	return sch, resolved.UnavailableMembers, scopeEntries, nil
}

func resolvedClassKeys(resolved *surface.ResolvedSurface) []identity.Key {
	keys := make([]identity.Key, 0, len(resolved.Classes))
	for k := range resolved.Classes {
		keys = append(keys, k)
	}
	return keys
}

// applySurfaceClassOverlays installs surface-scoped class bodies and model invariants.
func applySurfaceClassOverlays(sch *schema.Schema, model *core.Model, resolved *surface.ResolvedSurface) error {
	filtered, err := surface.BuildFilteredModel(model, resolved)
	if err != nil {
		return fmt.Errorf("build filtered model: %w", err)
	}
	if err := validateSimulationModel(filtered); err != nil {
		return err
	}
	for _, domain := range filtered.Domains {
		for _, subdomain := range domain.Subdomains {
			for _, class := range subdomain.Classes {
				if err := sch.ReplaceInScopeClass(class); err != nil {
					return fmt.Errorf("install scoped class %s: %w", class.Key.String(), err)
				}
			}
		}
	}
	sch.SetModelInvariants(resolved.ModelInvariants)
	return nil
}

// includeOutOfScopeExtents lets invariant evaluation bind empty sets for OOS class names.
func includeOutOfScopeExtents(core *simulationCore, catalog *schema.Catalog) {
	if core == nil || core.checkers == nil || core.checkers.invariantChecker == nil {
		return
	}
	core.checkers.invariantChecker.IncludeClassExtents(catalog.ClassNameMap())
}

func newWiredSimulationEngine(
	config SimulationConfig,
	catalog *schema.Catalog,
	core *simulationCore,
	scopeEntries []surface.ScopeEntry,
) *SimulationEngine {
	return &SimulationEngine{
		config:              config,
		simState:            core.simState,
		bindingsBuilder:     core.bindingsBuilder,
		catalog:             catalog,
		stepExecutor:        core.stepExecutor,
		selector:            core.selector,
		invariantChecker:    core.checkers.invariantChecker,
		dataTypeChecker:     core.checkers.dataTypeChecker,
		livenessChecker:     core.livenessChecker,
		stateMachineChecker: NewStateMachineChecker(catalog),
		simulationCoverage:  core.simulationCoverage,
		scopeEntries:        scopeEntries,
	}
}

// simulationCore holds wired runtime components after catalog setup.
type simulationCore struct {
	simState           *instance.State
	bindingsBuilder    *state.BindingsBuilder
	stepExecutor       *StepExecutor
	selector           *ActionSelector
	checkers           *simulationCheckers
	livenessChecker    *LivenessChecker
	simulationCoverage *SimulationCoverageTracker
}

func wireSimulationCore(
	sch *schema.Schema,
	catalog *schema.Catalog,
	rng *rand.Rand,
) (*simulationCore, error) {
	evalCtx, err := sch.NewEvalContext()
	if err != nil {
		return nil, fmt.Errorf("expression registry setup: %w", err)
	}

	simState, bindingsBuilder, derivedEval, err := setupState(sch, catalog, evalCtx)
	if err != nil {
		return nil, err
	}
	if derivedEval != nil {
		derivedEval.SetCatalog(catalog)
	}

	checkers, err := setupCheckers(sch, evalCtx)
	if err != nil {
		return nil, err
	}

	simulationCoverage := NewSimulationCoverageTracker()
	stepExecutor, selector, livenessChecker, err := setupExecutors(executorSetupDeps{
		bindingsBuilder:    bindingsBuilder,
		derivedEval:        derivedEval,
		checkers:           checkers,
		catalog:            catalog,
		rng:                rng,
		simulationCoverage: simulationCoverage,
	})
	if err != nil {
		return nil, err
	}

	return &simulationCore{
		simState:           simState,
		bindingsBuilder:    bindingsBuilder,
		stepExecutor:       stepExecutor,
		selector:           selector,
		checkers:           checkers,
		livenessChecker:    livenessChecker,
		simulationCoverage: simulationCoverage,
	}, nil
}

func newSimulationRNG(seed int64) *rand.Rand {
	return rand.New(rand.NewSource(seed)) //nolint:gosec // simulation uses deterministic seeded RNG
}

func setupSimulationCatalog(sch *schema.Schema) *schema.Catalog {
	// Catalog (caller graphs, association nav, extents) is built inside schema.New.
	return sch.Catalog()
}

// setupState creates simulation state and bindings builder, registers associations,
// and sets up derived attribute evaluation. sch is the sole static model for the run.
func setupState(
	sch *schema.Schema,
	catalog *schema.Catalog,
	evalCtx *evaluator.EvalContext,
) (*instance.State, *state.BindingsBuilder, *DerivedAttributeEvaluator, error) {
	simState := instance.NewState(sch)
	bindingsBuilder := state.NewBindingsBuilder(simState)

	registerCatalogAssociations(catalog, bindingsBuilder)

	derivedEval, err := NewDerivedAttributeEvaluator(sch, bindingsBuilder, evalCtx)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("derived attribute setup: %w", err)
	}
	if derivedEval.HasDerivedAttributes() {
		bindingsBuilder.SetDerivedResolver(derivedEval)
	}

	if err := bindingsBuilder.RegisterNamedSets(sch); err != nil {
		return nil, nil, nil, fmt.Errorf("named set setup: %w", err)
	}

	return simState, bindingsBuilder, derivedEval, nil
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

// setupCheckers constructs constraint checkers from schema (no *core.Model).
func setupCheckers(sch *schema.Schema, evalCtx *evaluator.EvalContext) (*simulationCheckers, error) {
	invariantChecker, err := invariants.NewInvariantChecker(sch)
	if err != nil {
		return nil, fmt.Errorf("invariant checker setup: %w", err)
	}
	invariantChecker.SetEvalContext(evalCtx)

	dataTypeChecker, _ := invariants.NewDataTypeChecker(sch)
	indexChecker := invariants.NewIndexUniquenessChecker(sch)
	multChecker := invariants.NewMultiplicityChecker(sch)
	assocInstancePairChecker := invariants.NewAssociationInstancePairChecker(sch)
	assocUniquenessChecker := invariants.NewAssociationUniquenessChecker(sch)
	associationInvChecker, err := invariants.NewAssociationInvariantChecker(sch)
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

func registerCatalogAssociations(catalog *schema.Catalog, bindingsBuilder *state.BindingsBuilder) {
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
		// Association-class host only when the AC class is on the surface; otherwise plain.
		if assoc.AssociationClassKey != nil {
			if linkInfo := catalog.GetClassInfo(*assoc.AssociationClassKey); linkInfo != nil {
				bindingsBuilder.AddAssociationClassHost(
					assoc.Key,
					assoc.Name,
					evaluator.AssociationHostEndpoints{
						FromClassKey: assoc.FromClassKey.String(),
						ToClassKey:   assoc.ToClassKey.String(),
					},
					linkInfo.Class.Name,
					evaluator.AssociationHostMultiplicities{From: fromMult, To: toMult},
				)
				continue
			}
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

type executorSetupDeps struct {
	bindingsBuilder    *state.BindingsBuilder
	derivedEval        *DerivedAttributeEvaluator
	checkers           *simulationCheckers
	catalog            *schema.Catalog
	rng                *rand.Rand
	simulationCoverage *SimulationCoverageTracker
}

// setupExecutors creates step executor, action selector, and liveness checker.
func setupExecutors(deps executorSetupDeps) (*StepExecutor, *ActionSelector, *LivenessChecker, error) {
	actionExecutor := buildActionExecutor(deps.bindingsBuilder, deps.checkers, deps.catalog, deps.rng)

	if len(deps.catalog.AllEventBearingClasses()) == 0 {
		return nil, nil, nil, fmt.Errorf("no event-bearing simulatable classes found in model")
	}

	stepExecutor, selector, livenessChecker := buildStepExecutor(
		actionExecutor, deps.bindingsBuilder, deps.derivedEval, deps.catalog, deps.rng, deps.simulationCoverage,
	)
	return stepExecutor, selector, livenessChecker, nil
}

// buildActionExecutor creates the action executor with its dependencies.
func buildActionExecutor(
	bindingsBuilder *state.BindingsBuilder,
	checkers *simulationCheckers,
	catalog *schema.Catalog,
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
		guardEvaluator, catalog.Schema(), rng,
	)
}

// buildStepParameterGenerator creates surface and nested parameter generators from model named sets.
func buildStepParameterGenerator(
	bindingsBuilder *state.BindingsBuilder,
	catalog *schema.Catalog,
) (*actions.ParameterBinder, *StepParameterGenerator) {
	paramBinder := actions.NewParameterBinder()
	wireParameterLookups(paramBinder, bindingsBuilder, catalog)
	paramSampler := actions.NewParameterSampler(paramBinder, bindingsBuilder.NamedSetValues())
	wirePeerFieldDistinctLookup(paramSampler, bindingsBuilder)
	return paramBinder, NewStepParameterGenerator(paramBinder, paramSampler)
}

func wireParameterLookups(
	paramBinder *actions.ParameterBinder,
	bindingsBuilder *state.BindingsBuilder,
	catalog *schema.Catalog,
) {
	if paramBinder == nil || catalog == nil {
		return
	}
	paramBinder.SetObjectInstanceLookup(func(objectClassRef string) []object.Object {
		return objectInstancesForClassRef(bindingsBuilder.State(), catalog, objectClassRef)
	})
}

func wirePeerFieldDistinctLookup(
	paramSampler *actions.ParameterSampler,
	bindingsBuilder *state.BindingsBuilder,
) {
	if paramSampler == nil {
		return
	}
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
}

// objectInstancesForClassRef returns extent elements for in-scope instances matching
// an object-of class reference (subkey, display name, or TLA name).
func objectInstancesForClassRef(
	simState *instance.State,
	catalog *schema.Catalog,
	objectClassRef string,
) []object.Object {
	if simState == nil || catalog == nil || objectClassRef == "" {
		return nil
	}
	want := identity.NormalizeSubKey(objectClassRef)
	var out []object.Object
	for _, info := range catalog.AllScopedClasses() {
		if !objectClassRefMatches(want, objectClassRef, info) {
			continue
		}
		for _, inst := range simState.InstancesByClass(info.ClassKey) {
			out = append(out, state.ClassExtentElement(inst.ID, inst.Attributes))
		}
		return out
	}
	return nil
}

func objectClassRefMatches(wantNorm, objectClassRef string, info *schema.ClassSimInfo) bool {
	if info == nil {
		return false
	}
	if info.ClassKey.SubKey == objectClassRef || info.ClassKey.String() == objectClassRef {
		return true
	}
	if identity.NormalizeSubKey(info.Class.Name) == wantNorm {
		return true
	}
	if model_class.ClassTLAName(info.Class.Name) == objectClassRef {
		return true
	}
	return identity.NormalizeSubKey(model_class.ClassTLAName(info.Class.Name)) == wantNorm
}

// buildStepExecutor creates the step executor, action selector, and liveness checker.
func buildStepExecutor(
	actionExecutor *actions.ActionExecutor,
	bindingsBuilder *state.BindingsBuilder,
	derivedEval *DerivedAttributeEvaluator,
	catalog *schema.Catalog,
	rng *rand.Rand,
	simulationCoverage *SimulationCoverageTracker,
) (*StepExecutor, *ActionSelector, *LivenessChecker) {
	paramBinder, paramGen := buildStepParameterGenerator(bindingsBuilder, catalog)
	stateActionExec := NewStateActionExecutor(actionExecutor)
	chainHandler := NewCreationChainHandler(catalog, actionExecutor, stateActionExec, paramBinder, rng)
	stepExecutor := NewStepExecutor(StepExecutorDeps{
		ActionExecutor:     actionExecutor,
		StateActionExec:    stateActionExec,
		ChainHandler:       chainHandler,
		ParamGen:           paramGen,
		Catalog:            catalog,
		DerivedEval:        derivedEval,
		RNG:                rng,
		SimulationCoverage: simulationCoverage,
		BindingsBuilder:    bindingsBuilder,
	})

	return stepExecutor, NewActionSelector(catalog, derivedEval, bindingsBuilder, paramGen.Sampler, rng), NewLivenessChecker(catalog)
}

// Run executes the simulation loop and returns the result.
func (e *SimulationEngine) Run() (*SimulationResult, error) {
	result := &SimulationResult{}
	domainExhaustedSkips := 0

	for step := range e.config.MaxSteps {
		// Pick the next action.
		pending, err := e.selector.SelectAction(e.simState)
		if err != nil {
			result.TerminationReason = "deadlock"
			break
		}

		// Execute the step (association structural checks run after nested work inside the step).
		stepResult, err := e.stepExecutor.Execute(pending, e.simState, step+1)
		if err != nil {
			// Domain exhausted after selection: skip and reselect (eligibility filter
			// should prevent this; soft-skip avoids hard failure if state raced).
			if isNamedSetDomainExhaustedError(err) {
				domainExhaustedSkips++
				if domainExhaustedSkips > e.config.MaxSteps {
					result.TerminationReason = "deadlock"
					break
				}
				continue
			}
			return nil, fmt.Errorf("step %d execution error: %w", step+1, err)
		}
		domainExhaustedSkips = 0

		// Class/attribute invariants after the full step graph is built (including nesting).
		// Model + association structural checks run in the step executor after nesting.
		stepResult.Violations = append(stepResult.Violations, e.invariantChecker.CheckClassInvariants(e.simState, e.bindingsBuilder)...)
		stepResult.Violations = append(stepResult.Violations, e.invariantChecker.CheckAttributeInvariants(e.simState, e.bindingsBuilder)...)

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
	result.SimulationCoverage = e.simulationCoverage

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
func (e *SimulationEngine) State() *instance.State {
	return e.simState
}

// SurfaceReport returns simulation scope plus external drivers for this run.
func (e *SimulationEngine) SurfaceReport() *SurfaceReport {
	report := BuildSurfaceReport(e.catalog)
	report.Scope = append([]surface.ScopeEntry(nil), e.scopeEntries...)
	return report
}
