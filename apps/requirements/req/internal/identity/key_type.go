package identity

import (
	"strings"

	"github.com/pkg/errors"
)

const (

	// Models do not have a key type.
	// It is a string that is unique in the system.

	// Keys without parents (parent is the model itself).
	KEY_TYPE_ACTOR                = "actor"
	KEY_TYPE_ACTOR_GENERALIZATION = "ageneralization"
	KEY_TYPE_DOMAIN               = "domain"
	KEY_TYPE_DOMAIN_ASSOCIATION   = "dassociation"
	KEY_TYPE_GLOBAL_FUNCTION      = "gfunc"
	KEY_TYPE_INVARIANT            = "invariant"

	// Keys with domain parents.
	KEY_TYPE_SUBDOMAIN = "subdomain"

	// Keys with subdomain parents.
	KEY_TYPE_USE_CASE                 = "usecase"
	KEY_TYPE_USE_CASE_GENERALIZATION  = "ucgeneralization"
	KEY_TYPE_CLASS                    = "class"
	KEY_TYPE_CLASS_GENERALIZATION     = "cgeneralization"

	// Keys with model, domain, subdomain parents.
	KEY_TYPE_CLASS_ASSOCIATION = "cassociation"

	// Keys with class parents.
	KEY_TYPE_ATTRIBUTE  = "attribute"
	KEY_TYPE_STATE      = "state"
	KEY_TYPE_EVENT      = "event"
	KEY_TYPE_GUARD      = "guard" // Used for both guards and the logic inside guards.
	KEY_TYPE_ACTION     = "action"
	KEY_TYPE_QUERY      = "query"
	KEY_TYPE_TRANSITION = "transition"

	// Keys with class attribute parents.
	KEY_TYPE_ATTRIBUTE_DERIVATION = "aderive"

	// Keys with state parents.
	KEY_TYPE_STATE_ACTION = "saction"

	// Keys with actions as parents.
	KEY_TYPE_ACTION_REQUIRE   = "arequire"
	KEY_TYPE_ACTION_GUARANTEE = "aguarantee"
	KEY_TYPE_ACTION_SAFETY    = "asafety"

	// Keys with queries as parents.
	KEY_TYPE_QUERY_REQUIRE   = "qrequire"
	KEY_TYPE_QUERY_GUARANTEE = "qguarantee"

	// Keys with use case parents.
	KEY_TYPE_SCENARIO = "scenario"

	// Keys with scenario parents.
	KEY_TYPE_SCENARIO_OBJECT = "sobject"
)

func NewActorKey(subKey string) (key Key, err error) {
	return newRootKey(KEY_TYPE_ACTOR, subKey)
}

func NewActorGeneralizationKey(subKey string) (key Key, err error) {
	return newRootKey(KEY_TYPE_ACTOR_GENERALIZATION, subKey)
}

func NewDomainKey(subKey string) (key Key, err error) {
	return newRootKey(KEY_TYPE_DOMAIN, subKey)
}

func NewGlobalFunctionKey(subKey string) (key Key, err error) {
	return newRootKey(KEY_TYPE_GLOBAL_FUNCTION, subKey)
}

func NewInvariantKey(subKey string) (key Key, err error) {
	return newRootKey(KEY_TYPE_INVARIANT, subKey)
}

func NewDomainAssociationKey(problemDomainKey, solutionDomainKey Key) (key Key, err error) {
	// Both must be domains.
	if problemDomainKey.GetKeyType() != KEY_TYPE_DOMAIN {
		return Key{}, errors.Errorf("problem domain key cannot be of type '%s' for 'dassociation' key", problemDomainKey.GetKeyType())
	}
	if solutionDomainKey.GetKeyType() != KEY_TYPE_DOMAIN {
		return Key{}, errors.Errorf("solution domain key cannot be of type '%s' for 'dassociation' key", solutionDomainKey.GetKeyType())
	}
	// No parent, problem domain subKey as subKey, solution domain subKey as subKey2.
	return newKeyWithSubKey2("", KEY_TYPE_DOMAIN_ASSOCIATION, problemDomainKey.GetSubKey(), solutionDomainKey.GetSubKey())
}

