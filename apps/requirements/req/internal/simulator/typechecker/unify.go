package typechecker

import (
	"fmt"

	"github.com/glemzurg/go-tlaplus/internal/simulator/types"
)

// UnificationError represents a type unification failure.
type UnificationError struct {
	Left    types.Type
	Right   types.Type
	Message string
}

func (e *UnificationError) Error() string {
	if e.Message != "" {
		return e.Message
	}
	return fmt.Sprintf("cannot unify %s with %s", e.Left, e.Right)
}

// Unify attempts to find a substitution that makes two types equal.
// Returns the most general unifier (mgu) or an error if unification fails.
//
// The algorithm implements Robinson's unification with occurs check.
func Unify(t1, t2 types.Type) (types.Substitution, error) {
	subst := make(types.Substitution)
	if err := unifyOne(t1, t2, subst); err != nil {
		return nil, err
	}
	return subst, nil
}

// unifyOne unifies two types, accumulating substitutions in subst.
func unifyOne(t1, t2 types.Type, subst types.Substitution) error {
	// Apply current substitutions
	t1 = subst.Apply(t1)
	t2 = subst.Apply(t2)

	// Same type: nothing to do
	if t1.Equals(t2) {
		return nil
	}

	// Any type unifies with anything
	if _, ok := t1.(types.Any); ok {
		return nil
	}
	if _, ok := t2.(types.Any); ok {
		return nil
	}

	// Type variable on left
	if tv1, ok := t1.(types.TypeVar); ok {
		return bind(tv1, t2, subst)
	}

	// Type variable on right
	if tv2, ok := t2.(types.TypeVar); ok {
		return bind(tv2, t1, subst)
	}

	// Structural unification for compound types
	switch left := t1.(type) {
	case types.Set:
		if right, ok := t2.(types.Set); ok {
			return unifyOne(left.Element, right.Element, subst)
		}

	case types.Tuple:
		if right, ok := t2.(types.Tuple); ok {
			return unifyOne(left.Element, right.Element, subst)
		}

	case types.Bag:
		if right, ok := t2.(types.Bag); ok {
			return unifyOne(left.Element, right.Element, subst)
		}

	case types.Function:
		if right, ok := t2.(types.Function); ok {
			if len(left.Params) != len(right.Params) {
				return &UnificationError{
					Left:    t1,
					Right:   t2,
					Message: fmt.Sprintf("function arity mismatch: %d vs %d parameters", len(left.Params), len(right.Params)),
				}
			}
			// Unify each parameter
			for i := range left.Params {
				if err := unifyOne(left.Params[i], right.Params[i], subst); err != nil {
					return err
				}
			}
			// Unify return types
			return unifyOne(left.Return, right.Return, subst)
		}

	case types.Record:
		if right, ok := t2.(types.Record); ok {
			if len(left.Fields) != len(right.Fields) {
				return &UnificationError{
					Left:    t1,
					Right:   t2,
					Message: fmt.Sprintf("record field count mismatch: %d vs %d fields", len(left.Fields), len(right.Fields)),
				}
			}
			for name, leftField := range left.Fields {
				rightField, exists := right.Fields[name]
				if !exists {
					return &UnificationError{
						Left:    t1,
						Right:   t2,
						Message: fmt.Sprintf("record missing field: %s", name),
					}
				}
				if err := unifyOne(leftField, rightField, subst); err != nil {
					return err
				}
			}
			return nil
		}
	}

	// Types don't match
	return &UnificationError{Left: t1, Right: t2}
}

// bind adds a binding for a type variable, with occurs check.
func bind(tv types.TypeVar, t types.Type, subst types.Substitution) error {
	// Check if already bound to this type
	if t.Equals(tv) {
		return nil
	}

	// Occurs check: prevent infinite types like a = List[a]
	if occursIn(tv.ID, t) {
		return &UnificationError{
			Left:    tv,
			Right:   t,
			Message: fmt.Sprintf("infinite type: %s occurs in %s", tv, t),
		}
	}

	// Add to substitution
	subst[tv.ID] = t
	return nil
}

// occursIn checks if a type variable ID occurs in a type.
func occursIn(id int, t types.Type) bool {
	freeVars := t.FreeTypeVars()
	_, exists := freeVars[id]
	return exists
}

// UnifyAll unifies a list of type pairs.
func UnifyAll(pairs [][2]types.Type) (types.Substitution, error) {
	subst := make(types.Substitution)
	for _, pair := range pairs {
		if err := unifyOne(pair[0], pair[1], subst); err != nil {
			return nil, err
		}
	}
	return subst, nil
}

// UnifyWithSubst unifies two types given an existing substitution.
// Returns the extended substitution.
func UnifyWithSubst(t1, t2 types.Type, subst types.Substitution) (types.Substitution, error) {
	// Create a copy of the substitution
	newSubst := make(types.Substitution, len(subst))
	for k, v := range subst {
		newSubst[k] = v
	}

	if err := unifyOne(t1, t2, newSubst); err != nil {
		return nil, err
	}
	return newSubst, nil
}
