package engine

import (
	"fmt"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_class"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/evaluator"
)

// acToLegNameSuffix tags simulator-local AC→endpoint leg keys without embedding domain vocabulary.
const acToLegNameSuffix = "__ac_to_leg"

// AssociationClassInfo holds decomposed link metadata for one association-class role.
type AssociationClassInfo struct {
	AssociationClassKey identity.Key
	HostAssociation     model_class.Association
	FromClassKey        identity.Key
	ToClassKey          identity.Key
	FromLegAssocKey     identity.Key
	ToLegAssocKey       identity.Key
	FromLegName         string
	ToLegName           string
	FromLegFromMult     evaluator.Multiplicity
	FromLegToMult       evaluator.Multiplicity
	ToLegFromMult       evaluator.Multiplicity
	ToLegToMult         evaluator.Multiplicity
	InactiveStates      map[string]bool
}

func buildAssociationClassIndex(model *core.Model, scopedClasses map[identity.Key]*ClassInfo) (
	map[identity.Key]*AssociationClassInfo,
	map[identity.Key]bool,
	error,
) {
	index := make(map[identity.Key]*AssociationClassInfo)
	hostAssocKeys := make(map[identity.Key]bool)

	for _, assoc := range model.GetClassAssociations() {
		if assoc.AssociationClassKey == nil {
			continue
		}
		acKey := *assoc.AssociationClassKey
		if _, inScope := scopedClasses[acKey]; !inScope {
			continue
		}
		if _, fromIn := scopedClasses[assoc.FromClassKey]; !fromIn {
			continue
		}
		if _, toIn := scopedClasses[assoc.ToClassKey]; !toIn {
			continue
		}

		parentKey, err := associationParentKey(assoc.Key)
		if err != nil {
			return nil, nil, err
		}
		toLegKey, err := identity.NewClassAssociationKey(
			parentKey,
			acKey,
			assoc.ToClassKey,
			assoc.Name+acToLegNameSuffix,
		)
		if err != nil {
			return nil, nil, fmt.Errorf("association class %s to-leg key: %w", acKey.String(), err)
		}

		acClass := scopedClasses[acKey].Class
		info := &AssociationClassInfo{
			AssociationClassKey: acKey,
			HostAssociation:     assoc,
			FromClassKey:        assoc.FromClassKey,
			ToClassKey:          assoc.ToClassKey,
			FromLegAssocKey:     assoc.Key,
			ToLegAssocKey:       toLegKey,
			FromLegName:         assoc.Name,
			ToLegName:           assoc.Name + acToLegNameSuffix,
			FromLegFromMult:     oneMultiplicity(),
			FromLegToMult:       multiplicityFromModel(assoc.ToMultiplicity),
			ToLegFromMult:       multiplicityFromModel(assoc.FromMultiplicity),
			ToLegToMult:         oneMultiplicity(),
			InactiveStates:      inactiveAssociationClassStates(acClass),
		}
		index[acKey] = info
		hostAssocKeys[assoc.Key] = true
	}

	return index, hostAssocKeys, nil
}

func associationParentKey(assocKey identity.Key) (identity.Key, error) {
	if assocKey.ParentKey != "" {
		return identity.ParseKey(assocKey.ParentKey)
	}
	return identity.Key{}, nil
}

func oneMultiplicity() evaluator.Multiplicity {
	return evaluator.Multiplicity{LowerBound: 1, HigherBound: 1}
}

func multiplicityFromModel(m model_class.Multiplicity) evaluator.Multiplicity {
	return evaluator.Multiplicity{
		LowerBound:  m.LowerBound,
		HigherBound: m.HigherBound,
	}
}

func decomposedAssociationInfos(
	index map[identity.Key]*AssociationClassInfo,
) []AssociationInfo {
	var result []AssociationInfo
	for _, acInfo := range index {
		result = append(result, fromLegAssociationInfo(acInfo), toLegAssociationInfo(acInfo))
	}
	return result
}