func NewSubdomainKey(domainKey Key, subKey string) (key Key, err error) {
	// The parent must be a domain.
	if domainKey.GetKeyType() != KEY_TYPE_DOMAIN {
		return Key{}, errors.Errorf("parent key cannot be of type '%s' for 'subdomain' key", domainKey.GetKeyType())
	}
	return newKey(domainKey.String(), KEY_TYPE_SUBDOMAIN, subKey)
}

func NewUseCaseKey(subdomainKey Key, subKey string) (key Key, err error) {
	// The parent must be a subdomain.
	if subdomainKey.GetKeyType() != KEY_TYPE_SUBDOMAIN {
		return Key{}, errors.Errorf("parent key cannot be of type '%s' for 'usecase' key", subdomainKey.GetKeyType())
	}
	return newKey(subdomainKey.String(), KEY_TYPE_USE_CASE, subKey)
}

func NewUseCaseGeneralizationKey(subdomainKey Key, subKey string) (key Key, err error) {
	// The parent must be a subdomain.
	if subdomainKey.GetKeyType() != KEY_TYPE_SUBDOMAIN {
		return Key{}, errors.Errorf("parent key cannot be of type '%s' for 'ucgeneralization' key", subdomainKey.GetKeyType())
	}
	return newKey(subdomainKey.String(), KEY_TYPE_USE_CASE_GENERALIZATION, subKey)
}

func NewClassKey(subdomainKey Key, subKey string) (key Key, err error) {
	// The parent must be a subdomain.
	if subdomainKey.GetKeyType() != KEY_TYPE_SUBDOMAIN {
		return Key{}, errors.Errorf("parent key cannot be of type '%s' for 'class' key", subdomainKey.GetKeyType())
	}
	return newKey(subdomainKey.String(), KEY_TYPE_CLASS, subKey)
}

func NewGeneralizationKey(subdomainKey Key, subKey string) (key Key, err error) {
	// The parent must be a subdomain.
	if subdomainKey.GetKeyType() != KEY_TYPE_SUBDOMAIN {
		return Key{}, errors.Errorf("parent key cannot be of type '%s' for 'generalization' key", subdomainKey.GetKeyType())
	}
	return newKey(subdomainKey.String(), KEY_TYPE_CLASS_GENERALIZATION, subKey)
}

func NewScenarioKey(useCaseKey Key, subKey string) (key Key, err error) {
	// The parent must be a use case.
	if useCaseKey.GetKeyType() != KEY_TYPE_USE_CASE {
		return Key{}, errors.Errorf("parent key cannot be of type '%s' for 'scenario' key", useCaseKey.GetKeyType())
	}
	return newKey(useCaseKey.String(), KEY_TYPE_SCENARIO, subKey)
}

func NewScenarioObjectKey(scenarioKey Key, subKey string) (key Key, err error) {
	// The parent must be a scenario.
	if scenarioKey.GetKeyType() != KEY_TYPE_SCENARIO {
		return Key{}, errors.Errorf("parent key cannot be of type '%s' for 'sobject' key", scenarioKey.GetKeyType())
	}
	return newKey(scenarioKey.String(), KEY_TYPE_SCENARIO_OBJECT, subKey)
}

func NewStateKey(classKey Key, subKey string) (key Key, err error) {
	// The parent must be a class.
	if classKey.GetKeyType() != KEY_TYPE_CLASS {
		return Key{}, errors.Errorf("parent key cannot be of type '%s' for 'state' key", classKey.GetKeyType())
	}
	return newKey(classKey.String(), KEY_TYPE_STATE, subKey)
}

