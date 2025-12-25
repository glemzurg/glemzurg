package parser_json

import (
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/requirements"
	"github.com/stretchr/testify/assert"
)

func TestUnpackPackRequirementsRoundTrip(t *testing.T) {
	// Create a sample Requirements object. For simplicity, use a minimal one.
	original := requirements.Requirements{
		Model: requirements.Model{
			Key:     "model1",
			Name:    "Test Model",
			Details: "Details",
			Actors: []requirements.Actor{
				{Key: "actor1", Name: "User", Type: "person"},
			},
			Domains: []requirements.Domain{
				{
					Key:      "domain1",
					Name:     "Domain1",
					Realized: true,
					Subdomains: []requirements.Subdomain{
						{
							Key:  "subdomain1",
							Name: "Subdomain1",
							Classes: []requirements.Class{
								{
									Key:  "class1",
									Name: "Class1",
									Attributes: []requirements.Attribute{
										{Key: "attr1", Name: "Attr1"},
									},
									States: []requirements.State{
										{
											Key:  "state1",
											Name: "State1",
											Actions: []requirements.StateAction{
												{Key: "sa1", ActionKey: "action1", When: "enter"},
											},
										},
									},
									Events: []requirements.Event{
										{Key: "event1", Name: "Event1"},
									},
									Guards: []requirements.Guard{
										{Key: "guard1", Name: "Guard1"},
									},
									Actions: []requirements.Action{
										{Key: "action1", Name: "Action1"},
									},
									Transitions: []requirements.Transition{
										{Key: "trans1", FromStateKey: "state1", ToStateKey: "state2"},
									},
								},
							},
							UseCases: []requirements.UseCase{
								{
									Key:  "usecase1",
									Name: "UseCase1",
									Actors: map[string]requirements.UseCaseActor{
										"actor1": {UmlComment: "comment"},
									},
									Scenarios: []requirements.Scenario{
										{
											Key:  "scenario1",
											Name: "Scenario1",
											Objects: []requirements.ScenarioObject{
												{Key: "obj1", Name: "Obj1"},
											},
										},
									},
								},
							},
							Generalizations: []requirements.Generalization{
								{Key: "gen1", Name: "Gen1"},
							},
							Associations: []requirements.Association{
								{Key: "assoc1", Name: "Assoc1"},
							},
						},
					},
				},
			},
			DomainAssociations: []requirements.DomainAssociation{
				{Key: "da1", ProblemDomainKey: "domain1", SolutionDomainKey: "domain2"},
			},
			Associations: []requirements.Association{
				{Key: "assoc2", Name: "Assoc2"},
			},
		},
	}

	// Populate the flattened fields as they would be in a real Requirements object.
	original.Actors = original.Model.Actors
	original.Domains = original.Model.Domains
	original.DomainAssociations = original.Model.DomainAssociations
	original.Associations = append(original.Associations, original.Model.Associations...)
	original.Subdomains = make(map[string][]requirements.Subdomain)
	original.Subdomains["domain1"] = original.Model.Domains[0].Subdomains
	original.Classes = make(map[string][]requirements.Class)
	original.Classes["subdomain1"] = original.Model.Domains[0].Subdomains[0].Classes
	original.Attributes = make(map[string][]requirements.Attribute)
	original.Attributes["class1"] = original.Model.Domains[0].Subdomains[0].Classes[0].Attributes
	original.States = make(map[string][]requirements.State)
	original.States["class1"] = original.Model.Domains[0].Subdomains[0].Classes[0].States
	original.Events = make(map[string][]requirements.Event)
	original.Events["class1"] = original.Model.Domains[0].Subdomains[0].Classes[0].Events
	original.Guards = make(map[string][]requirements.Guard)
	original.Guards["class1"] = original.Model.Domains[0].Subdomains[0].Classes[0].Guards
	original.Actions = make(map[string][]requirements.Action)
	original.Actions["class1"] = original.Model.Domains[0].Subdomains[0].Classes[0].Actions
	original.Transitions = make(map[string][]requirements.Transition)
	original.Transitions["class1"] = original.Model.Domains[0].Subdomains[0].Classes[0].Transitions
	original.StateActions = make(map[string][]requirements.StateAction)
	original.StateActions["state1"] = original.Model.Domains[0].Subdomains[0].Classes[0].States[0].Actions
	original.UseCases = make(map[string][]requirements.UseCase)
	original.UseCases["subdomain1"] = original.Model.Domains[0].Subdomains[0].UseCases
	original.UseCaseActors = make(map[string]map[string]requirements.UseCaseActor)
	original.UseCaseActors["usecase1"] = original.Model.Domains[0].Subdomains[0].UseCases[0].Actors
	original.Scenarios = make(map[string][]requirements.Scenario)
	original.Scenarios["usecase1"] = original.Model.Domains[0].Subdomains[0].UseCases[0].Scenarios
	original.ScenarioObjects = make(map[string][]requirements.ScenarioObject)
	original.ScenarioObjects["scenario1"] = original.Model.Domains[0].Subdomains[0].UseCases[0].Scenarios[0].Objects
	original.Generalizations = original.Model.Domains[0].Subdomains[0].Generalizations
	original.Associations = append(original.Associations, original.Model.Domains[0].Subdomains[0].Associations...)

	// Unpack and pack back
	tree := UnpackRequirements(original)
	back := PackRequirements(tree)

	// Assert they are equal
	assert.Equal(t, original, back)
}