package parser_json

import "github.com/glemzurg/glemzurg/apps/requirements/req/internal/requirements/model_class"

// generalizationInOut is how two or more things in the system build on each other (like a super type and sub type).
type generalizationInOut struct {
	Key        string `json:"key"`
	Name       string `json:"name"`
	Details    string `json:"details"`     // Markdown.
	IsComplete bool   `json:"is_complete"` // Are the specializations complete, or can an instantiation of this generalization exist without a specialization.
	IsStatic   bool   `json:"is_static"`   // Are the specializations static and unchanging or can they change during runtime.
	UmlComment string `json:"uml_comment"`
}

// ToRequirements converts the generalizationInOut to model_class.Generalization.
func (g generalizationInOut) ToRequirements() model_class.Generalization {
	return model_class.Generalization{
		Key:        g.Key,
		Name:       g.Name,
		Details:    g.Details,
		IsComplete: g.IsComplete,
		IsStatic:   g.IsStatic,
		UmlComment: g.UmlComment,
	}
}

// FromRequirements creates a generalizationInOut from model_class.Generalization.
func FromRequirementsGeneralization(g model_class.Generalization) generalizationInOut {
	return generalizationInOut{
		Key:        g.Key,
		Name:       g.Name,
		Details:    g.Details,
		IsComplete: g.IsComplete,
		IsStatic:   g.IsStatic,
		UmlComment: g.UmlComment,
	}
}
