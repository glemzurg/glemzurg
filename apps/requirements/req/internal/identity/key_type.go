package identity

import (
	"strings"

	"github.com/pkg/errors"
)

const (

	// Models do not have a key type.
	// It is a string that is unique in the system.

	// Keys without parents (parent is the model itself).
	KEY_TYPE_ACTOR  = "actor"
	KEY_TYPE_DOMAIN = "domain"

	// Keys with domain parents.
	KEY_TYPE_SUBDOMAIN          = "subdomain"
	KEY_TYPE_DOMAIN_ASSOCIATION = "dassociation"

	// Keys with subdomain parents.
	KEY_TYPE_USE_CASE       = "usecase"
	KEY_TYPE_CLASS          = "class"
	KEY_TYPE_GENERALIZATION = "generalization"

	// Keys with model, domain, subdomain parents.
	KEY_TYPE_CLASS_ASSOCIATION = "cassociation"

	// Keys with class parents.
	KEY_TYPE_ATTRIBUTE  = "attribute"
	KEY_TYPE_STATE      = "state"
	KEY_TYPE_EVENT      = "event"
	KEY_TYPE_GUARD      = "guard"
	KEY_TYPE_ACTION     = "action"
	KEY_TYPE_TRANSITION = "transition"

	// remaining
	KEY_TYPE_SCENARIO        = "scenario"
	KEY_TYPE_SCENARIO_OBJECT = "sobject"
	KEY_TYPE_STATE_ACTION    = "saction"
)

func NewActorKey(subKey string) (key Key, err error) {
	return newRootKey(KEY_TYPE_ACTOR, subKey)
}

func NewDomainKey(subKey string) (key Key, err error) {
	return newRootKey(KEY_TYPE_DOMAIN, subKey)
}

func NewDomainAssociationKey(problemDomainKey, solutionDomainKey Key) (key Key, err error) {
	// Both must be domains.
	if problemDomainKey.KeyType() != KEY_TYPE_DOMAIN {
		return Key{}, errors.Errorf("problem domain key cannot be of type '%s' for 'dassociation' key", problemDomainKey.KeyType())
	}
	if solutionDomainKey.KeyType() != KEY_TYPE_DOMAIN {
		return Key{}, errors.Errorf("solution domain key cannot be of type '%s' for 'dassociation' key", solutionDomainKey.KeyType())
	}
	// Parent is the problem domain, subKey is the solution domain's subKey.
	return newKey(problemDomainKey.String(), KEY_TYPE_DOMAIN_ASSOCIATION, solutionDomainKey.SubKey())
}

func NewSubdomainKey(domainKey Key, subKey string) (key Key, err error) {
	// The parent must be a domain.
	if domainKey.KeyType() != KEY_TYPE_DOMAIN {
		return Key{}, errors.Errorf("parent key cannot be of type '%s' for 'subdomain' key", domainKey.KeyType())
	}
	return newKey(domainKey.String(), KEY_TYPE_SUBDOMAIN, subKey)
}

func NewUseCaseKey(subdomainKey Key, subKey string) (key Key, err error) {
	// The parent must be a subdomain.
	if subdomainKey.KeyType() != KEY_TYPE_SUBDOMAIN {
		return Key{}, errors.Errorf("parent key cannot be of type '%s' for 'usecase' key", subdomainKey.KeyType())
	}
	return newKey(subdomainKey.String(), KEY_TYPE_USE_CASE, subKey)
}

func NewClassKey(subdomainKey Key, subKey string) (key Key, err error) {
	// The parent must be a subdomain.
	if subdomainKey.KeyType() != KEY_TYPE_SUBDOMAIN {
		return Key{}, errors.Errorf("parent key cannot be of type '%s' for 'class' key", subdomainKey.KeyType())
	}
	return newKey(subdomainKey.String(), KEY_TYPE_CLASS, subKey)
}

func NewGeneralizationKey(subdomainKey Key, subKey string) (key Key, err error) {
	// The parent must be a subdomain.
	if subdomainKey.KeyType() != KEY_TYPE_SUBDOMAIN {
		return Key{}, errors.Errorf("parent key cannot be of type '%s' for 'generalization' key", subdomainKey.KeyType())
	}
	return newKey(subdomainKey.String(), KEY_TYPE_GENERALIZATION, subKey)
}

