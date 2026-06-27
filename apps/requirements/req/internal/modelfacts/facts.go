package modelfacts

import (
	"fmt"
	"slices"
	"sort"
	"strings"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_class"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_domain"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_logic"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
)

// SubdomainPath is a domain folder and subdomain folder under a model root, e.g. "billing/ledger".
type SubdomainPath struct {
	DomainSubKey    string
	SubdomainSubKey string
}

// ParseSubdomainPath splits "domain/subdomain" from a filesystem path or flag value.
func ParseSubdomainPath(path string) (SubdomainPath, error) {
	path = strings.Trim(path, "/")
	parts := strings.Split(path, "/")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return SubdomainPath{}, fmt.Errorf("subdomain path must be domain/subdomain, got %q", path)
	}
	return SubdomainPath{DomainSubKey: parts[0], SubdomainSubKey: parts[1]}, nil
}

// FindSubdomain locates a subdomain in a parsed model by domain and subdomain folder names.
func FindSubdomain(model core.Model, path SubdomainPath) (model_domain.Subdomain, error) {
	for _, domain := range model.Domains {
		if domain.Key.SubKey != path.DomainSubKey {
			continue
		}
		for _, subdomain := range domain.Subdomains {
			if subdomain.Key.SubKey == path.SubdomainSubKey {
				return subdomain, nil
			}
		}
		return model_domain.Subdomain{}, fmt.Errorf("subdomain %q not found in domain %q", path.SubdomainSubKey, path.DomainSubKey)
	}
	return model_domain.Subdomain{}, fmt.Errorf("domain %q not found in model %q", path.DomainSubKey, model.Key)
}

// AssociationInvariantFact is one association invariant for facts rendering.
type AssociationInvariantFact struct {
	Label       string
	Description string
	Spec        string
}

// SubdomainFacts groups readable model facts for one subdomain.
type SubdomainFacts struct {
	Associations          []string
	AssociationInvariants []AssociationInvariantFact
	Indexes               []string
}

// FactsForSubdomain returns association, association invariant, and index uniqueness facts for one subdomain.
func FactsForSubdomain(subdomain model_domain.Subdomain) SubdomainFacts {
	return SubdomainFacts{
		Associations:          AssociationFactsForSubdomain(subdomain),
		AssociationInvariants: AssociationInvariantFactsForSubdomain(subdomain),
		Indexes:               IndexFactsForSubdomain(subdomain),
	}
}

// AssociationFactsForSubdomain returns human-readable association facts for associations whose
// from- and to-classes both belong to the subdomain.
func AssociationFactsForSubdomain(subdomain model_domain.Subdomain) []string {
	classByKey := make(map[string]model_class.Class, len(subdomain.Classes))
	for key, class := range subdomain.Classes {
		classByKey[key.String()] = class
	}

	var facts []string
	for _, assoc := range subdomain.ClassAssociations {
		if !associationWhollyInSubdomain(assoc, subdomain.Key) {
			continue
		}
		fromClass, okFrom := classByKey[assoc.FromClassKey.String()]
		toClass, okTo := classByKey[assoc.ToClassKey.String()]
		if !okFrom || !okTo {
			continue
		}
		var assocClass *model_class.Class
		if assoc.AssociationClassKey != nil {
			if ac, ok := classByKey[assoc.AssociationClassKey.String()]; ok {
				assocClass = &ac
			}
		}
		facts = append(facts, FormatAssociationFact(assoc, fromClass, toClass, assocClass))
	}

	sort.Strings(facts)
	return facts
}

// AssociationInvariantFactsForSubdomain returns human-readable facts for association-authored invariants.
func AssociationInvariantFactsForSubdomain(subdomain model_domain.Subdomain) []AssociationInvariantFact {
	classByKey := make(map[string]model_class.Class, len(subdomain.Classes))
	for key, class := range subdomain.Classes {
		classByKey[key.String()] = class
	}

	var facts []AssociationInvariantFact
	for _, assoc := range subdomain.ClassAssociations {
		if !associationWhollyInSubdomain(assoc, subdomain.Key) || len(assoc.Invariants) == 0 {
			continue
		}
		fromClass, okFrom := classByKey[assoc.FromClassKey.String()]
		if !okFrom {
			continue
		}
		for _, inv := range assoc.Invariants {
			facts = append(facts, FormatAssociationInvariantFact(assoc, fromClass, inv))
		}
	}

	for _, class := range subdomain.Classes {
		for _, inv := range class.Invariants {
			if inv.OverAssociationKey == nil {
				continue
			}
			assoc, ok := subdomain.ClassAssociations[*inv.OverAssociationKey]
			if !ok {
				continue
			}
			facts = append(facts, FormatAssociationInvariantFact(assoc, class, inv))
		}
	}

	sort.Slice(facts, func(i, j int) bool {
		return associationInvariantFactSortKey(facts[i]) < associationInvariantFactSortKey(facts[j])
	})
	return facts
}

