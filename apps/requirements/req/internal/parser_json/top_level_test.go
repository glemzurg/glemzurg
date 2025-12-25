package parser_json

import (
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/requirements"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/requirements/data_type"
	"github.com/stretchr/testify/assert"
)

func TestUnpackPackRequirementsRoundTrip(t *testing.T) {

	// What are we writing?
	original := requirements.Requirements{

		// Model
		Model: requirements.Model{
			Key:  "model_key",
			Name: "Model",
		},

		// Generalizations.
		Generalizations: []requirements.Generalization{
			{
				Key:  "model_key/generalization/generalization_a",
				Name: "GeneralizationA",
			},
		},

		// Actors.
		Actors: []requirements.Actor{
			{
				Key:  "model_key/actor/actor_a",
				Name: "ActorA",
				Type: "person",
			},
		},

		// Organization.
		Domains: []requirements.Domain{
			{
				Key:  "domain_key_a",
				Name: "DomainA",
			},
			{
				Key:  "domain_key_b",
				Name: "DomainB",
			},
		},
		Subdomains: map[string][]requirements.Subdomain{
			"domain_key_a": {
				{
					Key:  "domain_key_a/subdomain_aa",
					Name: "SubdomainAA",
				},
				{
					Key:  "domain_key_a/subdomain_ab",
					Name: "SubdomainAB",
				},
			},
			"domain_key_b": {
				{
					Key:  "domain_key_b/subdomain_ba",
					Name: "SubdomainBA",
				},
				{
					Key:  "domain_key_b/subdomain_bb",
					Name: "SubdomainBB",
				},
			},
		},
		DomainAssociations: []requirements.DomainAssociation{
			{
				Key:               "model_key/domain_association/1",
				ProblemDomainKey:  "domain_key_a",
				SolutionDomainKey: "domain_key_b",
			},
		},

		// Classes.
		Classes: map[string][]requirements.Class{
			"domain_key_a/subdomain_aa": {
				{
					Key:  "domain_key_a/subdomain_aa/class_a",
					Name: "ClassA",
				},
				{
					Key:  "domain_key_a/subdomain_aa/class_b",
					Name: "ClassB",
				},
			},
			"domain_key_b/subdomain_bb": {
				{
					Key:  "domain_key_b/subdomain_ba/class_a",
					Name: "ClassA",
				},
				{
					Key:  "domain_key_b/subdomain_bb/class_b",
					Name: "ClassB",
				},
			},
		},
		Attributes: map[string][]requirements.Attribute{
			"domain_key_a/subdomain_aa/class_a": {
				{
					Key:           "domain_key_a/subdomain_aa/class_a/attribute_a",
					Name:          "AttributeA",
					DataTypeRules: "unconstrained",
					DataType: &data_type.DataType{
						Key:            "domain_key_a/subdomain_aa/class_a/attribute_a",
						CollectionType: "atomic",
						Atomic: &data_type.Atomic{
							ConstraintType: "unconstrained",
						},
					},
				},
				{
					Key:  "domain_key_a/subdomain_aa/class_a/attribute_b",
					Name: "AttributeB",
				},
			},
			"domain_key_b/subdomain_ba/class_a": {
				{
					Key:  "domain_key_b/subdomain_ba/class_a/attribute_a",
					Name: "AttributeA",
				},
			},
			"domain_key_b/subdomain_bb/class_b": {
				{
					Key:  "domain_key_b/subdomain_bb/class_b/attribute_b",
					Name: "AttributeB",
				},
			},
		},
		Associations: []requirements.Association{
			{
				Key:              "model_key/association/1",
				Name:             "Child",
				FromClassKey:     "domain_key_a/subdomain_aa/class_a",
				FromMultiplicity: requirements.Multiplicity{LowerBound: 0, HigherBound: 1},
				ToClassKey:       "domain_key_b/subdomain_ba/class_a",
				ToMultiplicity:   requirements.Multiplicity{LowerBound: 2, HigherBound: 3},
			},
			{
				Key:              "model_key/association/2",
				Name:             "Parent",
				FromClassKey:     "domain_key_a/subdomain_aa/class_b",
				FromMultiplicity: requirements.Multiplicity{LowerBound: 0, HigherBound: 1},
				ToClassKey:       "domain_key_b/subdomain_bb/class_b",
				ToMultiplicity:   requirements.Multiplicity{LowerBound: 2, HigherBound: 3},
			},
		},

		// Class States.
		States: map[string][]requirements.State{
			"domain_key_a/subdomain_aa/class_a": {
				{
					Key:  "domain_key_a/subdomain_aa/class_a/state_a",
					Name: "StateA",
				},
			},
		},
		Events: map[string][]requirements.Event{
			"domain_key_a/subdomain_aa/class_a": {
				{
					Key:  "domain_key_a/subdomain_aa/class_a/event_a",
					Name: "EventA",
				},
			},
		},
		Guards: map[string][]requirements.Guard{
			"domain_key_a/subdomain_aa/class_a": {
				{
					Key:  "domain_key_a/subdomain_aa/class_a/guard_a",
					Name: "GuardA",
				},
			},
		},
		Actions: map[string][]requirements.Action{
			"domain_key_a/subdomain_aa/class_a": {
				{
					Key:  "domain_key_a/subdomain_aa/class_a/action_a",
					Name: "ActionA",
				},
			},
		},
		Transitions: map[string][]requirements.Transition{
			"domain_key_a/subdomain_aa/class_a": {
				{
					Key:        "domain_key_a/subdomain_aa/class_a/transition_a",
					EventKey:   "domain_key_a/subdomain_aa/class_a/event_a",
					ActionKey:  "domain_key_a/subdomain_aa/class_a/action_a",
					ToStateKey: "domain_key_a/subdomain_aa/class_a/state_a",
				},
			},
		},
		StateActions: map[string][]requirements.StateAction{
			"domain_key_a/subdomain_aa/class_a/state_a": {
				{
					Key:       "domain_key_a/subdomain_aa/class_a/state_a/state_action_a",
					ActionKey: "domain_key_a/subdomain_aa/class_a/action_a",
					When:      "entry",
				},
			},
		},

		// Use Cases.
		UseCases: map[string][]requirements.UseCase{
			"domain_key_a/subdomain_aa": {
				{
					Key:      "domain_key_a/subdomain_aa/use_case_a",
					Name:     "UseCaseA",
					Level:    "sea",
					ReadOnly: false,
				},
			},
		},
		UseCaseActors: map[string]map[string]requirements.UseCaseActor{
			"domain_key_a/subdomain_aa/use_case_a": {
				"model_key/actor/actor_a": {},
			},
		},

		// Scenarios.
		Scenarios: map[string][]requirements.Scenario{
			"domain_key_a/subdomain_aa/use_case_a": {
				{
					Key:     "domain_key_a/subdomain_aa/use_case_a/scenario_a",
					Name:    "ScenarioA",
					Details: "Scenario details",
				},
			},
		},

		// Scenario Objects.
		ScenarioObjects: map[string][]requirements.ScenarioObject{
			"domain_key_a/subdomain_aa/use_case_a/scenario_a": {
				{
					Key:          "domain_key_a/subdomain_aa/use_case_a/scenario_a/object_a",
					ObjectNumber: 1,
					Name:         "model_key/object/object_a",
					NameStyle:    "name",
					ClassKey:     "domain_key_a/subdomain_aa/class_a",
					Multi:        false,
					UmlComment:   "Object comment",
				},
			},
		},
	}

	inOut := UnpackRequirements(original)
	back := PackRequirements(inOut)
	assert.Equal(t, original, back)
}
