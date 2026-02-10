package model_state

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/pkg/errors"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
)

// Query is a business logic query of a class that does not change the state of a class.
// Guarantees describe filtering/selection criteria for returned data, NOT state changes.
type Query struct {
	Key           identity.Key
	Name          string
	Details       string
	Requires      []string // Human-readable preconditions for this query.
	Guarantees    []string // Human-readable filtering criteria for returned data.
	TlaRequires   []string // TLA+ expressions for preconditions.
	TlaGuarantees []string // TLA+ expressions for filtering criteria (NOT state changes).
	// Children
	Parameters []Parameter // Typed parameters for this query.
}

func NewQuery(key identity.Key, name, details string, requires, guarantees, tlaRequires, tlaGuarantees []string, parameters []Parameter) (query Query, err error) {

	query = Query{
		Key:           key,
		Name:          name,
		Details:       details,
		Requires:      requires,
		Guarantees:    guarantees,
		TlaRequires:   tlaRequires,
		TlaGuarantees: tlaGuarantees,
		Parameters:    parameters,
	}

	if err = query.Validate(); err != nil {
		return Query{}, err
	}

	return query, nil
}

// Validate validates the Query struct.
func (q *Query) Validate() error {
	return validation.ValidateStruct(q,
		validation.Field(&q.Key, validation.Required, validation.By(func(value interface{}) error {
			k := value.(identity.Key)
			if err := k.Validate(); err != nil {
				return err
			}
			if k.KeyType() != identity.KEY_TYPE_QUERY {
				return errors.Errorf("invalid key type '%s' for query", k.KeyType())
			}
			return nil
		})),
		validation.Field(&q.Name, validation.Required),
	)
}

// ValidateWithParent validates the Query, its key's parent relationship, and all children.
// The parent must be a Class.
func (q *Query) ValidateWithParent(parent *identity.Key) error {
	// Validate the object itself.
	if err := q.Validate(); err != nil {
		return err
	}
	// Validate the key has the correct parent.
	if err := q.Key.ValidateParent(parent); err != nil {
		return err
	}
	// Validate all children.
	for i := range q.Parameters {
		if err := q.Parameters[i].ValidateWithParent(); err != nil {
			return err
		}
	}
	return nil
}
