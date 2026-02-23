package parser_ai

// inputGlobalFunction represents a global function/definition in JSON.
// Global functions are referenced from expressions throughout the model.
// Names must start with underscore (e.g., _Max, _SetOfValues).
type inputGlobalFunction struct {
	Name       string     `json:"name"`
	Parameters []string   `json:"parameters,omitempty"`
	Logic      inputLogic `json:"logic"`
}
