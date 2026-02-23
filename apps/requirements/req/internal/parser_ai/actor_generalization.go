package parser_ai

// inputActorGeneralization represents an actor generalization JSON file.
// Actor generalizations define super/sub-type hierarchies between actors.
type inputActorGeneralization struct {
	Name          string   `json:"name"`
	Details       string   `json:"details,omitempty"`
	SuperclassKey string   `json:"superclass_key"`
	SubclassKeys  []string `json:"subclass_keys"`
	IsComplete    bool     `json:"is_complete,omitempty"`
	IsStatic      bool     `json:"is_static,omitempty"`
	UMLComment    string   `json:"uml_comment,omitempty"`
}
