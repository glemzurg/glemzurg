package engine

import (
	"math/rand"
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_class"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_state"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/actions"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/invariants"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/object"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/state"
	"github.com/stretchr/testify/require"
)

func TestEvenplayCurrencyDuplicateAbbrAddReportsIndexUniquenessViolation(t *testing.T) {
	model := loadEvenplayWalletModel(t)
	currencyClass, addEvent, ok := findCurrencyClassAndAddEvent(model)
	require.True(t, ok)

	simState := state.NewSimulationState()
	bb := state.NewBindingsBuilder(simState)
	require.NoError(t, bb.RegisterNamedSets(model))
	catalog := NewClassCatalog(model)
	registerCatalogAssociations(catalog, bb)
	invChecker, err := invariants.NewInvariantChecker(model)
	require.NoError(t, err)

	ae := actions.NewActionExecutor(
		bb,
		actions.InvariantRuntimeCheckers{Checker: invChecker, DataType: nil},
		&invariants.StructuralInvariantCheckers{
			Index: invariants.NewIndexUniquenessChecker(model),
		},
		actions.NewGuardEvaluator(bb),
		catalog,
		rand.New(rand.NewSource(42)), //nolint:gosec // deterministic test seed
	)

	params := currencyAddParams("USD")

	first, err := ae.ExecuteTransition(
		currencyClass, addEvent, nil, params,
		actions.CreationLinkSource{SourceAssocKey: nil, SourceID: nil}, nil,
	)
	require.NoError(t, err)
	require.True(t, first.WasCreation)
	require.False(t, hasIndexUniquenessViolation(first.Violations), "first Add with abbr USD should not violate index uniqueness: %v", first.Violations)

	second, err := ae.ExecuteTransition(
		currencyClass, addEvent, nil, params,
		actions.CreationLinkSource{SourceAssocKey: nil, SourceID: nil}, nil,
	)
	require.NoError(t, err)
	require.True(t, second.WasCreation)
	require.True(t, hasActionRequiresViolation(second.Violations),
		"second Add with same abbr should fail peer-uniqueness require at assessment, got: %v", second.Violations)
	require.False(t, hasIndexUniquenessViolation(second.Violations),
		"require failure should prevent guarantees from applying, so index uniqueness should not also fire: %v", second.Violations)
	require.Len(t, simState.InstancesByClass(currencyClass.Key), 2,
		"creation still materializes the instance before requires; rollback is not implemented yet")

	for _, v := range second.Violations {
		if v.Type == invariants.ViolationTypeActionRequires {
			require.Contains(t, v.Message, "FALSE")
			return
		}
	}
}

func hasIndexUniquenessViolation(violations invariants.ViolationErrors) bool {
	for _, v := range violations {
		if v.Type == invariants.ViolationTypeIndexUniqueness {
			return true
		}
	}
	return false
}

func hasActionRequiresViolation(violations invariants.ViolationErrors) bool {
	for _, v := range violations {
		if v.Type == invariants.ViolationTypeActionRequires {
			return true
		}
	}
	return false
}

func findCurrencyClassAndAddEvent(model *core.Model) (model_class.Class, model_state.Event, bool) {
	for _, d := range model.Domains {
		for _, sd := range d.Subdomains {
			for _, class := range sd.Classes {
				if class.Name != "Currency" {
					continue
				}
				for _, event := range class.Events {
					if event.Name == "Add" {
						return class, event, true
					}
				}
			}
		}
	}
	return model_class.Class{}, model_state.Event{}, false
}

func currencyAddParams(abbr string) map[string]object.Object {
	return map[string]object.Object{
		"Type":      object.NewString("REAL"),
		"Abbr":      object.NewString(abbr),
		"Name":      object.NewString("T"),
		"ISO":       object.NewString(abbr),
		"Precision": object.NewInteger(2),
	}
}
