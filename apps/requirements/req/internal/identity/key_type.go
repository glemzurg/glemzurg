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
	KEY_TYPE_CLASS          = "class"
	KEY_TYPE_STATE          = "state"
	KEY_TYPE_EVENT          = "event"
	KEY_TYPE_GUARD          = "guard"
	KEY_TYPE_GENERALIZATION = "generalization"
	KEY_TYPE_SCENARIO       = "scenario"
)

func NewActorKey(subKey string) (key Key, err error) {
	return newRootKey(KEY_TYPE_ACTOR, subKey)
}

func NewDomainKey(subKey string) (key Key, err error) {
	return newRootKey(KEY_TYPE_DOMAIN, subKey)
}

func NewDomainAssociationKey(domainKey Key, subKey string) (key Key, err error) {
	// The parent must be a domain.
	if domainKey.KeyType() != KEY_TYPE_DOMAIN {
		return Key{}, errors.Errorf("parent key cannot be of type '%s' for 'association' key", domainKey.KeyType())
	}
	return newKey(domainKey.String(), KEY_TYPE_DOMAIN_ASSOCIATION, subKey)
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
