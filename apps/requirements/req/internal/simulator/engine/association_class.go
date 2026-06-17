package engine

import (
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_class"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
)

// AssociationClassInfo holds native host-association metadata for one association-class role.
type AssociationClassInfo struct {
	AssociationClassKey identity.Key
	HostAssociation     model_class.Association
	FromClassKey        identity.Key
	ToClassKey          identity.Key
	InactiveStates      map[string]bool
}

func buildAssociationClassIndex(model *core.Model, scopedClasses map[identity.Key]*ClassInfo) map[identity.Key]*AssociationClassInfo {
	index := make(map[identity.Key]*AssociationClassInfo)

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

		acClass := scopedClasses[acKey].Class
		index[acKey] = &AssociationClassInfo{
			AssociationClassKey: acKey,
			HostAssociation:     assoc,
			FromClassKey:        assoc.FromClassKey,
			ToClassKey:          assoc.ToClassKey,
			InactiveStates:      inactiveAssociationClassStates(acClass),
		}
	}

	return index
}

// CreationCascadeClassKey returns the class whose creation event satisfies a mandatory host association.
func CreationCascadeClassKey(ai AssociationInfo) identity.Key {
	if ai.Association.AssociationClassKey != nil {
		return *ai.Association.AssociationClassKey
	}
	return ai.ToClassKey
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