// IndexFactsForSubdomain returns human-readable uniqueness facts for class attribute indexes.
func IndexFactsForSubdomain(subdomain model_domain.Subdomain) []string {
	classes := make([]model_class.Class, 0, len(subdomain.Classes))
	for _, class := range subdomain.Classes {
		classes = append(classes, class)
	}
	sort.Slice(classes, func(i, j int) bool {
		return classes[i].Name < classes[j].Name
	})

	var facts []string
	for _, class := range classes {
		indexMap := map[uint][]string{}
		for _, attr := range class.Attributes {
			for _, indexNum := range attr.IndexNums {
				indexMap[indexNum] = append(indexMap[indexNum], attr.Name)
			}
		}
		if len(indexMap) == 0 {
			continue
		}

		indexNums := make([]uint, 0, len(indexMap))
		for indexNum := range indexMap {
			indexNums = append(indexNums, indexNum)
		}
		slices.Sort(indexNums)

		for _, indexNum := range indexNums {
			names := indexMap[indexNum]
			sort.Strings(names)
			facts = append(facts, FormatIndexFact(class.Name, names))
		}
	}

	sort.Strings(facts)
	return facts
}

func associationWhollyInSubdomain(assoc model_class.Association, subdomainKey identity.Key) bool {
	subdomainStr := subdomainKey.String()
	if assoc.FromClassKey.ParentKey != subdomainStr || assoc.ToClassKey.ParentKey != subdomainStr {
		return false
	}
	if assoc.AssociationClassKey != nil && assoc.AssociationClassKey.ParentKey != subdomainStr {
		return false
	}
	return true
}

// FormatAssociationFact renders one class association as a review sentence.
//
// Multiplicity follows UML end notation as stored in the model: ToMultiplicity is how
// many to-class instances per one from-class instance; FromMultiplicity is how many
// from-class instances per one to-class instance.
func FormatAssociationFact(assoc model_class.Association, fromClass, toClass model_class.Class, associationClass *model_class.Class) string {
	fromPhrase := classPhrase(fromClass.Name)
	toPhrase := classPhrase(toClass.Name)

	forward := endConstraint(assoc.ToMultiplicity, fromPhrase, toPhrase, assoc.Name)
	inverse := endConstraint(assoc.FromMultiplicity, toPhrase, fromPhrase, "")

	var b strings.Builder
	b.WriteString(forward)
	if inverse != "" {
		b.WriteString("; ")
		b.WriteString(inverse)
	}
	if associationClass != nil {
		b.WriteString("; each ")
		b.WriteString(pairingPhrase(fromPhrase, toPhrase))
		b.WriteString(" is a ")
		b.WriteString(classPhrase(associationClass.Name).lower())
	}
	if details := strings.TrimSpace(assoc.Details); details != "" {
		b.WriteString(" (")
		b.WriteString(singleLine(details))
		b.WriteString(")")
	}
	b.WriteString(".")
	return b.String()
}

// FormatAssociationInvariantFact renders one association invariant for facts pages.
func FormatAssociationInvariantFact(assoc model_class.Association, fromClass model_class.Class, inv model_logic.Logic) AssociationInvariantFact {
	desc := singleLine(strings.TrimSpace(inv.Description))
	spec := logicSpecDisplay(inv)

	fact := AssociationInvariantFact{
		Label: fmt.Sprintf("%s (%s)", fromClass.Name, associationLabel(assoc.Name)),
	}

	switch {
	case desc != "":
		fact.Description = ensureSentence(desc)
		fact.Spec = spec
	case spec != "":
		fact.Description = ensureSentence(spec)
	default:
		fact.Description = "must satisfy an unspecified invariant."
	}

	return fact
}

