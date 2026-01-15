package identity

import "github.com/pkg/errors"

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
	KEY_TYPE_USE_CASE = "usecase"

	// remaining
	KEY_TYPE_CLASS             = "class"
	KEY_TYPE_STATE             = "state"
	KEY_TYPE_EVENT             = "event"
	KEY_TYPE_GUARD             = "guard"
	KEY_TYPE_ACTION            = "action"
	KEY_TYPE_TRANSITION        = "transition"
	KEY_TYPE_GENERALIZATION    = "generalization"
	KEY_TYPE_SCENARIO          = "scenario"
	KEY_TYPE_SCENARIO_OBJECT   = "sobject"
	KEY_TYPE_CLASS_ASSOCIATION = "cassociation"
	KEY_TYPE_ATTRIBUTE         = "attribute"
	KEY_TYPE_STATE_ACTION      = "saction"
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

func NewClassAssociationKey(domainKey Key, subKey string) (key Key, err error) {
	// The parent must be a domain.
	if domainKey.KeyType() != KEY_TYPE_DOMAIN {
		return Key{}, errors.Errorf("parent key cannot be of type '%s' for 'cassociation' key", domainKey.KeyType())
	}
	return newKey(domainKey.String(), KEY_TYPE_CLASS_ASSOCIATION, subKey)
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
