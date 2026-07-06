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
func FactsForSubdomain(model core.Model, subdomain model_domain.Subdomain) SubdomainFacts {
	return SubdomainFacts{
		Associations:          AssociationFactsForSubdomain(model, subdomain),
		AssociationInvariants: AssociationInvariantFactsForSubdomain(model, subdomain),
		Indexes:               IndexFactsForSubdomain(model, subdomain),
	}
}

// AssociationFactsForSubdomain returns human-readable association facts for associations that
// touch the subdomain via either end or an association class in that subdomain.
func AssociationFactsForSubdomain(model core.Model, subdomain model_domain.Subdomain) []string {
	ctx := newSubdomainFactsContext(model, subdomain)

	var facts []string
	for _, assoc := range model.GetClassAssociations() {
		if !associationTouchesSubdomain(assoc, subdomain.Key) {
			continue
		}
		fromClass, okFrom := ctx.classByKeyString(assoc.FromClassKey.String())
		toClass, okTo := ctx.classByKeyString(assoc.ToClassKey.String())
		if !okFrom || !okTo {
			continue
		}
		var assocClass *model_class.Class
		assocClassDisplay := ""
		if assoc.AssociationClassKey != nil {
			if ac, ok := ctx.classByKeyString(assoc.AssociationClassKey.String()); ok {
				assocClass = &ac
				assocClassDisplay = ctx.classDisplayName(ac)
			}
		}
		facts = append(facts, FormatAssociationFact(assoc, associationFactEnds{
			fromClass:               fromClass,
			toClass:                 toClass,
			associationClass:        assocClass,
			fromDisplay:             ctx.classDisplayName(fromClass),
			toDisplay:               ctx.classDisplayName(toClass),
			associationClassDisplay: assocClassDisplay,
		}))
	}

	sort.Strings(facts)
	return facts
}

// AssociationInvariantFactsForSubdomain returns human-readable facts for association-authored invariants.
func AssociationInvariantFactsForSubdomain(model core.Model, subdomain model_domain.Subdomain) []AssociationInvariantFact {
	ctx := newSubdomainFactsContext(model, subdomain)
	allAssociations := model.GetClassAssociations()

	var facts []AssociationInvariantFact
	for _, assoc := range allAssociations {
		if !associationTouchesSubdomain(assoc, subdomain.Key) || len(assoc.Invariants) == 0 {
			continue
		}
		labelClass, ok := ctx.preferredAssociationLabelClass(assoc)
		if !ok {
			continue
		}
		label := ctx.classDisplayName(labelClass)
		for _, inv := range assoc.Invariants {
			facts = append(facts, FormatAssociationInvariantFact(assoc, inv, label))
		}
	}

	for _, class := range subdomain.Classes {
		for _, inv := range class.Invariants {
			if inv.OverAssociationKey == nil {
				continue
			}
			assoc, ok := allAssociations[*inv.OverAssociationKey]
			if !ok || !associationTouchesSubdomain(assoc, subdomain.Key) {
				continue
			}
			facts = append(facts, FormatAssociationInvariantFact(assoc, inv, ctx.classDisplayName(class)))
		}
	}

	sort.Slice(facts, func(i, j int) bool {
		return associationInvariantFactSortKey(facts[i]) < associationInvariantFactSortKey(facts[j])
	})
	return facts
}

// IndexFactsForSubdomain returns human-readable uniqueness facts for class attribute indexes.
func IndexFactsForSubdomain(model core.Model, subdomain model_domain.Subdomain) []string {
	ctx := newSubdomainFactsContext(model, subdomain)
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
			facts = append(facts, FormatIndexFact(ctx.classDisplayName(class), names))
		}
	}

	sort.Strings(facts)
	return facts
}

// associationFactEnds groups association endpoint classes and their scoped display names.
type associationFactEnds struct {
	fromClass, toClass      model_class.Class
	associationClass        *model_class.Class
	fromDisplay, toDisplay  string
	associationClassDisplay string
}

