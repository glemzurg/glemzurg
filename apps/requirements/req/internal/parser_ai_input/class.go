package parser_ai_input

// inputAttribute represents an attribute within a class.
type inputAttribute struct {
	Name             string `json:"name"`
	DataTypeRules    string `json:"data_type_rules,omitempty"`
	Details          string `json:"details,omitempty"`
	DerivationPolicy string `json:"derivation_policy,omitempty"`
	Nullable         bool   `json:"nullable,omitempty"`
	UmlComment       string `json:"uml_comment,omitempty"`
}

// inputClass represents a class.json file.
type inputClass struct {
	Name       string                    `json:"name"`
	Details    string                    `json:"details,omitempty"`
	ActorKey   string                    `json:"actor_key,omitempty"`
	UmlComment string                    `json:"uml_comment,omitempty"`
	Attributes map[string]inputAttribute `json:"attributes,omitempty"`
	Indexes    [][]string                `json:"indexes,omitempty"`
}
