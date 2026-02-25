package parser_json

import (
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_scenario"
)

// objectInOut is an object that participates in a scenario.
type objectInOut struct {
	Key          string `json:"key"`
	ObjectNumber uint   `json:"object_number"` // Order in the scenario diagram.
	Name         string `json:"name"`          // The name or id of the object.
	NameStyle    string `json:"name_style"`    // Used to format the name in the diagram.
	ClassKey     string `json:"class_key"`     // The class key this object is an instance of.
	Multi        bool   `json:"multi"`
	UmlComment   string `json:"uml_comment"`
}

// ToRequirements converts the objectInOut to model_scenario.Object.
func (s objectInOut) ToRequirements() (model_scenario.Object, error) {
	key, err := identity.ParseKey(s.Key)
	if err != nil {
		return model_scenario.Object{}, err
	}

	classKey, err := identity.ParseKey(s.ClassKey)
	if err != nil {
		return model_scenario.Object{}, err
	}

	return model_scenario.Object{
		Key:          key,
		ObjectNumber: s.ObjectNumber,
		Name:         s.Name,
		NameStyle:    s.NameStyle,
		ClassKey:     classKey,
		Multi:        s.Multi,
		UmlComment:   s.UmlComment,
	}, nil
}

// FromRequirements creates a objectInOut from model_scenario.Object.
func FromRequirementsObject(s model_scenario.Object) objectInOut {
	return objectInOut{
		Key:          s.Key.String(),
		ObjectNumber: s.ObjectNumber,
		Name:         s.Name,
		NameStyle:    s.NameStyle,
		ClassKey:     s.ClassKey.String(),
		Multi:        s.Multi,
		UmlComment:   s.UmlComment,
	}
}