// FormatAssociationFact renders one class association as a review sentence.
//
// Multiplicity follows UML end notation as stored in the model: ToMultiplicity is how
// many to-class instances per one from-class instance; FromMultiplicity is how many
// from-class instances per one to-class instance.
func FormatAssociationFact(assoc model_class.Association, ends associationFactEnds) string {
	fromPhrase := classPhrase(ends.fromDisplay)
	toPhrase := classPhrase(ends.toDisplay)

	forward := endConstraint(assoc.ToMultiplicity, fromPhrase, toPhrase, assoc.Name)
	inverse := endConstraint(assoc.FromMultiplicity, toPhrase, fromPhrase, "")

	var b strings.Builder
	b.WriteString(forward)
	if inverse != "" {
		b.WriteString("; ")
		b.WriteString(inverse)
	}
	if ends.associationClass != nil {
		b.WriteString("; each ")
		b.WriteString(pairingPhrase(fromPhrase, toPhrase))
		b.WriteString(" is a ")
		b.WriteString(classPhrase(ends.associationClassDisplay).lower())
	}
	if uniq := formatAssociationUniquenessDisplay(assoc.Uniqueness, ends.fromClass, ends.toClass); uniq != "" {
		b.WriteString("; each ")
		b.WriteString(pairingPhrase(fromPhrase, toPhrase))
		b.WriteString(" has the uniqueness ")
		b.WriteString(uniq)
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
func FormatAssociationInvariantFact(assoc model_class.Association, inv model_logic.Logic, fromDisplay string) AssociationInvariantFact {
	desc := singleLine(strings.TrimSpace(inv.Description))
	spec := logicSpecDisplay(inv)

	fact := AssociationInvariantFact{
		Label: fmt.Sprintf("%s (%s)", fromDisplay, associationLabel(assoc.Name)),
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

// formatAssociationUniquenessDisplay renders the uniqueness tuple as endpoint attribute
// names separated by →. Blank sides stay empty when that endpoint lists no attributes.
func formatAssociationUniquenessDisplay(
	uniqueness *model_class.AssociationUniqueness,
	fromClass, toClass model_class.Class,
) string {
	if uniqueness == nil {
		return ""
	}
	fromAttrs := attributeListPhrase(attributeNamesFromClass(fromClass, uniqueness.FromAttributeKeys))
	toAttrs := attributeListPhrase(attributeNamesFromClass(toClass, uniqueness.ToAttributeKeys))
	switch {
	case fromAttrs == "" && toAttrs == "":
		return ""
	case fromAttrs == "":
		return "→ " + toAttrs
	case toAttrs == "":
		return fromAttrs + " →"
	default:
		return fromAttrs + " → " + toAttrs
	}
}

func attributeNamesFromClass(class model_class.Class, keys []identity.Key) []string {
	if len(keys) == 0 {
		return nil
	}
	names := make([]string, 0, len(keys))
	for _, key := range keys {
		names = append(names, attributeNameFromClass(class, key))
	}
	return names
}

func attributeNameFromClass(class model_class.Class, attrKey identity.Key) string {
	for _, attr := range class.Attributes {
		if attr.Key == attrKey {
			return attr.Name
		}
	}
	return attrKey.SubKey
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

// lower returns the class display name for fact prose. Names come from the model (or scoped
// markdown display), not from class keys.
func (c classPhrase) lower() string {
	return string(c)
}

func (c classPhrase) plural() string {
	s := string(c)
	if strings.Contains(s, "::") {
		parts := strings.Split(s, "::")
		parts[len(parts)-1] = pluralizeDisplayPhrase(parts[len(parts)-1])
		return strings.Join(parts, "::")
	}
	return pluralizeDisplayPhrase(s)
}

func pluralizeWord(word string) string {
	lower := strings.ToLower(word)
	switch {
	case strings.HasSuffix(lower, "y") && len(lower) > 1:
		return strings.TrimSuffix(lower, "y") + "ies"
	case strings.HasSuffix(lower, "s"):
		return lower + "es"
	default:
		return lower + "s"
	}
}

func pluralizeDisplayWord(word string) string {
	plural := pluralizeWord(word)
	if len(word) > 0 && word[0] >= 'A' && word[0] <= 'Z' {
		return strings.ToUpper(plural[:1]) + plural[1:]
	}
	return plural
}

func pluralizeDisplayPhrase(phrase string) string {
	words := strings.Fields(phrase)
	if len(words) == 0 {
		return phrase
	}
	words[len(words)-1] = pluralizeDisplayWord(words[len(words)-1])
	return strings.Join(words, " ")
}

func pairingPhrase(from, to classPhrase) string {
	return fmt.Sprintf("%s–%s pairing", from.lower(), to.lower())
}

func singleLine(s string) string {
	s = strings.ReplaceAll(s, "\n", " ")
	return strings.Join(strings.Fields(s), " ")
}
