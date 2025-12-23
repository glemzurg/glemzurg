package parser_json

// scenarioObject is an object that participates in a scenario.
type scenarioObject struct {
	Key          string
	ObjectNumber uint   // Order in the scenario diagram.
	Name         string // The name or id of the object.
	NameStyle    string // Used to format the name in the diagram.
	ClassKey     string // The class key this object is an instance of.
	Multi        bool
	UmlComment   string
}
