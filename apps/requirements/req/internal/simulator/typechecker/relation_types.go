package typechecker

import (
	"github.com/glemzurg/go-tlaplus/internal/simulator/types"
)

// RelationTypeBuilder helps construct relation types for a class.
type RelationTypeBuilder struct {
	fieldTypes map[string]types.Type
}

// NewRelationTypeBuilder creates a new builder for relation types.
func NewRelationTypeBuilder() *RelationTypeBuilder {
	return &RelationTypeBuilder{
		fieldTypes: make(map[string]types.Type),
	}
}

// AddForwardRelation adds a forward relation type.
// Field name is the relation name (e.g., "Lines").
// targetType is the record type of the target class.
// The relation type will be Set[targetType].
func (b *RelationTypeBuilder) AddForwardRelation(fieldName string, targetType types.Type) {
	b.fieldTypes[fieldName] = types.Set{Element: targetType}
}

// AddReverseRelation adds a reverse relation type.
// Field name should include the underscore prefix (e.g., "_Lines").
// sourceType is the record type of the source class.
// The relation type will be Set[sourceType].
func (b *RelationTypeBuilder) AddReverseRelation(fieldName string, sourceType types.Type) {
	b.fieldTypes[fieldName] = types.Set{Element: sourceType}
}

// Build returns the constructed field types map.
func (b *RelationTypeBuilder) Build() map[string]types.Type {
	return b.fieldTypes
}

// BuildRelationTypesForClass constructs a map of relation field names to their types
// based on associations involving the given class.
//
// Parameters:
//   - classKey: The identity key of the class (as string)
//   - associations: List of associations involving this class
//   - getClassType: Function to get the Record type for a class key
//
// Returns a map suitable for TypeChecker.SetRelationTypes().
func BuildRelationTypesForClass(
	classKey string,
	associations []AssociationInfo,
	getClassType func(classKey string) types.Type,
) map[string]types.Type {
	builder := NewRelationTypeBuilder()

	for _, assoc := range associations {
		if assoc.FromClassKey == classKey {
			// Forward relation: this class is the "from" side
			// Access .Name to get Set[ToClass]
			targetType := getClassType(assoc.ToClassKey)
			if targetType != nil {
				builder.AddForwardRelation(assoc.Name, targetType)
			}
		}

		if assoc.ToClassKey == classKey {
			// Reverse relation: this class is the "to" side
			// Access ._Name to get Set[FromClass]
			sourceType := getClassType(assoc.FromClassKey)
			if sourceType != nil {
				builder.AddReverseRelation("_"+assoc.Name, sourceType)
			}
		}
	}

	return builder.Build()
}

// AssociationInfo contains the information needed to build relation types.
// This is a simplified view of model_class.Association to avoid import cycles.
type AssociationInfo struct {
	Key          string // Association's identity key as string
	Name         string // Display name (e.g., "Lines")
	FromClassKey string // From class identity key as string
	ToClassKey   string // To class identity key as string
}
