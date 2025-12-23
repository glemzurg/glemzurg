package parser_json

// eventParameterInOut is a parameter for events.
type eventParameterInOut struct {
	Name   string `json:"name"`
	Source string `json:"source,omitempty"` // Where the values for this parameter are coming from.
}
