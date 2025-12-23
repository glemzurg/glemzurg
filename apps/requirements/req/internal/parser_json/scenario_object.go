package parser_json

// scenarioObjectInOut is an object that participates in a scenario.
type scenarioObjectInOut struct {
	Key          string `json:"key"`
	ObjectNumber uint   `json:"object_number"`        // Order in the scenario diagram.
	Name         string `json:"name"`                 // The name or id of the object.
	NameStyle    string `json:"name_style,omitempty"` // Used to format the name in the diagram.
	ClassKey     string `json:"class_key,omitempty"`  // The class key this object is an instance of.
	Multi        bool   `json:"multi"`
	UmlComment   string `json:"uml_comment,omitempty"`
}
