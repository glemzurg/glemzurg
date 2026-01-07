package parser_json

import "github.com/glemzurg/glemzurg/apps/requirements/req/internal/requirements/model_scenario"

// scenarioObjectInOut is an object that participates in a scenario.
type scenarioObjectInOut struct {
	Key          string `json:"key"`
	ObjectNumber uint   `json:"object_number"` // Order in the scenario diagram.
	Name         string `json:"name"`          // The name or id of the object.
	NameStyle    string `json:"name_style"`    // Used to format the name in the diagram.
	ClassKey     string `json:"class_key"`     // The class key this object is an instance of.
	Multi        bool   `json:"multi"`
	UmlComment   string `json:"uml_comment"`
}

// ToRequirements converts the scenarioObjectInOut to model_scenario.ScenarioObject.
func (s scenarioObjectInOut) ToRequirements() model_scenario.ScenarioObject {
	return model_scenario.ScenarioObject{
		Key:          s.Key,
		ObjectNumber: s.ObjectNumber,
		Name:         s.Name,
		NameStyle:    s.NameStyle,
		ClassKey:     s.ClassKey,
		Multi:        s.Multi,
		UmlComment:   s.UmlComment,
	}
}

// FromRequirements creates a scenarioObjectInOut from model_scenario.ScenarioObject.
func FromRequirementsScenarioObject(s model_scenario.ScenarioObject) scenarioObjectInOut {
	return scenarioObjectInOut{
		Key:          s.Key,
		ObjectNumber: s.ObjectNumber,
		Name:         s.Name,
		NameStyle:    s.NameStyle,
		ClassKey:     s.ClassKey,
		Multi:        s.Multi,
		UmlComment:   s.UmlComment,
	}
}
