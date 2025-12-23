package parser_json

// domainAssociationInOut is when a domain enforces requirements on another domain.
type domainAssociationInOut struct {
	Key               string `json:"key"`                 // The key of unique in the model.
	ProblemDomainKey  string `json:"problem_domain_key"`  // The domain that enforces requirements on the other domain.
	SolutionDomainKey string `json:"solution_domain_key"` // The domain that has requirements enforced upon it.
	UmlComment        string `json:"uml_comment"`
}
