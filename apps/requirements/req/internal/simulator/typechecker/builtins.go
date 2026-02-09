package typechecker

import (
	"github.com/glemzurg/go-tlaplus/internal/simulator/types"
)

// addBuiltins registers all builtin function type signatures.
// These are polymorphic types following the pattern:
//
//	_Seq!Head : ∀a. Tuple[a] → a
//	_Bags!SetToBag : ∀a. Set[a] → Bag[a]
func (tc *TypeChecker) addBuiltins() {
	// === Sequence Operations ===

	// _Seq!Head : ∀a. Tuple[a] → a
	tc.addPolymorphicBuiltin("_Seq!Head", 1, func(tvs []types.TypeVar) types.Function {
		a := tvs[0]
		return types.Function{
			Params: []types.Type{types.Tuple{Element: a}},
			Return: a,
		}
	})

	// _Seq!Tail : ∀a. Tuple[a] → Tuple[a]
	tc.addPolymorphicBuiltin("_Seq!Tail", 1, func(tvs []types.TypeVar) types.Function {
		a := tvs[0]
		return types.Function{
			Params: []types.Type{types.Tuple{Element: a}},
			Return: types.Tuple{Element: a},
		}
	})

	// _Seq!Append : ∀a. (Tuple[a], a) → Tuple[a]
	tc.addPolymorphicBuiltin("_Seq!Append", 1, func(tvs []types.TypeVar) types.Function {
		a := tvs[0]
		return types.Function{
			Params: []types.Type{types.Tuple{Element: a}, a},
			Return: types.Tuple{Element: a},
		}
	})

	// _Seq!Len : ∀a. Tuple[a] → Number
	tc.addPolymorphicBuiltin("_Seq!Len", 1, func(tvs []types.TypeVar) types.Function {
		a := tvs[0]
		return types.Function{
			Params: []types.Type{types.Tuple{Element: a}},
			Return: types.Number{},
		}
	})

	// _Seq!SubSeq : ∀a. (Tuple[a], Number, Number) → Tuple[a]
	tc.addPolymorphicBuiltin("_Seq!SubSeq", 1, func(tvs []types.TypeVar) types.Function {
		a := tvs[0]
		return types.Function{
			Params: []types.Type{types.Tuple{Element: a}, types.Number{}, types.Number{}},
			Return: types.Tuple{Element: a},
		}
	})

	// _Seq!SelectSeq : ∀a. (Tuple[a], (a → Boolean)) → Tuple[a]
	// Note: Simplified - higher-order functions need special handling
	tc.addPolymorphicBuiltin("_Seq!SelectSeq", 1, func(tvs []types.TypeVar) types.Function {
		a := tvs[0]
		return types.Function{
			Params: []types.Type{
				types.Tuple{Element: a},
				types.Function{Params: []types.Type{a}, Return: types.Boolean{}},
			},
			Return: types.Tuple{Element: a},
		}
	})

	// _Seq!Concat : ∀a. (Tuple[a], Tuple[a]) → Tuple[a]
	tc.addPolymorphicBuiltin("_Seq!Concat", 1, func(tvs []types.TypeVar) types.Function {
		a := tvs[0]
		return types.Function{
			Params: []types.Type{types.Tuple{Element: a}, types.Tuple{Element: a}},
			Return: types.Tuple{Element: a},
		}
	})

	// === Bag Operations ===

	// _Bags!SetToBag : ∀a. Set[a] → Bag[a]
	tc.addPolymorphicBuiltin("_Bags!SetToBag", 1, func(tvs []types.TypeVar) types.Function {
		a := tvs[0]
		return types.Function{
			Params: []types.Type{types.Set{Element: a}},
			Return: types.Bag{Element: a},
		}
	})

	// _Bags!BagToSet : ∀a. Bag[a] → Set[a]
	tc.addPolymorphicBuiltin("_Bags!BagToSet", 1, func(tvs []types.TypeVar) types.Function {
		a := tvs[0]
		return types.Function{
			Params: []types.Type{types.Bag{Element: a}},
			Return: types.Set{Element: a},
		}
	})

	// _Bags!BagIn : ∀a. (a, Bag[a]) → Boolean
	tc.addPolymorphicBuiltin("_Bags!BagIn", 1, func(tvs []types.TypeVar) types.Function {
		a := tvs[0]
		return types.Function{
			Params: []types.Type{a, types.Bag{Element: a}},
			Return: types.Boolean{},
		}
	})

	// _Bags!EmptyBag : ∀a. () → Bag[a]
	tc.addPolymorphicBuiltin("_Bags!EmptyBag", 1, func(tvs []types.TypeVar) types.Function {
		a := tvs[0]
		return types.Function{
			Params: []types.Type{},
			Return: types.Bag{Element: a},
		}
	})

	// _Bags!CopiesIn : ∀a. (a, Bag[a]) → Number
	tc.addPolymorphicBuiltin("_Bags!CopiesIn", 1, func(tvs []types.TypeVar) types.Function {
		a := tvs[0]
		return types.Function{
			Params: []types.Type{a, types.Bag{Element: a}},
			Return: types.Number{},
		}
	})

	// _Bags!BagCardinality : ∀a. Bag[a] → Number
	tc.addPolymorphicBuiltin("_Bags!BagCardinality", 1, func(tvs []types.TypeVar) types.Function {
		a := tvs[0]
		return types.Function{
			Params: []types.Type{types.Bag{Element: a}},
			Return: types.Number{},
		}
	})

	// _Bags!BagUnion : ∀a. Set[Bag[a]] → Bag[a]
	tc.addPolymorphicBuiltin("_Bags!BagUnion", 1, func(tvs []types.TypeVar) types.Function {
		a := tvs[0]
		return types.Function{
			Params: []types.Type{types.Set{Element: types.Bag{Element: a}}},
			Return: types.Bag{Element: a},
		}
	})

	// === FiniteSet Operations ===

	// _FiniteSet!Cardinality : ∀a. Set[a] → Number
	tc.addPolymorphicBuiltin("_FiniteSet!Cardinality", 1, func(tvs []types.TypeVar) types.Function {
		a := tvs[0]
		return types.Function{
			Params: []types.Type{types.Set{Element: a}},
			Return: types.Number{},
		}
	})

	// _FiniteSet!IsFiniteSet : ∀a. Set[a] → Boolean
	tc.addPolymorphicBuiltin("_FiniteSet!IsFiniteSet", 1, func(tvs []types.TypeVar) types.Function {
		a := tvs[0]
		return types.Function{
			Params: []types.Type{types.Set{Element: a}},
			Return: types.Boolean{},
		}
	})

	// _FiniteSet!CHOOSE : ∀a. Set[a] → a
	tc.addPolymorphicBuiltin("_FiniteSet!CHOOSE", 1, func(tvs []types.TypeVar) types.Function {
		a := tvs[0]
		return types.Function{
			Params: []types.Type{types.Set{Element: a}},
			Return: a,
		}
	})

	// === TLC Operations ===

	// _TLC!Print : ∀a. (String, a) → a
	tc.addPolymorphicBuiltin("_TLC!Print", 1, func(tvs []types.TypeVar) types.Function {
		a := tvs[0]
		return types.Function{
			Params: []types.Type{types.String{}, a},
			Return: a,
		}
	})

	// _TLC!PrintT : ∀a. a → Boolean
	tc.addPolymorphicBuiltin("_TLC!PrintT", 1, func(tvs []types.TypeVar) types.Function {
		a := tvs[0]
		return types.Function{
			Params: []types.Type{a},
			Return: types.Boolean{},
		}
	})

	// _TLC!Assert : ∀a. (Boolean, a) → Boolean
	tc.addPolymorphicBuiltin("_TLC!Assert", 1, func(tvs []types.TypeVar) types.Function {
		a := tvs[0]
		return types.Function{
			Params: []types.Type{types.Boolean{}, a},
			Return: types.Boolean{},
		}
	})

	// === Standard Math ===

	// Min : (Number, Number) → Number
	tc.env.Bind("Min", types.Monotype(types.Function{
		Params: []types.Type{types.Number{}, types.Number{}},
		Return: types.Number{},
	}))

	// Max : (Number, Number) → Number
	tc.env.Bind("Max", types.Monotype(types.Function{
		Params: []types.Type{types.Number{}, types.Number{}},
		Return: types.Number{},
	}))

	// Abs : Number → Number
	tc.env.Bind("Abs", types.Monotype(types.Function{
		Params: []types.Type{types.Number{}},
		Return: types.Number{},
	}))

	// === String Operations ===

	// _Strings!Len : String → Number
	tc.env.Bind("_Strings!Len", types.Monotype(types.Function{
		Params: []types.Type{types.String{}},
		Return: types.Number{},
	}))

	// _Strings!SubString : (String, Number, Number) → String
	tc.env.Bind("_Strings!SubString", types.Monotype(types.Function{
		Params: []types.Type{types.String{}, types.Number{}, types.Number{}},
		Return: types.String{},
	}))

	// === DOMAIN ===

	// DOMAIN for records returns a set of strings (field names)
	// DOMAIN for tuples/sequences returns a set of numbers (indices)
	// DOMAIN for functions returns the domain set
	// We use Any type since DOMAIN is highly polymorphic
	tc.addPolymorphicBuiltin("DOMAIN", 1, func(tvs []types.TypeVar) types.Function {
		a := tvs[0]
		return types.Function{
			Params: []types.Type{a},
			Return: types.Set{Element: types.Any{}},
		}
	})
}

// addPolymorphicBuiltin adds a polymorphic builtin with the given number of type variables.
func (tc *TypeChecker) addPolymorphicBuiltin(name string, numTypeVars int, makeType func([]types.TypeVar) types.Function) {
	// Create type variables
	tvs := make([]types.TypeVar, numTypeVars)
	varIDs := make([]int, numTypeVars)
	for i := 0; i < numTypeVars; i++ {
		tv := types.NewTypeVar("")
		tvs[i] = tv
		varIDs[i] = tv.ID
	}

	// Create the function type using the type variables
	fnType := makeType(tvs)

	// Create a scheme quantifying over the type variables
	scheme := types.Scheme{
		TypeVars: varIDs,
		Type:     fnType,
	}

	tc.env.Bind(name, scheme)
}