func NewEventKey(classKey Key, subKey string) (key Key, err error) {
	// The parent must be a class.
	if classKey.GetKeyType() != KEY_TYPE_CLASS {
		return Key{}, errors.Errorf("parent key cannot be of type '%s' for 'event' key", classKey.GetKeyType())
	}
	return newKey(classKey.String(), KEY_TYPE_EVENT, subKey)
}

func NewGuardKey(classKey Key, subKey string) (key Key, err error) {
	// The parent must be a class.
	if classKey.GetKeyType() != KEY_TYPE_CLASS {
		return Key{}, errors.Errorf("parent key cannot be of type '%s' for 'guard' key", classKey.GetKeyType())
	}
	return newKey(classKey.String(), KEY_TYPE_GUARD, subKey)
}

func NewActionKey(classKey Key, subKey string) (key Key, err error) {
	// The parent must be a class.
	if classKey.GetKeyType() != KEY_TYPE_CLASS {
		return Key{}, errors.Errorf("parent key cannot be of type '%s' for 'action' key", classKey.GetKeyType())
	}
	return newKey(classKey.String(), KEY_TYPE_ACTION, subKey)
}

func NewQueryKey(classKey Key, subKey string) (key Key, err error) {
	// The parent must be a class.
	if classKey.GetKeyType() != KEY_TYPE_CLASS {
		return Key{}, errors.Errorf("parent key cannot be of type '%s' for 'query' key", classKey.GetKeyType())
	}
	return newKey(classKey.String(), KEY_TYPE_QUERY, subKey)
}

func NewTransitionKey(classKey Key, from, event, guard, action, to string) (key Key, err error) {
	// The parent must be a class.
	if classKey.GetKeyType() != KEY_TYPE_CLASS {
		return Key{}, errors.Errorf("parent key cannot be of type '%s' for 'transition' key", classKey.GetKeyType())
	}
	// Event cannot be blank.
	if event == "" {
		return Key{}, errors.New("event cannot be empty for transition key")
	}
	// Default from to "initial" when blank.
	if from == "" {
		from = "initial"
	}
	// Default to to "final" when blank.
	if to == "" {
		to = "final"
	}
	// Cannot transition directly from initial to final.
	if from == "initial" && to == "final" {
		return Key{}, errors.New("cannot transition directly from initial to final")
	}
	// SubKey format: from/event/guard/action/to
	subKey := from + "/" + event + "/" + guard + "/" + action + "/" + to
	return newKey(classKey.String(), KEY_TYPE_TRANSITION, subKey)
}

func NewClassAssociationKey(parentKey, fromClassKey, toClassKey Key, name string) (key Key, err error) {
	// Both must be classes.
	if fromClassKey.GetKeyType() != KEY_TYPE_CLASS {
		return Key{}, errors.Errorf("from class key cannot be of type '%s' for 'cassociation' key", fromClassKey.GetKeyType())
	}
	if toClassKey.GetKeyType() != KEY_TYPE_CLASS {
		return Key{}, errors.Errorf("to class key cannot be of type '%s' for 'cassociation' key", toClassKey.GetKeyType())
	}

	// Name is required.
	if strings.TrimSpace(name) == "" {
		return Key{}, errors.New("name cannot be empty for class association key")
	}

	// Convert name to distilled format: trim, lowercase, replace internal spaces with underscores.
	subKey3 := distillName(name)

	var subKey, subKey2 string

	switch parentKey.GetKeyType() {
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
		return Key{}, errors.Errorf("parent key cannot be of type '%s' for 'cassociation' key, must be subdomain, domain, or empty (model)", parentKey.GetKeyType())
	}

	// For model parent (empty key), use empty string; otherwise use the key string.
	var parentKeyStr string
	if parentKey.GetKeyType() != "" {
		parentKeyStr = parentKey.String()
	}

	return newKeyWithSubKey3(parentKeyStr, KEY_TYPE_CLASS_ASSOCIATION, subKey, subKey2, subKey3)
}

