package parser_ai

// inputUseCaseGeneralization represents a use case generalization JSON file.
// Use case generalizations define super/sub-type hierarchies between use cases.
type inputUseCaseGeneralization struct {
	Name          string   `json:"name"`
	Details       string   `json:"details,omitempty"`
	SuperclassKey string   `json:"superclass_key"`
	SubclassKeys  []string `json:"subclass_keys"`
	IsComplete    bool     `json:"is_complete,omitempty"`
	IsStatic      bool     `json:"is_static,omitempty"`
	UMLComment    string   `json:"uml_comment,omitempty"`
}
