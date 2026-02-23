package test_helper

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPruneToModelOnly(t *testing.T) {
	model := GetTestModel()
	pruned := PruneToModelOnly(model)

	err := pruned.Validate()
	assert.Nil(t, err, "pruned model should be valid")

	// Verify children are stripped.
	assert.Nil(t, pruned.ClassAssociations)
	for _, domain := range pruned.Domains {
		assert.Nil(t, domain.ClassAssociations)
		for _, subdomain := range domain.Subdomains {
			assert.Nil(t, subdomain.Classes)
			assert.Nil(t, subdomain.Generalizations)
			assert.Nil(t, subdomain.UseCases)
			assert.Nil(t, subdomain.UseCaseGeneralizations)
			assert.Nil(t, subdomain.ClassAssociations)
			assert.Nil(t, subdomain.UseCaseShares)
		}
	}

	// Verify direct children are kept.
	assert.NotEmpty(t, pruned.Actors)
	assert.NotEmpty(t, pruned.ActorGeneralizations)
	assert.NotEmpty(t, pruned.Domains)
	assert.NotEmpty(t, pruned.DomainAssociations)
	assert.NotEmpty(t, pruned.Invariants)
	assert.NotEmpty(t, pruned.GlobalFunctions)
}

func TestPruneToClassAttributes(t *testing.T) {
	model := GetTestModel()
	pruned := PruneToClassAttributes(model)

	err := pruned.Validate()
	assert.Nil(t, err, "pruned model should be valid")

	// Verify class associations are stripped.
	assert.Nil(t, pruned.ClassAssociations)
	for _, domain := range pruned.Domains {
		assert.Nil(t, domain.ClassAssociations)
		for _, subdomain := range domain.Subdomains {
			assert.Nil(t, subdomain.UseCases)
			assert.Nil(t, subdomain.UseCaseGeneralizations)
			assert.Nil(t, subdomain.ClassAssociations)
			assert.Nil(t, subdomain.UseCaseShares)

			// Verify state machine parts are stripped from classes.
			for _, class := range subdomain.Classes {
				assert.Nil(t, class.States)
				assert.Nil(t, class.Events)
				assert.Nil(t, class.Guards)
				assert.Nil(t, class.Actions)
				assert.Nil(t, class.Queries)
				assert.Nil(t, class.Transitions)
			}
		}
	}
}

func TestPruneToClassAssociations(t *testing.T) {
	model := GetTestModel()
	pruned := PruneToClassAssociations(model)

	err := pruned.Validate()
	assert.Nil(t, err, "pruned model should be valid")

	// Verify use cases are stripped.
	for _, domain := range pruned.Domains {
		for _, subdomain := range domain.Subdomains {
			assert.Nil(t, subdomain.UseCases)
			assert.Nil(t, subdomain.UseCaseGeneralizations)
			assert.Nil(t, subdomain.UseCaseShares)

			// Verify state machine parts are stripped from classes.
			for _, class := range subdomain.Classes {
				assert.Nil(t, class.States)
				assert.Nil(t, class.Events)
				assert.Nil(t, class.Guards)
				assert.Nil(t, class.Actions)
				assert.Nil(t, class.Queries)
				assert.Nil(t, class.Transitions)
			}
		}
	}

	// Verify class associations exist at some level.
	hasAssociations := len(pruned.ClassAssociations) > 0
	for _, domain := range pruned.Domains {
		if len(domain.ClassAssociations) > 0 {
			hasAssociations = true
		}
		for _, subdomain := range domain.Subdomains {
			if len(subdomain.ClassAssociations) > 0 {
				hasAssociations = true
			}
		}
	}
	assert.True(t, hasAssociations, "should have class associations at some level")
}

func TestPruneToStateMachine(t *testing.T) {
	model := GetTestModel()
	pruned := PruneToStateMachine(model)

	err := pruned.Validate()
	assert.Nil(t, err, "pruned model should be valid")

	// Verify use cases are stripped.
	for _, domain := range pruned.Domains {
		for _, subdomain := range domain.Subdomains {
			assert.Nil(t, subdomain.UseCases)
			assert.Nil(t, subdomain.UseCaseGeneralizations)
			assert.Nil(t, subdomain.UseCaseShares)
		}
	}
}

func TestPruneToNoSteps(t *testing.T) {
	model := GetTestModel()
	pruned := PruneToNoSteps(model)

	err := pruned.Validate()
	assert.Nil(t, err, "pruned model should be valid")

	// Verify steps are nil on all scenarios.
	for _, domain := range pruned.Domains {
		for _, subdomain := range domain.Subdomains {
			for _, useCase := range subdomain.UseCases {
				for _, scenario := range useCase.Scenarios {
					assert.Nil(t, scenario.Steps, "scenario steps should be nil")
				}
			}
		}
	}

	// Verify use cases still exist.
	hasUseCases := false
	for _, domain := range pruned.Domains {
		for _, subdomain := range domain.Subdomains {
			if len(subdomain.UseCases) > 0 {
				hasUseCases = true
			}
		}
	}
	assert.True(t, hasUseCases, "should still have use cases")
}
