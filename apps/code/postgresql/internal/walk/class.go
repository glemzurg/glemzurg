package walk

import (
	"fmt"
	"path"
	"sort"
	"strings"

	"github.com/glemzurg/glemzurg/apps/requirements/req/pkg/core"
)

func classes(model core.Model, domainKey, subKey string, ctx *context) ([]ClassVisit, error) {
	keys, err := classSubKeys(model, domainKey, subKey)
	if err != nil {
		return nil, err
	}
	var visits []ClassVisit
	for _, classKey := range keys {
		visit, err := oneClass(model, domainKey, subKey, classKey, ctx)
		if err != nil {
			return nil, err
		}
		visits = append(visits, visit)
	}
	return visits, nil
}

func classSubKeys(model core.Model, domainKey, subKey string) ([]string, error) {
	var keys []string
	found := false
	for key, domain := range model.Domains {
		if key.SubKey != domainKey {
			continue
		}
		for skey, sub := range domain.Subdomains {
			if skey.SubKey != subKey {
				continue
			}
			found = true
			for classKey := range sub.Classes {
				keys = append(keys, classKey.SubKey)
			}
		}
	}
	if !found {
		return nil, fmt.Errorf("subdomain %s/%s is missing", domainKey, subKey)
	}
	sort.Strings(keys)
	return keys, nil
}

func oneClass(model core.Model, domainKey, subKey, classKey string, ctx *context) (ClassVisit, error) {
	for key, domain := range model.Domains {
		if key.SubKey != domainKey {
			continue
		}
		for skey, sub := range domain.Subdomains {
			if skey.SubKey != subKey {
				continue
			}
			for ckey, class := range sub.Classes {
				if ckey.SubKey != classKey {
					continue
				}
				rel := path.Join("domains", domainKey, "subdomains", subKey, "classes", classKey, "class.json")
				visit := ClassVisit{Key: classKey, Name: class.Name, Attributes: len(class.Attributes)}
				note, ok := ctx.bind(rel)
				if !ok {
					return visit, nil
				}
				visit.Include = copyBool(note.Include)
				var have []string
				for _, attr := range class.Attributes {
					have = append(have, attr.Key.SubKey)
				}
				if missing := unknownAttributeKeys(have, note.Attributes); len(missing) > 0 {
					return ClassVisit{}, fmt.Errorf("%s: unknown attribute %s", rel, strings.Join(missing, ", "))
				}
				for _, attr := range class.Attributes {
					noted, exists := note.Attributes[attr.Key.SubKey]
					if !exists {
						continue
					}
					visit.Noted = append(visit.Noted, AttrVisit{
						Key:     attr.Key.SubKey,
						Type:    noted.Type,
						Default: noted.Default,
						Include: copyBool(noted.Include),
					})
				}
				return visit, nil
			}
		}
	}
	return ClassVisit{}, fmt.Errorf("class %s/%s/%s is missing", domainKey, subKey, classKey)
}
