package parser_json

import "github.com/glemzurg/glemzurg/apps/requirements/req/internal/requirements/model_domain"

// domainAssociationInOut is when a domain enforces requirements on another domain.
type domainAssociationInOut struct {
	Key               string `json:"key"`                 // The key of unique in the model.
	ProblemDomainKey  string `json:"problem_domain_key"`  // The domain that enforces requirements on the other domain.
	SolutionDomainKey string `json:"solution_domain_key"` // The domain that has requirements enforced upon it.
	UmlComment        string `json:"uml_comment"`
}

// ToRequirements converts the domainAssociationInOut to model_domain.DomainAssociation.
func (d domainAssociationInOut) ToRequirements() model_domain.DomainAssociation {
	return model_domain.DomainAssociation{
		Key:               d.Key,
		ProblemDomainKey:  d.ProblemDomainKey,
		SolutionDomainKey: d.SolutionDomainKey,
		UmlComment:        d.UmlComment,
	}
}

// FromRequirements creates a domainAssociationInOut from model_domain.DomainAssociation.
func FromRequirementsDomainAssociation(d model_domain.DomainAssociation) domainAssociationInOut {
	return domainAssociationInOut{
		Key:               d.Key,
		ProblemDomainKey:  d.ProblemDomainKey,
		SolutionDomainKey: d.SolutionDomainKey,
		UmlComment:        d.UmlComment,
	}
}
