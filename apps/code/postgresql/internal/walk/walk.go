// Package walk pairs a requirements model with PostgreSQL annotation notes.
// Notes are matched by the relative path of the model file they describe.
package walk

import (
	"fmt"
	"sort"
	"strings"

	"github.com/glemzurg/glemzurg/apps/code/postgresql/internal/annot"
	"github.com/glemzurg/glemzurg/apps/requirements/req/core"
)

// Result is one walk of a model.
type Result struct {
	ModelName    string
	ModelKey     string
	Schema       string
	Include      *bool
	Associations []AssocVisit
	Domains      []DomainVisit
}

// DomainVisit is one domain and the children the walk binds.
type DomainVisit struct {
	Key          string
	Name         string
	Include      *bool
	Associations []AssocVisit
	Subdomains   []SubdomainVisit
}

// SubdomainVisit is one subdomain and the classes and associations under it.
type SubdomainVisit struct {
	Key          string
	Name         string
	Include      *bool
	Classes      []ClassVisit
	Associations []AssocVisit
}

// ClassVisit is one class. Noted lists attributes that have a note, in model order.
type ClassVisit struct {
	Key        string
	Name       string
	Attributes int
	Include    *bool
	Noted      []AttrVisit
}

// AttrVisit is the PostgreSQL note for one attribute.
type AttrVisit struct {
	Key     string
	Type    string
	Default string
	Include *bool
}

// AssocVisit is one class association.
type AssocVisit struct {
	Label    string
	OnDelete string
	Include  *bool
}

type context struct {
	notes   annot.Set
	visited map[string]bool
}

func newContext(notes annot.Set) *context {
	return &context{notes: notes, visited: map[string]bool{}}
}

// Walk visits the model and binds PostgreSQL notes for the model, domains,
// subdomains, classes, and class associations. A note whose path is not one of
// those files is reported as not applied.
func Walk(model core.Model, notes annot.Set) (Result, error) {
	ctx := newContext(notes)
	result := Result{ModelName: model.Name, ModelKey: model.Key}
	if err := applyModel(model, ctx, &result); err != nil {
		return Result{}, err
	}
	if err := applyDomains(model, ctx, &result); err != nil {
		return Result{}, err
	}
	if err := ctx.unbound(); err != nil {
		return Result{}, err
	}
	return result, nil
}

func (c *context) bind(rel string) (annot.Note, bool) {
	c.visited[rel] = true
	return c.notes.Note(rel)
}

func (c *context) unbound() error {
	var extra []string
	for _, rel := range c.notes.Paths() {
		if !c.visited[rel] {
			extra = append(extra, rel)
		}
	}
	if len(extra) == 0 {
		return nil
	}
	return fmt.Errorf("annotations not applied: %s", strings.Join(extra, ", "))
}

func applyModel(model core.Model, ctx *context, result *Result) error {
	if note, ok := ctx.bind("model.json"); ok {
		result.Schema = note.Schema
		result.Include = copyBool(note.Include)
	}
	assocs, err := modelAssociations(ctx, "model", "class_associations", "", model)
	if err != nil {
		return err
	}
	result.Associations = assocs
	return nil
}

func applyDomains(model core.Model, ctx *context, result *Result) error {
	for _, domainKey := range domainSubKeys(model) {
		if err := applyDomain(model, domainKey, ctx, result); err != nil {
			return err
		}
	}
	return nil
}

func domainSubKeys(model core.Model) []string {
	keys := make([]string, 0, len(model.Domains))
	for key := range model.Domains {
		keys = append(keys, key.SubKey)
	}
	sort.Strings(keys)
	return keys
}

func applyDomain(model core.Model, domainKey string, ctx *context, result *Result) error {
	visit, err := oneDomain(model, domainKey, ctx)
	if err != nil {
		return err
	}
	result.Domains = append(result.Domains, visit)
	return nil
}

func copyBool(value *bool) *bool {
	if value == nil {
		return nil
	}
	copied := *value
	return &copied
}

func unknownAttributeKeys(have []string, noted map[string]annot.Attribute) []string {
	known := make(map[string]struct{}, len(have))
	for _, key := range have {
		known[key] = struct{}{}
	}
	var missing []string
	for key := range noted {
		if _, ok := known[key]; !ok {
			missing = append(missing, key)
		}
	}
	sort.Strings(missing)
	return missing
}
