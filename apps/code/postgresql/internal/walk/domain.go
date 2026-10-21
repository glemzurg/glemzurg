package walk

import (
	"fmt"
	"path"
	"sort"

	"github.com/glemzurg/glemzurg/apps/requirements/req/pkg/core"
)

func modelAssociations(ctx *context, level, dir, labelPrefix string, model core.Model) ([]AssocVisit, error) {
	return ctx.associations(level, dir, labelPrefix, func(emit func(fromParent, fromSub, toParent, toSub, name string)) {
		for _, assoc := range model.ClassAssociations {
			emit(assoc.FromClassKey.ParentKey, assoc.FromClassKey.SubKey, assoc.ToClassKey.ParentKey, assoc.ToClassKey.SubKey, assoc.Key.SubKey3)
		}
	})
}

func oneDomain(model core.Model, domainKey string, ctx *context) (DomainVisit, error) {
	for key, domain := range model.Domains {
		if key.SubKey != domainKey {
			continue
		}
		visit := DomainVisit{Key: domainKey, Name: domain.Name}
		if note, ok := ctx.bind(path.Join("domains", domainKey, "domain.json")); ok {
			visit.Include = copyBool(note.Include)
		}
		assocs, err := ctx.associations("domain", path.Join("domains", domainKey, "class_associations"), domainKey+"/", func(emit func(fromParent, fromSub, toParent, toSub, name string)) {
			for _, assoc := range domain.ClassAssociations {
				emit(assoc.FromClassKey.ParentKey, assoc.FromClassKey.SubKey, assoc.ToClassKey.ParentKey, assoc.ToClassKey.SubKey, assoc.Key.SubKey3)
			}
		})
		if err != nil {
			return DomainVisit{}, err
		}
		visit.Associations = assocs
		visit.Subdomains, err = subdomains(model, domainKey, ctx)
		if err != nil {
			return DomainVisit{}, err
		}
		return visit, nil
	}
	return DomainVisit{}, fmt.Errorf("domain %s is missing", domainKey)
}

func subdomains(model core.Model, domainKey string, ctx *context) ([]SubdomainVisit, error) {
	var keys []string
	for key, domain := range model.Domains {
		if key.SubKey != domainKey {
			continue
		}
		for subKey := range domain.Subdomains {
			keys = append(keys, subKey.SubKey)
		}
	}
	sort.Strings(keys)
	var visits []SubdomainVisit
	for _, subKey := range keys {
		visit, err := oneSubdomain(model, domainKey, subKey, ctx)
		if err != nil {
			return nil, err
		}
		visits = append(visits, visit)
	}
	return visits, nil
}

func oneSubdomain(model core.Model, domainKey, subKey string, ctx *context) (SubdomainVisit, error) {
	for key, domain := range model.Domains {
		if key.SubKey != domainKey {
			continue
		}
		for skey, sub := range domain.Subdomains {
			if skey.SubKey != subKey {
				continue
			}
			visit := SubdomainVisit{Key: subKey, Name: sub.Name}
			if note, ok := ctx.bind(path.Join("domains", domainKey, "subdomains", subKey, "subdomain.json")); ok {
				visit.Include = copyBool(note.Include)
			}
			classes, err := classes(model, domainKey, subKey, ctx)
			if err != nil {
				return SubdomainVisit{}, err
			}
			visit.Classes = classes
			dir := path.Join("domains", domainKey, "subdomains", subKey, "class_associations")
			prefix := domainKey + "/" + subKey + "/"
			visit.Associations, err = ctx.associations("subdomain", dir, prefix, func(emit func(fromParent, fromSub, toParent, toSub, name string)) {
				for _, assoc := range sub.ClassAssociations {
					emit(assoc.FromClassKey.ParentKey, assoc.FromClassKey.SubKey, assoc.ToClassKey.ParentKey, assoc.ToClassKey.SubKey, assoc.Key.SubKey3)
				}
			})
			if err != nil {
				return SubdomainVisit{}, err
			}
			return visit, nil
		}
	}
	return SubdomainVisit{}, fmt.Errorf("subdomain %s/%s is missing", domainKey, subKey)
}
