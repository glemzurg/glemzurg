package model_data_type

import (
	"strings"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
)

func t_StrPtr(s string) *string {
	return &s
}

// t_dtKey builds a valid KEY_TYPE_DATA_TYPE identity.Key for tests. The "name" is
// used to scope a synthetic attribute parent so different calls produce different keys.
// The name is normalized to a lowercase identifier so any test string works.
func t_dtKey(name string) identity.Key {
	domainKey, err := identity.NewDomainKey("d")
	if err != nil {
		panic(err)
	}
	subdomainKey, err := identity.NewSubdomainKey(domainKey, "s")
	if err != nil {
		panic(err)
	}
	classKey, err := identity.NewClassKey(subdomainKey, "c")
	if err != nil {
		panic(err)
	}
	attrKey, err := identity.NewAttributeKey(classKey, t_normalizeName(name))
	if err != nil {
		panic(err)
	}
	dtKey, err := identity.NewDataTypeKey(attrKey, "")
	if err != nil {
		panic(err)
	}
	return dtKey
}

// t_nestedDtKey builds a nested data_type key (field child) under parentKey.
func t_nestedDtKey(parentKey identity.Key, field string) identity.Key {
	k, err := identity.NewDataTypeKey(parentKey, t_normalizeName(field))
	if err != nil {
		panic(err)
	}
	return k
}

func t_normalizeName(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	s = strings.ReplaceAll(s, " ", "_")
	s = strings.ReplaceAll(s, "-", "_")
	if s == "" {
		s = "x"
	}
	return s
}
