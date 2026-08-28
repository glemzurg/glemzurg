package model_bridge

import (
	"fmt"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/registry"
)

// LoadResult contains the result of loading all definitions from a model.
type LoadResult struct {
	// Registry is the populated registry containing all loaded definitions.
	Registry *registry.Registry

	// Results contains the individual build results for each expression.
	Results []*BuildResult

	// Errors contains all build errors that occurred.
	Errors []error
}

// HasErrors returns true if any errors occurred during loading.
func (r *LoadResult) HasErrors() bool {
	return len(r.Errors) > 0
}

// successCount returns the number of successfully loaded definitions.
func (r *LoadResult) successCount() int {
	count := 0
	for _, result := range r.Results {
		if result.IsSuccess() {
			count++
		}
	}
	return count
}

// errorCount returns the number of failed definitions.
func (r *LoadResult) errorCount() int {
	return len(r.Errors)
}

// definitionsBySource returns all successful definitions grouped by their source type.
func (r *LoadResult) definitionsBySource() map[ExpressionSource][]*registry.Definition {
	bySource := make(map[ExpressionSource][]*registry.Definition)
	for _, result := range r.Results {
		if result.IsSuccess() {
			bySource[result.Source.Source] = append(bySource[result.Source.Source], result.Definition)
		}
	}
	return bySource
}

// Loader loads TLA+ definitions from a model into a registry.
type Loader struct {
	builder *DefinitionBuilder
}

// NewLoader creates a new Loader.
func NewLoader() *Loader {
	return &Loader{
		builder: NewDefinitionBuilder(),
	}
}

// LoadFromModel extracts all TLA+ expressions from a model, parses them,
// and registers them in a new registry.
// Returns the populated registry and any errors that occurred.
func (l *Loader) LoadFromModel(model *core.Model) *LoadResult {
	result := &LoadResult{
		Registry: registry.NewRegistry(),
		Results:  make([]*BuildResult, 0),
		Errors:   make([]error, 0),
	}

	// Extract all expressions from the model
	expressions := ExtractFromModel(model)

	// Build each expression
	for _, expr := range expressions {
		buildResult := l.builder.Build(expr, result.Registry)
		result.Results = append(result.Results, buildResult)

		if buildResult.Error != nil {
			result.Errors = append(result.Errors, buildResult.Error)
		}
	}

	return result
}

// loadFromExpressions parses and registers a list of extracted expressions
// into a new registry.
func (l *Loader) loadFromExpressions(expressions []ExtractedExpression) *LoadResult {
	result := &LoadResult{
		Registry: registry.NewRegistry(),
		Results:  make([]*BuildResult, 0),
		Errors:   make([]error, 0),
	}

	// Build each expression
	for _, expr := range expressions {
		buildResult := l.builder.Build(expr, result.Registry)
		result.Results = append(result.Results, buildResult)

		if buildResult.Error != nil {
			result.Errors = append(result.Errors, buildResult.Error)
		}
	}

	return result
}

// loadIntoRegistry parses and registers expressions into an existing registry.
// This is useful for incremental loading.
func (l *Loader) loadIntoRegistry(expressions []ExtractedExpression, reg *registry.Registry) *LoadResult {
	result := &LoadResult{
		Registry: reg,
		Results:  make([]*BuildResult, 0),
		Errors:   make([]error, 0),
	}

	// Build each expression
	for _, expr := range expressions {
		buildResult := l.builder.Build(expr, reg)
		result.Results = append(result.Results, buildResult)

		if buildResult.Error != nil {
			result.Errors = append(result.Errors, buildResult.Error)
		}
	}

	return result
}

// mustLoadFromModel is like LoadFromModel but panics if any errors occur.
// Useful for tests and static initialization.
func (l *Loader) mustLoadFromModel(model *core.Model) *LoadResult {
	result := l.LoadFromModel(model)
	if result.HasErrors() {
		panic(fmt.Sprintf("failed to load model: %v", result.Errors))
	}
	return result
}

// loadFromModelStrict loads from a model and returns an error if any
// expression fails to load.
func (l *Loader) loadFromModelStrict(model *core.Model) (*LoadResult, error) {
	result := l.LoadFromModel(model)
	if result.HasErrors() {
		return result, fmt.Errorf("failed to load %d expressions: %w", result.errorCount(), result.Errors[0])
	}
	return result, nil
}
