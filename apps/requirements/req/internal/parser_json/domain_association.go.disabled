package parser_json

import (
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_domain"
)

// domainAssociationInOut is when a domain enforces requirements on another domain.
type domainAssociationInOut struct {
	Key               string `json:"key"`                 // The key of unique in the model.
	ProblemDomainKey  string `json:"problem_domain_key"`  // The domain that enforces requirements on the other domain.
	SolutionDomainKey string `json:"solution_domain_key"` // The domain that has requirements enforced upon it.
	UmlComment        string `json:"uml_comment"`
}

// ToRequirements converts the domainAssociationInOut to model_domain.Association.
func (d domainAssociationInOut) ToRequirements() (model_domain.Association, error) {
	key, err := identity.ParseKey(d.Key)
	if err != nil {
		return model_domain.Association{}, err
	}

	problemDomainKey, err := identity.ParseKey(d.ProblemDomainKey)
	if err != nil {
		return model_domain.Association{}, err
	}

	solutionDomainKey, err := identity.ParseKey(d.SolutionDomainKey)
	if err != nil {
		return model_domain.Association{}, err
	}

	return model_domain.Association{
		Key:               key,
		ProblemDomainKey:  problemDomainKey,
		SolutionDomainKey: solutionDomainKey,
		UmlComment:        d.UmlComment,
	}, nil
}

// FromRequirements creates a domainAssociationInOut from model_domain.Association.
func FromRequirementsDomainAssociation(d model_domain.Association) domainAssociationInOut {
	return domainAssociationInOut{
		Key:               d.Key.String(),
		ProblemDomainKey:  d.ProblemDomainKey.String(),
		SolutionDomainKey: d.SolutionDomainKey.String(),
		UmlComment:        d.UmlComment,
	}
}
