package parser_ai

// inputDomainAssociation represents a domain-level association JSON file.
// Domain associations describe constraint relationships between domains
// (problem domain enforces requirements on solution domain).
type inputDomainAssociation struct {
	ProblemDomainKey  string `json:"problem_domain_key"`
	SolutionDomainKey string `json:"solution_domain_key"`
	UmlComment        string `json:"uml_comment,omitempty"`
}