func NewScenarioKey(useCaseKey Key, subKey string) (key Key, err error) {
	// The parent must be a use case.
	if useCaseKey.KeyType() != KEY_TYPE_USE_CASE {
		return Key{}, errors.Errorf("parent key cannot be of type '%s' for 'scenario' key", useCaseKey.KeyType())
	}
	return newKey(useCaseKey.String(), KEY_TYPE_SCENARIO, subKey)
}

func NewScenarioObjectKey(scenarioKey Key, subKey string) (key Key, err error) {
	// The parent must be a scenario.
	if scenarioKey.KeyType() != KEY_TYPE_SCENARIO {
		return Key{}, errors.Errorf("parent key cannot be of type '%s' for 'sobject' key", scenarioKey.KeyType())
	}
	return newKey(scenarioKey.String(), KEY_TYPE_SCENARIO_OBJECT, subKey)
}

func NewStateKey(classKey Key, subKey string) (key Key, err error) {
	// The parent must be a class.
	if classKey.KeyType() != KEY_TYPE_CLASS {
		return Key{}, errors.Errorf("parent key cannot be of type '%s' for 'state' key", classKey.KeyType())
	}
	return newKey(classKey.String(), KEY_TYPE_STATE, subKey)
}

func NewEventKey(classKey Key, subKey string) (key Key, err error) {
	// The parent must be a class.
	if classKey.KeyType() != KEY_TYPE_CLASS {
		return Key{}, errors.Errorf("parent key cannot be of type '%s' for 'event' key", classKey.KeyType())
	}
	return newKey(classKey.String(), KEY_TYPE_EVENT, subKey)
}

func NewGuardKey(classKey Key, subKey string) (key Key, err error) {
	// The parent must be a class.
	if classKey.KeyType() != KEY_TYPE_CLASS {
		return Key{}, errors.Errorf("parent key cannot be of type '%s' for 'guard' key", classKey.KeyType())
	}
	return newKey(classKey.String(), KEY_TYPE_GUARD, subKey)
}

func NewActionKey(classKey Key, subKey string) (key Key, err error) {
	// The parent must be a class.
	if classKey.KeyType() != KEY_TYPE_CLASS {
		return Key{}, errors.Errorf("parent key cannot be of type '%s' for 'action' key", classKey.KeyType())
	}
	return newKey(classKey.String(), KEY_TYPE_ACTION, subKey)
}

func NewTransitionKey(classKey Key, subKey string) (key Key, err error) {
	// The parent must be a class.
	if classKey.KeyType() != KEY_TYPE_CLASS {
		return Key{}, errors.Errorf("parent key cannot be of type '%s' for 'transition' key", classKey.KeyType())
	}
	return newKey(classKey.String(), KEY_TYPE_TRANSITION, subKey)
}

