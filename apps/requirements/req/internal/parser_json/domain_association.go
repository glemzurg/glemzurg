package parser_json

// domainAssociation is when a domain enforces requirements on another domain.
type domainAssociation struct {
	Key               string // The key of unique in the model.
	ProblemDomainKey  string // The domain that enforces requirements on the other domain.
	SolutionDomainKey string // The domain that has requirements enforced upon it.
	UmlComment        string
}
