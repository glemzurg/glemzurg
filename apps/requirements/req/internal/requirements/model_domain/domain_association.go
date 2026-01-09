package model_domain

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/pkg/errors"
)

// When a domain enforces requirements on another domain.
type Association struct {
	Key               string // The key of unique in the model.
	ProblemDomainKey  string // The domain that enforces requirements on the other domain.
	SolutionDomainKey string // The domain that has requirements enforced upon it.
	UmlComment        string
}

func NewAssociation(key, problemDomainKey, solutionDomainKey, umlComment string) (association Association, err error) {

	association = Association{
		Key:               key,
		ProblemDomainKey:  problemDomainKey,
		SolutionDomainKey: solutionDomainKey,
		UmlComment:        umlComment,
	}

	err = validation.ValidateStruct(&association,
		validation.Field(&association.Key, validation.Required),
		validation.Field(&association.ProblemDomainKey, validation.Required),
		validation.Field(&association.SolutionDomainKey, validation.Required),
	)
	if err != nil {
		return Association{}, errors.WithStack(err)
	}

	return association, nil
}
