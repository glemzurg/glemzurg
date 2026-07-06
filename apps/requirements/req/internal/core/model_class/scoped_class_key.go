package model_class

import (
	"fmt"
	"strings"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/pkg/errors"
)

// ResolveScopedClassKey parses a scoped class reference relative to the authoring subdomain.
// Accepted forms: class, subdomain/class, domain/subdomain/class.
func ResolveScopedClassKey(authoringSubdomainKey identity.Key, scoped string) (identity.Key, error) {
	scoped = strings.TrimSpace(scoped)
	if scoped == "" {
		return identity.Key{}, errors.New("scoped class key is empty")
	}

	parts := strings.Split(scoped, "/")
	switch len(parts) {
	case 1:
		return identity.NewClassKey(authoringSubdomainKey, parts[0])
	case 2:
		fullKeyStr := authoringSubdomainKey.ParentKey + "/" + identity.KEY_TYPE_SUBDOMAIN + "/" + parts[0] + "/" + identity.KEY_TYPE_CLASS + "/" + parts[1]
		key, err := identity.ParseKey(fullKeyStr)
		if err != nil {
			return identity.Key{}, errors.WithStack(err)
		}
		return key, nil
	case 3:
		fullKeyStr := identity.KEY_TYPE_DOMAIN + "/" + parts[0] + "/" + identity.KEY_TYPE_SUBDOMAIN + "/" + parts[1] + "/" + identity.KEY_TYPE_CLASS + "/" + parts[2]
		key, err := identity.ParseKey(fullKeyStr)
		if err != nil {
			return identity.Key{}, errors.WithStack(err)
		}
		return key, nil
	default:
		return identity.Key{}, fmt.Errorf("scoped class key %q: expected class, subdomain/class, or domain/subdomain/class", scoped)
	}
}

// FormatScopedClassKey returns the shortest scoped class reference from one class to another.
func FormatScopedClassKey(fromClassKey, targetClassKey identity.Key) (string, error) {
	if fromClassKey.ParentKey == targetClassKey.ParentKey {
		return targetClassKey.SubKey, nil
	}

	fromDomain, _, _, err := classKeyLocation(fromClassKey)
	if err != nil {
		return "", err
	}
	targetDomain, targetSub, targetClass, err := classKeyLocation(targetClassKey)
	if err != nil {
		return "", err
	}

	if fromDomain == targetDomain {
		return targetSub + "/" + targetClass, nil
	}
	return targetDomain + "/" + targetSub + "/" + targetClass, nil
}

func classKeyLocation(classKey identity.Key) (domainName, subdomainName, className string, err error) {
	if classKey.KeyType != identity.KEY_TYPE_CLASS {
		return "", "", "", errors.Errorf("key %q is not a class", classKey.String())
	}
	className = classKey.SubKey

	subdomainKey, err := identity.ParseKey(classKey.ParentKey)
	if err != nil {
		return "", "", "", errors.WithStack(err)
	}
	if subdomainKey.KeyType != identity.KEY_TYPE_SUBDOMAIN {
		return "", "", "", errors.Errorf("class parent %q is not a subdomain", classKey.ParentKey)
	}
	subdomainName = subdomainKey.SubKey

	domainKey, err := identity.ParseKey(subdomainKey.ParentKey)
	if err != nil {
		return "", "", "", errors.WithStack(err)
	}
	if domainKey.KeyType != identity.KEY_TYPE_DOMAIN {
		return "", "", "", errors.Errorf("subdomain parent %q is not a domain", subdomainKey.ParentKey)
	}
	domainName = domainKey.SubKey

	return domainName, subdomainName, className, nil
}

// FormatDomainScopedClassKey returns subdomain/class for a class key.
func FormatDomainScopedClassKey(classKey identity.Key) (string, error) {
	_, subdomainName, className, err := classKeyLocation(classKey)
	if err != nil {
		return "", err
	}
	return subdomainName + "/" + className, nil
}

// FormatModelScopedClassKey returns domain/subdomain/class for a class key.
func FormatModelScopedClassKey(classKey identity.Key) (string, error) {
	domainName, subdomainName, className, err := classKeyLocation(classKey)
	if err != nil {
		return "", err
	}
	return domainName + "/" + subdomainName + "/" + className, nil
}

// defaultSubdomainKeySubkey matches the reserved single-subdomain key enforced by domain validation.
const defaultSubdomainKeySubkey = "default"

