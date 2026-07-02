package evaluator

import "strings"

// NormalizeAssociationFieldName converts a model association or class display name
// into the TLA+ identifier the simulator uses (spaces removed, case preserved).
func NormalizeAssociationFieldName(name string) string {
	return strings.ReplaceAll(strings.TrimSpace(name), " ", "")
}
