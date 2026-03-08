package model_state

import (
	"fmt"

	"github.com/pkg/errors"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/coreerr"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_logic"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
)

// Query is a business logic query of a class that does not change the state of a class.
// Guarantees describe filtering/selection criteria for returned data, NOT state changes.
type Query struct {
	Key     identity.Key
	Name    string
	Details string
	// Children
	Parameters []Parameter         // Typed parameters for this query.
	Requires   []model_logic.Logic // Preconditions for this query.
	Guarantees []model_logic.Logic // Filtering criteria for returned data (NOT state changes).
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
	// Validate the key.
	if err := q.Key.Validate(); err != nil {
		return coreerr.New(coreerr.QueryKeyInvalid, fmt.Sprintf("Key: %s", err.Error()), "Key")
	}
	if q.Key.KeyType != identity.KEY_TYPE_QUERY {
		return coreerr.NewWithValues(coreerr.QueryKeyTypeInvalid, fmt.Sprintf("Key: invalid key type '%s' for query", q.Key.KeyType), "Key", q.Key.KeyType, identity.KEY_TYPE_QUERY)
	}

	if q.Name == "" {
		return coreerr.New(coreerr.QueryNameRequired, "Name is required", "Name")
	}

	reqLetTargets := make(map[string]bool)
	for i, req := range q.Requires {
		if err := req.Validate(); err != nil {
			return errors.Wrapf(err, "requires %d", i)
		}
		if req.Type != model_logic.LogicTypeAssessment && req.Type != model_logic.LogicTypeLet {
			return coreerr.NewWithValues(coreerr.QueryRequiresTypeInvalid, fmt.Sprintf("requires %d: logic kind must be '%s' or '%s', got '%s'", i, model_logic.LogicTypeAssessment, model_logic.LogicTypeLet, req.Type), "Requires", req.Type, fmt.Sprintf("one of: %s, %s", model_logic.LogicTypeAssessment, model_logic.LogicTypeLet))
		}
		if req.Type == model_logic.LogicTypeLet {
			if reqLetTargets[req.Target] {
				return coreerr.NewWithValues(coreerr.QueryRequiresDuplicateLet, fmt.Sprintf("requires %d: duplicate let target %q", i, req.Target), "Requires", req.Target, "")
			}
			reqLetTargets[req.Target] = true
		}
	}
	guarTargets := make(map[string]bool)
	for i, guar := range q.Guarantees {
		if err := guar.Validate(); err != nil {
			return errors.Wrapf(err, "guarantee %d", i)
		}
		if guar.Type != model_logic.LogicTypeQuery && guar.Type != model_logic.LogicTypeLet {
			return coreerr.NewWithValues(coreerr.QueryGuaranteeTypeInvalid, fmt.Sprintf("guarantee %d: logic kind must be '%s' or '%s', got '%s'", i, model_logic.LogicTypeQuery, model_logic.LogicTypeLet, guar.Type), "Guarantees", guar.Type, fmt.Sprintf("one of: %s, %s", model_logic.LogicTypeQuery, model_logic.LogicTypeLet))
		}
		// Each guarantee and let must set a unique target.
		if guarTargets[guar.Target] {
			if guar.Type == model_logic.LogicTypeLet {
				return coreerr.NewWithValues(coreerr.QueryGuaranteeDuplicateLet, fmt.Sprintf("guarantee %d: duplicate let target %q", i, guar.Target), "Guarantees", guar.Target, "")
			}
			return coreerr.NewWithValues(coreerr.QueryGuaranteeDuplicateTarget, fmt.Sprintf("guarantee %d: duplicate target %q — each output identifier can only appear once per query", i, guar.Target), "Guarantees", guar.Target, "")
		}
		guarTargets[guar.Target] = true
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
	// Validate logic children with query as parent.
	for i := range q.Requires {
		if err := q.Requires[i].ValidateWithParent(&q.Key); err != nil {
			return errors.Wrapf(err, "requires %d", i)
		}
	}
	for i := range q.Guarantees {
		if err := q.Guarantees[i].ValidateWithParent(&q.Key); err != nil {
			return errors.Wrapf(err, "guarantee %d", i)
		}
	}
	// Validate all children.
	for i := range q.Parameters {
		if err := q.Parameters[i].ValidateWithParent(); err != nil {
			return err
		}
	}
	return nil
}
