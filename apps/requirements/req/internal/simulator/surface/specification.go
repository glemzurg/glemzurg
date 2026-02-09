package surface

import (
	"fmt"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model"
)

// SurfaceSpecification defines the scope of a simulation run.
// It uses an inclusion model: only explicitly included items participate.
// An empty specification means "simulate everything" (backward compatible).
type SurfaceSpecification struct {
	// IncludeDomains lists domain keys to include. All non-realized subdomains
	// and their classes within these domains are included.
	IncludeDomains []identity.Key

	// IncludeSubdomains lists subdomain keys to include.
	// More granular than IncludeDomains — includes only the named subdomains.
	IncludeSubdomains []identity.Key

	// IncludeClasses lists individual class keys to include.
	// Most granular level — includes only the named classes.
	IncludeClasses []identity.Key

	// ExcludeClasses lists class keys to exclude from the resolved set.
	// Applied AFTER includes. Useful for "domain X except class Y" patterns.
	ExcludeClasses []identity.Key
}

// IsEmpty returns true if no scope constraints are specified (simulate everything).
func (s *SurfaceSpecification) IsEmpty() bool {
	return len(s.IncludeDomains) == 0 &&
		len(s.IncludeSubdomains) == 0 &&
		len(s.IncludeClasses) == 0 &&
		len(s.ExcludeClasses) == 0
}

// Validate checks that all keys in the specification exist in the model.
func (s *SurfaceSpecification) Validate(model *req_model.Model) error {
	// Build lookup sets from the model.
	domainKeys := make(map[identity.Key]bool)
	subdomainKeys := make(map[identity.Key]bool)
	classKeys := make(map[identity.Key]bool)

	for domainKey, domain := range model.Domains {
		domainKeys[domainKey] = true
		for subdomainKey, subdomain := range domain.Subdomains {
			subdomainKeys[subdomainKey] = true
			for classKey := range subdomain.Classes {
				classKeys[classKey] = true
			}
		}
	}

	// Validate IncludeDomains.
	for _, dk := range s.IncludeDomains {
		if !domainKeys[dk] {
			return fmt.Errorf("IncludeDomains references unknown domain: %s", dk.String())
		}
	}

	// Validate IncludeSubdomains.
	for _, sk := range s.IncludeSubdomains {
		if !subdomainKeys[sk] {
			return fmt.Errorf("IncludeSubdomains references unknown subdomain: %s", sk.String())
		}
	}

	// Validate IncludeClasses.
	for _, ck := range s.IncludeClasses {
		if !classKeys[ck] {
			return fmt.Errorf("IncludeClasses references unknown class: %s", ck.String())
		}
	}

	// Validate ExcludeClasses.
	for _, ck := range s.ExcludeClasses {
		if !classKeys[ck] {
			return fmt.Errorf("ExcludeClasses references unknown class: %s", ck.String())
		}
	}

	return nil
}
