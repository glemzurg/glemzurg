package object

import (
	"math/big"
	"strings"
)

// Null returns the simulator absent-value sentinel (empty set).
func Null() Object {
	return NewSet()
}

// IsNull reports whether value represents NULL (unset or empty set).
func IsNull(value Object) bool {
	if value == nil {
		return true
	}
	if set, ok := value.(*Set); ok {
		return set.Size() == 0
	}
	return false
}

// NormalizeSimulatorValue maps empty strings to Null so absent STRING values use NULL.
func NormalizeSimulatorValue(value Object) Object {
	if value == nil {
		return nil
	}
	if str, ok := value.(*String); ok && str.Value() == "" {
		return Null()
	}
	return value
}

// CoerceToNumber returns value as a Number when it already is one, or when it is a
// string that parses as an integer/rational (e.g. "42", "-3", "1/2"). Used so
// INT/span attributes that slipped into storage as strings still work in arithmetic.
func CoerceToNumber(value Object) (*Number, bool) {
	if value == nil {
		return nil, false
	}
	if n, ok := value.(*Number); ok {
		return n, true
	}
	str, ok := value.(*String)
	if !ok {
		return nil, false
	}
	text := strings.TrimSpace(str.Value())
	if text == "" {
		return nil, false
	}
	if rat, ok := new(big.Rat).SetString(text); ok {
		return &Number{rat: rat}, true
	}
	return nil, false
}
