package model_class

// AttributeTLAFieldName is the human-facing TLA+ identifier for an attribute on self:
// the attribute display name with spaces removed (case preserved).
func AttributeTLAFieldName(attributeName string) string {
	return AssociationTLAFieldName(attributeName)
}
