package object

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
