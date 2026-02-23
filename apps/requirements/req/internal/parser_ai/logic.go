package parser_ai

// inputLogic represents a formal logic specification in JSON.
type inputLogic struct {
	Description   string `json:"description"`
	Notation      string `json:"notation,omitempty"`
	Specification string `json:"specification,omitempty"`
}

// inputParameter represents a typed parameter in JSON.
type inputParameter struct {
	Name          string `json:"name"`
	DataTypeRules string `json:"data_type_rules,omitempty"`
}