func associationInvariantFactSortKey(fact AssociationInvariantFact) string {
	return fact.Label + ": " + fact.Description + " " + fact.Spec
}

func ensureSentence(text string) string {
	if text == "" {
		return text
	}
	if !strings.HasSuffix(text, ".") {
		return text + "."
	}
	return text
}

func logicSpecDisplay(inv model_logic.Logic) string {
	spec := singleLine(strings.TrimSpace(inv.Spec.Specification))
	if spec == "" {
		return ""
	}
	if inv.Target != "" {
		return "LET " + inv.Target + " = " + spec
	}
	return spec
}

// FormatIndexFact renders one class index as a review-friendly uniqueness sentence.
func FormatIndexFact(className string, attrNames []string) string {
	classPlural := classPhrase(className).plural()
	attrs := attributeListPhrase(attrNames)

	if len(attrNames) == 1 {
		return fmt.Sprintf("No %s can share the same %s.", classPlural, attrs)
	}
	return fmt.Sprintf("No %s can share the same %s combination.", classPlural, attrs)
}

func attributeListPhrase(names []string) string {
	switch len(names) {
	case 0:
		return ""
	case 1:
		return names[0]
	case 2:
		return names[0] + " and " + names[1]
	default:
		return strings.Join(names[:len(names)-1], ", ") + ", and " + names[len(names)-1]
	}
}

func endConstraint(m model_class.Multiplicity, subject, object classPhrase, assocName string) string {
	subjectLower := subject.lower()
	objectLower := object.lower()
	objectPlural := object.plural()

	switch {
	case m.LowerBound == 1 && m.HigherBound == 1:
		if assocName != "" {
			return fmt.Sprintf("each %s (%s) links to exactly one %s", subjectLower, associationLabel(assocName), objectLower)
		}
		return fmt.Sprintf("each %s links to exactly one %s", subjectLower, objectLower)

	case m.LowerBound == 0 && m.HigherBound == 1:
		if assocName != "" {
			return fmt.Sprintf("each %s (%s) may link to at most one %s", subjectLower, associationLabel(assocName), objectLower)
		}
		return fmt.Sprintf("each %s may link to at most one %s", subjectLower, objectLower)

	case m.LowerBound == 1 && m.HigherBound == 0:
		if assocName != "" {
			return fmt.Sprintf("each %s (%s) links to one or more %s", subjectLower, associationLabel(assocName), objectPlural)
		}
		return fmt.Sprintf("each %s links to one or more %s", subjectLower, objectPlural)

	case m.LowerBound == 0 && m.HigherBound == 0:
		if assocName != "" {
			return fmt.Sprintf("each %s (%s) links to any number of %s", subjectLower, associationLabel(assocName), objectPlural)
		}
		return fmt.Sprintf("each %s may link to any number of %s", subjectLower, objectPlural)

	default:
		label := ""
		if assocName != "" {
			label = fmt.Sprintf(" (%s)", associationLabel(assocName))
		}
		if m.LowerBound == m.HigherBound {
			return fmt.Sprintf("each %s%s links to exactly %d %s", subjectLower, label, m.LowerBound, objectPlural)
		}
		if m.HigherBound == 0 {
			return fmt.Sprintf("each %s%s links to %d or more %s", subjectLower, label, m.LowerBound, objectPlural)
		}
		return fmt.Sprintf("each %s%s links to between %d and %d %s", subjectLower, label, m.LowerBound, m.HigherBound, objectPlural)
	}
}

// associationLabel lowercases the association name from the model for use in fact prose.
func associationLabel(assocName string) string {
	return strings.ToLower(strings.TrimSpace(assocName))
}

type classPhrase string

func (c classPhrase) lower() string {
	return strings.ToLower(string(c))
}

func (c classPhrase) plural() string {
	lower := c.lower()
	if strings.HasSuffix(lower, "y") && len(lower) > 1 {
		return strings.TrimSuffix(lower, "y") + "ies"
	}
	if strings.HasSuffix(lower, "s") {
		return lower + "es"
	}
	return lower + "s"
}

func pairingPhrase(from, to classPhrase) string {
	return fmt.Sprintf("%s–%s pairing", from.lower(), to.lower())
}

func singleLine(s string) string {
	s = strings.ReplaceAll(s, "\n", " ")
	return strings.Join(strings.Fields(s), " ")
}
