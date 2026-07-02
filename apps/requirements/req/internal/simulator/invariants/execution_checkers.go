package invariants

// StructuralInvariantCheckers groups implicit structural checks run after action execution.
type StructuralInvariantCheckers struct {
	Index                   *IndexUniquenessChecker
	Multiplicity            *MultiplicityChecker
	AssociationInstancePair *AssociationInstancePairChecker
	AssociationUniqueness   *AssociationUniquenessChecker
	AssociationInvariants   *AssociationInvariantChecker
}