// FormatClassMarkdownDisplayName returns the class display name for markdown prose relative to a
// viewing subdomain. Scope is determined from keys; domain and subdomain labels use their display
// names. Same subdomain yields ClassName; cross-subdomain yields SubdomainName::ClassName;
// cross-domain yields DomainName::SubdomainName::ClassName. Classes in the default subdomain omit
// the subdomain segment because that subdomain is implicit for single-subdomain domains.
func FormatClassMarkdownDisplayName(
	viewerSubdomainKey identity.Key,
	targetClass Class,
	targetDomainDisplayName, targetSubdomainDisplayName string,
) (string, error) {
	if viewerSubdomainKey.String() == targetClass.Key.ParentKey {
		return targetClass.Name, nil
	}

	targetDomainKey, targetSubdomainKey, err := classKeyLocationKeys(targetClass.Key)
	if err != nil {
		return "", err
	}
	targetInDefaultSubdomain := targetSubdomainKey.SubKey == defaultSubdomainKeySubkey

	viewerSubdomain, err := identity.ParseKey(viewerSubdomainKey.String())
	if err != nil {
		return "", errors.WithStack(err)
	}
	if viewerSubdomain.KeyType != identity.KEY_TYPE_SUBDOMAIN {
		return "", errors.Errorf("viewer key %q is not a subdomain", viewerSubdomainKey.String())
	}
	viewerDomain, err := identity.ParseKey(viewerSubdomain.ParentKey)
	if err != nil {
		return "", errors.WithStack(err)
	}
	if viewerDomain.KeyType != identity.KEY_TYPE_DOMAIN {
		return "", errors.Errorf("viewer parent %q is not a domain", viewerSubdomain.ParentKey)
	}

	if viewerDomain.SubKey == targetDomainKey.SubKey {
		if targetInDefaultSubdomain {
			return targetClass.Name, nil
		}
		return targetSubdomainDisplayName + "::" + targetClass.Name, nil
	}
	if targetInDefaultSubdomain {
		return targetDomainDisplayName + "::" + targetClass.Name, nil
	}
	return targetDomainDisplayName + "::" + targetSubdomainDisplayName + "::" + targetClass.Name, nil
}

// FormatClassMermaidNamespaceSegments returns namespace path segments for a class on a class diagram
// relative to the viewing subdomain. An empty slice places the class in the diagram root.
// Cross-subdomain yields [SubdomainName]; cross-domain with a default subdomain yields
// [DomainName]; cross-domain otherwise yields [DomainName, SubdomainName].
func FormatClassMermaidNamespaceSegments(
	viewerSubdomainKey identity.Key,
	targetClass Class,
	targetDomainDisplayName, targetSubdomainDisplayName string,
) ([]string, error) {
	if viewerSubdomainKey.String() == targetClass.Key.ParentKey {
		return nil, nil
	}

	targetDomainKey, targetSubdomainKey, err := classKeyLocationKeys(targetClass.Key)
	if err != nil {
		return nil, err
	}
	targetInDefaultSubdomain := targetSubdomainKey.SubKey == defaultSubdomainKeySubkey

	viewerSubdomain, err := identity.ParseKey(viewerSubdomainKey.String())
	if err != nil {
		return nil, errors.WithStack(err)
	}
	if viewerSubdomain.KeyType != identity.KEY_TYPE_SUBDOMAIN {
		return nil, errors.Errorf("viewer key %q is not a subdomain", viewerSubdomainKey.String())
	}
	viewerDomain, err := identity.ParseKey(viewerSubdomain.ParentKey)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	if viewerDomain.KeyType != identity.KEY_TYPE_DOMAIN {
		return nil, errors.Errorf("viewer parent %q is not a domain", viewerSubdomain.ParentKey)
	}

	if viewerDomain.SubKey == targetDomainKey.SubKey {
		if targetInDefaultSubdomain {
			return nil, nil
		}
		return []string{targetSubdomainDisplayName}, nil
	}
	if targetInDefaultSubdomain {
		return []string{targetDomainDisplayName}, nil
	}
	return []string{targetDomainDisplayName, targetSubdomainDisplayName}, nil
}

func classKeyLocationKeys(classKey identity.Key) (domainKey, subdomainKey identity.Key, err error) {
	if classKey.KeyType != identity.KEY_TYPE_CLASS {
		return identity.Key{}, identity.Key{}, errors.Errorf("key %q is not a class", classKey.String())
	}
	subdomainKey, err = identity.ParseKey(classKey.ParentKey)
	if err != nil {
		return identity.Key{}, identity.Key{}, errors.WithStack(err)
	}
	if subdomainKey.KeyType != identity.KEY_TYPE_SUBDOMAIN {
		return identity.Key{}, identity.Key{}, errors.Errorf("class parent %q is not a subdomain", classKey.ParentKey)
	}
	domainKey, err = identity.ParseKey(subdomainKey.ParentKey)
	if err != nil {
		return identity.Key{}, identity.Key{}, errors.WithStack(err)
	}
	if domainKey.KeyType != identity.KEY_TYPE_DOMAIN {
		return identity.Key{}, identity.Key{}, errors.Errorf("subdomain parent %q is not a domain", subdomainKey.ParentKey)
	}
	return domainKey, subdomainKey, nil
}