// distillName converts a name to distilled format: trim, lowercase, replace internal spaces with underscores,
// and replace forward slashes with tildes (since slashes are used as key separators).
func distillName(name string) string {
	name = strings.TrimSpace(name)
	name = strings.ToLower(name)
	name = strings.ReplaceAll(name, " ", "_")
	name = strings.ReplaceAll(name, "/", "~")
	return name
}

func NewAttributeKey(classKey Key, subKey string) (key Key, err error) {
	// The parent must be a class.
	if classKey.GetKeyType() != KEY_TYPE_CLASS {
		return Key{}, errors.Errorf("parent key cannot be of type '%s' for 'attribute' key", classKey.GetKeyType())
	}
	return newKey(classKey.String(), KEY_TYPE_ATTRIBUTE, subKey)
}

func NewAttributeDerivationKey(attributeKey Key, subKey string) (key Key, err error) {
	// The parent must be an attribute.
	if attributeKey.GetKeyType() != KEY_TYPE_ATTRIBUTE {
		return Key{}, errors.Errorf("parent key cannot be of type '%s' for 'aderive' key", attributeKey.GetKeyType())
	}
	return newKey(attributeKey.String(), KEY_TYPE_ATTRIBUTE_DERIVATION, subKey)
}

func NewStateActionKey(stateKey Key, when, subKey string) (key Key, err error) {
	// The parent must be a state.
	if stateKey.GetKeyType() != KEY_TYPE_STATE {
		return Key{}, errors.Errorf("parent key cannot be of type '%s' for 'saction' key", stateKey.GetKeyType())
	}
	// When cannot be empty.
	if when == "" {
		return Key{}, errors.New("when cannot be empty for state action key")
	}
	return newKey(stateKey.String(), KEY_TYPE_STATE_ACTION, when+"/"+subKey)
}

func NewActionRequireKey(actionKey Key, subKey string) (key Key, err error) {
	// The parent must be an action.
	if actionKey.GetKeyType() != KEY_TYPE_ACTION {
		return Key{}, errors.Errorf("parent key cannot be of type '%s' for 'arequire' key", actionKey.GetKeyType())
	}
	return newKey(actionKey.String(), KEY_TYPE_ACTION_REQUIRE, subKey)
}

func NewActionGuaranteeKey(actionKey Key, subKey string) (key Key, err error) {
	// The parent must be an action.
	if actionKey.GetKeyType() != KEY_TYPE_ACTION {
		return Key{}, errors.Errorf("parent key cannot be of type '%s' for 'aguarantee' key", actionKey.GetKeyType())
	}
	return newKey(actionKey.String(), KEY_TYPE_ACTION_GUARANTEE, subKey)
}

func NewActionSafetyKey(actionKey Key, subKey string) (key Key, err error) {
	// The parent must be an action.
	if actionKey.GetKeyType() != KEY_TYPE_ACTION {
		return Key{}, errors.Errorf("parent key cannot be of type '%s' for 'asafety' key", actionKey.GetKeyType())
	}
	return newKey(actionKey.String(), KEY_TYPE_ACTION_SAFETY, subKey)
}

func NewQueryRequireKey(queryKey Key, subKey string) (key Key, err error) {
	// The parent must be a query.
	if queryKey.GetKeyType() != KEY_TYPE_QUERY {
		return Key{}, errors.Errorf("parent key cannot be of type '%s' for 'qrequire' key", queryKey.GetKeyType())
	}
	return newKey(queryKey.String(), KEY_TYPE_QUERY_REQUIRE, subKey)
}

func NewQueryGuaranteeKey(queryKey Key, subKey string) (key Key, err error) {
	// The parent must be a query.
	if queryKey.GetKeyType() != KEY_TYPE_QUERY {
		return Key{}, errors.Errorf("parent key cannot be of type '%s' for 'qguarantee' key", queryKey.GetKeyType())
	}
	return newKey(queryKey.String(), KEY_TYPE_QUERY_GUARANTEE, subKey)
}
