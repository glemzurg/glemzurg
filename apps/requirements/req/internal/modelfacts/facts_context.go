package modelfacts

import (
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_class"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_domain"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
)

type subdomainFactsContext struct {
	viewerSubdomainKey      identity.Key
	classByKey              map[string]model_class.Class
	domainNameByClassKey    map[string]string
	subdomainNameByClassKey map[string]string
}

func newSubdomainFactsContext(model core.Model, subdomain model_domain.Subdomain) subdomainFactsContext {
	ctx := subdomainFactsContext{
		viewerSubdomainKey:      subdomain.Key,
		classByKey:              make(map[string]model_class.Class),
		domainNameByClassKey:    make(map[string]string),
		subdomainNameByClassKey: make(map[string]string),
	}
	for _, domain := range model.Domains {
		for _, sd := range domain.Subdomains {
			for key, class := range sd.Classes {
				keyStr := key.String()
				ctx.classByKey[keyStr] = class
				ctx.domainNameByClassKey[keyStr] = domain.Name
				ctx.subdomainNameByClassKey[keyStr] = sd.Name
			}
		}
	}
	return ctx
}

func (ctx subdomainFactsContext) classDisplayName(class model_class.Class) string {
	name, err := model_class.FormatClassMarkdownDisplayName(
		ctx.viewerSubdomainKey,
		class,
		ctx.domainNameByClassKey[class.Key.String()],
		ctx.subdomainNameByClassKey[class.Key.String()],
	)
	if err != nil {
		return class.Name
	}
	return name
}

func (ctx subdomainFactsContext) classByKeyString(keyStr string) (model_class.Class, bool) {
	class, ok := ctx.classByKey[keyStr]
	return class, ok
}

func (ctx subdomainFactsContext) preferredAssociationLabelClass(assoc model_class.Association) (model_class.Class, bool) {
	subdomainStr := ctx.viewerSubdomainKey.String()
	candidates := []identity.Key{assoc.FromClassKey, assoc.ToClassKey}
	if assoc.AssociationClassKey != nil {
		candidates = append(candidates, *assoc.AssociationClassKey)
	}
	for _, classKey := range candidates {
		if classKey.ParentKey != subdomainStr {
			continue
		}
		class, ok := ctx.classByKey[classKey.String()]
		if ok {
			return class, true
		}
	}
	if class, ok := ctx.classByKey[assoc.FromClassKey.String()]; ok {
		return class, true
	}
	return model_class.Class{}, false
}

func associationTouchesSubdomain(assoc model_class.Association, subdomainKey identity.Key) bool {
	subdomainStr := subdomainKey.String()
	if assoc.FromClassKey.ParentKey == subdomainStr {
		return true
	}
	if assoc.ToClassKey.ParentKey == subdomainStr {
		return true
	}
	return assoc.AssociationClassKey != nil && assoc.AssociationClassKey.ParentKey == subdomainStr
}