func NewClassAssociationKey(parentKey, fromClassKey, toClassKey Key) (key Key, err error) {
	// Both must be classes.
	if fromClassKey.KeyType() != KEY_TYPE_CLASS {
		return Key{}, errors.Errorf("from class key cannot be of type '%s' for 'cassociation' key", fromClassKey.KeyType())
	}
	if toClassKey.KeyType() != KEY_TYPE_CLASS {
		return Key{}, errors.Errorf("to class key cannot be of type '%s' for 'cassociation' key", toClassKey.KeyType())
	}

	var subKey, subKey2 string

	switch parentKey.KeyType() {
	case KEY_TYPE_SUBDOMAIN:
		// Parent is a subdomain - both classes must be in this subdomain.
		// subKey and subKey2 are "class/something" portions.
		fromPrefix := parentKey.String() + "/"
		toPrefix := parentKey.String() + "/"
		fromStr := fromClassKey.String()
		toStr := toClassKey.String()

		if !strings.HasPrefix(fromStr, fromPrefix) {
			return Key{}, errors.Errorf("from class key '%s' is not in subdomain '%s'", fromStr, parentKey.String())
		}
		if !strings.HasPrefix(toStr, toPrefix) {
			return Key{}, errors.Errorf("to class key '%s' is not in subdomain '%s'", toStr, parentKey.String())
		}

		subKey = strings.TrimPrefix(fromStr, fromPrefix)
		subKey2 = strings.TrimPrefix(toStr, toPrefix)

	case KEY_TYPE_DOMAIN:
		// Parent is a domain - both classes must be in this domain but NOT in the same subdomain.
		// subKey and subKey2 are "subdomain/where/class/something" portions.
		fromPrefix := parentKey.String() + "/"
		toPrefix := parentKey.String() + "/"
		fromStr := fromClassKey.String()
		toStr := toClassKey.String()

		if !strings.HasPrefix(fromStr, fromPrefix) {
			return Key{}, errors.Errorf("from class key '%s' is not in domain '%s'", fromStr, parentKey.String())
		}
		if !strings.HasPrefix(toStr, toPrefix) {
			return Key{}, errors.Errorf("to class key '%s' is not in domain '%s'", toStr, parentKey.String())
		}

		// Extract the subdomain portions to verify they're different.
		fromRemainder := strings.TrimPrefix(fromStr, fromPrefix)
		toRemainder := strings.TrimPrefix(toStr, toPrefix)

		// Parse to find subdomain - format is "subdomain/subdomainName/..."
		fromParts := strings.SplitN(fromRemainder, "/", 3)
		toParts := strings.SplitN(toRemainder, "/", 3)

		if len(fromParts) < 2 || fromParts[0] != KEY_TYPE_SUBDOMAIN {
			return Key{}, errors.Errorf("from class key '%s' does not have expected subdomain structure", fromStr)
		}
		if len(toParts) < 2 || toParts[0] != KEY_TYPE_SUBDOMAIN {
			return Key{}, errors.Errorf("to class key '%s' does not have expected subdomain structure", toStr)
		}

		// Classes in the same subdomain should use subdomain as parent instead.
		if fromParts[1] == toParts[1] {
			return Key{}, errors.Errorf("classes are in the same subdomain '%s', use subdomain as parent instead", fromParts[1])
		}

		subKey = fromRemainder
		subKey2 = toRemainder

	case "":
		// Parent is model (empty key) - classes must be in different domains.
		// subKey and subKey2 are the full class keys.
		fromStr := fromClassKey.String()
		toStr := toClassKey.String()

		// Parse to find domain - format is "domain/domainName/..."
		fromParts := strings.SplitN(fromStr, "/", 3)
		toParts := strings.SplitN(toStr, "/", 3)

		if len(fromParts) < 2 || fromParts[0] != KEY_TYPE_DOMAIN {
			return Key{}, errors.Errorf("from class key '%s' does not have expected domain structure", fromStr)
		}
		if len(toParts) < 2 || toParts[0] != KEY_TYPE_DOMAIN {
			return Key{}, errors.Errorf("to class key '%s' does not have expected domain structure", toStr)
		}

		// Classes in the same domain should use domain as parent instead.
		if fromParts[1] == toParts[1] {
			return Key{}, errors.Errorf("classes are in the same domain '%s', use domain as parent instead", fromParts[1])
		}

		subKey = fromStr
		subKey2 = toStr

	default:
		return Key{}, errors.Errorf("parent key cannot be of type '%s' for 'cassociation' key, must be subdomain, domain, or empty (model)", parentKey.KeyType())
	}

	// For model parent (empty key), use empty string; otherwise use the key string.
	var parentKeyStr string
	if parentKey.KeyType() != "" {
		parentKeyStr = parentKey.String()
	}

	return newKeyWithSubKey2(parentKeyStr, KEY_TYPE_CLASS_ASSOCIATION, subKey, &subKey2)
}

func NewAttributeKey(classKey Key, subKey string) (key Key, err error) {
	// The parent must be a class.
	if classKey.KeyType() != KEY_TYPE_CLASS {
		return Key{}, errors.Errorf("parent key cannot be of type '%s' for 'attribute' key", classKey.KeyType())
	}
	return newKey(classKey.String(), KEY_TYPE_ATTRIBUTE, subKey)
}

func NewStateActionKey(stateKey Key, subKey string) (key Key, err error) {
	// The parent must be a state.
	if stateKey.KeyType() != KEY_TYPE_STATE {
		return Key{}, errors.Errorf("parent key cannot be of type '%s' for 'saction' key", stateKey.KeyType())
	}
	return newKey(stateKey.String(), KEY_TYPE_STATE_ACTION, subKey)
}
