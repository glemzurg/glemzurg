package model_class

import (
	"fmt"
	"strings"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/pkg/errors"
)

// ClassKeysFromAssociationKey reconstructs the from- and to-class keys encoded in a class association key.
func ClassKeysFromAssociationKey(assocKey identity.Key) (fromClassKey, toClassKey identity.Key, err error) {
	if assocKey.KeyType != identity.KEY_TYPE_CLASS_ASSOCIATION {
		return identity.Key{}, identity.Key{}, errors.Errorf("key %q is not a class association", assocKey.String())
	}
	parentPrefix := ""
	if assocKey.ParentKey != "" {
		parentPrefix = assocKey.ParentKey + "/"
	}
	fromClassKey, err = identity.ParseKey(parentPrefix + assocKey.SubKey)
	if err != nil {
		return identity.Key{}, identity.Key{}, errors.WithStack(err)
	}
	toClassKey, err = identity.ParseKey(parentPrefix + assocKey.SubKey2)
	if err != nil {
		return identity.Key{}, identity.Key{}, errors.WithStack(err)
	}
	return fromClassKey, toClassKey, nil
}

// RelativeClassAssociationKey returns the shortest YAML key for a class invariant over_association_key.
// Same-subdomain associations use from_class_subkey/distilled_name when the owner is the to-class.
func RelativeClassAssociationKey(ownerClassKey, assocKey identity.Key) (string, error) {
	fromClassKey, toClassKey, err := ClassKeysFromAssociationKey(assocKey)
	if err != nil {
		return "", err
	}
	distilled := assocKey.SubKey3
	fromSub, err := classSubKeyFromClassKey(fromClassKey)
	if err != nil {
		return "", err
	}
	toSub, err := classSubKeyFromClassKey(toClassKey)
	if err != nil {
		return "", err
	}
	if ownerClassKey == toClassKey && fromClassKey.ParentKey == toClassKey.ParentKey {
		return fromSub + "/" + distilled, nil
	}
	if ownerClassKey == fromClassKey && fromClassKey.ParentKey == toClassKey.ParentKey {
		return fromSub + "/" + toSub + "/" + distilled, nil
	}
	return assocKey.String(), nil
}

// ResolveClassAssociationKeyFromRelative resolves over_association_key YAML/JSON relative to ownerClassKey.
// Full keys parse directly. Two-part from_subkey/distilled_name means from-class to owner-class.
func ResolveClassAssociationKeyFromRelative(subdomainKey, ownerClassKey identity.Key, keyStr string) (identity.Key, error) {
	keyStr = strings.TrimSpace(keyStr)
	if keyStr == "" {
		return identity.Key{}, errors.New("over_association_key is empty")
	}
	if strings.Contains(keyStr, identity.KEY_TYPE_CLASS_ASSOCIATION) || strings.HasPrefix(keyStr, identity.KEY_TYPE_DOMAIN+"/") {
		key, err := identity.ParseKey(keyStr)
		if err != nil {
			return identity.Key{}, errors.WithStack(err)
		}
		return key, nil
	}

	parts := strings.Split(keyStr, "/")
	switch len(parts) {
	case 2:
		fromClassKey, err := resolveClassKeyFromRelative(subdomainKey, parts[0])
		if err != nil {
			return identity.Key{}, err
		}
		parentKey, err := associationParentForClasses(subdomainKey, fromClassKey, ownerClassKey)
		if err != nil {
			return identity.Key{}, err
		}
		return identity.NewClassAssociationKey(parentKey, fromClassKey, ownerClassKey, parts[1])
	case 3:
		fromClassKey, err := resolveClassKeyFromRelative(subdomainKey, parts[0])
		if err != nil {
			return identity.Key{}, err
		}
		toClassKey, err := resolveClassKeyFromRelative(subdomainKey, parts[1])
		if err != nil {
			return identity.Key{}, err
		}
		parentKey, err := associationParentForClasses(subdomainKey, fromClassKey, toClassKey)
		if err != nil {
			return identity.Key{}, err
		}
		return identity.NewClassAssociationKey(parentKey, fromClassKey, toClassKey, parts[2])
	default:
		return identity.Key{}, fmt.Errorf("over_association_key %q: expected from_subkey/distilled_name or a full association key", keyStr)
	}
}

func classSubKeyFromClassKey(classKey identity.Key) (string, error) {
	if classKey.KeyType != identity.KEY_TYPE_CLASS {
		return "", errors.Errorf("key %q is not a class", classKey.String())
	}
	return classKey.SubKey, nil
}

func resolveClassKeyFromRelative(subdomainKey identity.Key, keyStr string) (identity.Key, error) {
	if !strings.Contains(keyStr, "/") {
		return identity.NewClassKey(subdomainKey, keyStr)
	}
	if strings.HasPrefix(keyStr, identity.KEY_TYPE_SUBDOMAIN+"/") {
		fullKeyStr := subdomainKey.ParentKey + "/" + keyStr
		return identity.ParseKey(fullKeyStr)
	}
	return identity.ParseKey(keyStr)
}

func associationParentForClasses(subdomainKey, fromClassKey, toClassKey identity.Key) (identity.Key, error) {
	if fromClassKey.ParentKey == toClassKey.ParentKey {
		return subdomainKey, nil
	}

	fromSubParsed, err := identity.ParseKey(fromClassKey.ParentKey)
	if err != nil {
		return identity.Key{}, errors.WithStack(err)
	}
	toSubParsed, err := identity.ParseKey(toClassKey.ParentKey)
	if err != nil {
		return identity.Key{}, errors.WithStack(err)
	}

	if fromSubParsed.ParentKey == toSubParsed.ParentKey {
		domainKey, err := identity.ParseKey(fromSubParsed.ParentKey)
		if err != nil {
			return identity.Key{}, errors.WithStack(err)
		}
		return domainKey, nil
	}

	return identity.Key{}, nil
}
