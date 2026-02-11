package model_state

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/pkg/errors"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_logic"
)

// Query is a business logic query of a class that does not change the state of a class.
// Guarantees describe filtering/selection criteria for returned data, NOT state changes.
type Query struct {
	Key        identity.Key
	Name       string
	Details    string
	Requires   []model_logic.Logic // Preconditions for this query.
	Guarantees []model_logic.Logic // Filtering criteria for returned data (NOT state changes).
	// Children
	Parameters []Parameter // Typed parameters for this query.
}

func NewQuery(key identity.Key, name, details string, requires, guarantees []model_logic.Logic, parameters []Parameter) (query Query, err error) {

	query = Query{
		Key:        key,
		Name:       name,
		Details:    details,
		Requires:   requires,
		Guarantees: guarantees,
		Parameters: parameters,
	}

	if err = query.Validate(); err != nil {
		return Query{}, err
	}

	return query, nil
}

// Validate validates the Query struct.
func (q *Query) Validate() error {
	if err := validation.ValidateStruct(q,
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
	); err != nil {
		return err
	}

	for i, req := range q.Requires {
		if err := req.Validate(); err != nil {
			return errors.Wrapf(err, "requires %d", i)
		}
	}
	for i, guar := range q.Guarantees {
		if err := guar.Validate(); err != nil {
			return errors.Wrapf(err, "guarantee %d", i)
		}
	}

	return nil
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
