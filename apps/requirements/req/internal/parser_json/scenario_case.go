package parser_json

// case_ represents a case in a switch node.
type case_ struct {
	Condition  string `json:"condition" yaml:"condition"`
	Statements []node `json:"statements" yaml:"statements"`
}