func fromLegAssociationInfo(acInfo *AssociationClassInfo) AssociationInfo {
	host := acInfo.HostAssociation
	return AssociationInfo{
		Association: model_class.Association{
			Key:                 acInfo.FromLegAssocKey,
			Name:                acInfo.FromLegName,
			FromClassKey:        acInfo.FromClassKey,
			ToClassKey:          acInfo.AssociationClassKey,
			FromMultiplicity:    model_class.Multiplicity{LowerBound: 1, HigherBound: 1},
			ToMultiplicity:      host.ToMultiplicity,
			AssociationClassKey: nil,
		},
		FromClassKey:  acInfo.FromClassKey,
		ToClassKey:    acInfo.AssociationClassKey,
		MandatoryTo:   host.ToMultiplicity.LowerBound >= 1,
		MandatoryFrom: false,
		MinTo:         host.ToMultiplicity.LowerBound,
		MinFrom:       0,
	}
}

func toLegAssociationInfo(acInfo *AssociationClassInfo) AssociationInfo {
	host := acInfo.HostAssociation
	return AssociationInfo{
		Association: model_class.Association{
			Key:              acInfo.ToLegAssocKey,
			Name:             acInfo.ToLegName,
			FromClassKey:     acInfo.AssociationClassKey,
			ToClassKey:       acInfo.ToClassKey,
			FromMultiplicity: host.FromMultiplicity,
			ToMultiplicity:   model_class.Multiplicity{LowerBound: 1, HigherBound: 1},
		},
		FromClassKey:  acInfo.AssociationClassKey,
		ToClassKey:    acInfo.ToClassKey,
		MandatoryTo:   false,
		MandatoryFrom: host.FromMultiplicity.LowerBound >= 1,
		MinTo:         0,
		MinFrom:       host.FromMultiplicity.LowerBound,
	}
}

// inactiveAssociationClassStates marks AC states that cannot reach any creation target state.
func inactiveAssociationClassStates(class model_class.Class) map[string]bool {
	creationStates := creationTargetStateNames(class)
	inactive := make(map[string]bool)
	for _, state := range class.States {
		if !stateCanReachCreation(class, state.Name, creationStates) {
			inactive[state.Name] = true
		}
	}
	return inactive
}

func stateCanReachCreation(class model_class.Class, startName string, creationStates map[string]bool) bool {
	if creationStates[startName] {
		return true
	}
	reachable := statesReachableFrom(class, map[string]bool{startName: true})
	for name := range creationStates {
		if reachable[name] {
			return true
		}
	}
	return false
}

func creationTargetStateNames(class model_class.Class) map[string]bool {
	targets := make(map[string]bool)
	for _, t := range class.Transitions {
		if t.FromStateKey != nil || t.ToStateKey == nil {
			continue
		}
		if name := stateKeyToNameInClass(*t.ToStateKey, class); name != "" {
			targets[name] = true
		}
	}
	return targets
}

func statesReachableFrom(class model_class.Class, seeds map[string]bool) map[string]bool {
	reachable := make(map[string]bool)
	queue := make([]string, 0, len(seeds))
	for name := range seeds {
		reachable[name] = true
		queue = append(queue, name)
	}
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		currentKey := stateNameToKeyInClass(current, class)
		if currentKey == nil {
			continue
		}
		for _, t := range class.Transitions {
			if t.FromStateKey == nil || *t.FromStateKey != *currentKey || t.ToStateKey == nil {
				continue
			}
			nextName := stateKeyToNameInClass(*t.ToStateKey, class)
			if nextName == "" || reachable[nextName] {
				continue
			}
			reachable[nextName] = true
			queue = append(queue, nextName)
		}
	}
	return reachable
}

func stateKeyToNameInClass(stateKey identity.Key, class model_class.Class) string {
	if s, ok := class.States[stateKey]; ok {
		return s.Name
	}
	return ""
}

func stateNameToKeyInClass(stateName string, class model_class.Class) *identity.Key {
	for _, s := range class.States {
		if s.Name == stateName {
			key := s.Key
			return &key
		}
	}
	return nil
}

// IsActiveAssociationClassInstance reports whether an AC row participates in live configuration.
func IsActiveAssociationClassInstance(catalog *ClassCatalog, instanceClassKey identity.Key, stateName string) bool {
	if catalog == nil || !catalog.IsAssociationClass(instanceClassKey) {
		return true
	}
	acInfo := catalog.LookupAssociationClass(instanceClassKey)
	if acInfo == nil || len(acInfo.InactiveStates) == 0 {
		return true
	}
	return !acInfo.InactiveStates[stateName]
}
