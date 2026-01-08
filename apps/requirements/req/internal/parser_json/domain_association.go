package parser_json

import "github.com/glemzurg/glemzurg/apps/requirements/req/internal/requirements/model_domain"

// domainAssociationInOut is when a domain enforces requirements on another domain.
type domainAssociationInOut struct {
	Key               string `json:"key"`                 // The key of unique in the model.
	ProblemDomainKey  string `json:"problem_domain_key"`  // The domain that enforces requirements on the other domain.
	SolutionDomainKey string `json:"solution_domain_key"` // The domain that has requirements enforced upon it.
	UmlComment        string `json:"uml_comment"`
}

// ToRequirements converts the domainAssociationInOut to model_domain.Association.
func (d domainAssociationInOut) ToRequirements() model_domain.Association {
	return model_domain.Association{
		Key:               d.Key,
		ProblemDomainKey:  d.ProblemDomainKey,
		SolutionDomainKey: d.SolutionDomainKey,
		UmlComment:        d.UmlComment,
	}
}

// FromRequirements creates a domainAssociationInOut from model_domain.Association.
func FromRequirementsDomainAssociation(d model_domain.Association) domainAssociationInOut {
	return domainAssociationInOut{
		Key:               d.Key,
		ProblemDomainKey:  d.ProblemDomainKey,
		SolutionDomainKey: d.SolutionDomainKey,
		UmlComment:        d.UmlComment,
	}
}
