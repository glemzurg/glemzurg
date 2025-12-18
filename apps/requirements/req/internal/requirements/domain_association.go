package requirements

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/pkg/errors"
)

// When a domain enforces requirements on another domain.
type DomainAssociation struct {
	Key               string // The key of unique in the model.
	ProblemDomainKey  string // The domain that enforces requirements on the other domain.
	SolutionDomainKey string // The domain that has requirements enforced upon it.
	UmlComment        string
}

func NewDomainAssociation(key, problemDomainKey, solutionDomainKey, umlComment string) (domainAssociation DomainAssociation, err error) {

	domainAssociation = DomainAssociation{
		Key:               key,
		ProblemDomainKey:  problemDomainKey,
		SolutionDomainKey: solutionDomainKey,
		UmlComment:        umlComment,
	}

	err = validation.ValidateStruct(&domainAssociation,
		validation.Field(&domainAssociation.Key, validation.Required),
		validation.Field(&domainAssociation.ProblemDomainKey, validation.Required),
		validation.Field(&domainAssociation.SolutionDomainKey, validation.Required),
	)
	if err != nil {
		return DomainAssociation{}, errors.WithStack(err)
	}

	return domainAssociation, nil
}
