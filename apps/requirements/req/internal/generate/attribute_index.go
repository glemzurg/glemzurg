package generate

import (
	"fmt"
	"slices"
	"sort"
	"strings"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_class"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_logic"
)

const attributeIndexKeyLabel = "key"

// ClassIndexListing groups attribute names that share one index number.
type ClassIndexListing struct {
	Heading    string
	Attributes []string
}

// attributeIndexBracketLabel names an index inside attribute bracket suffixes like [key,i3].
func attributeIndexBracketLabel(indexNum uint) string {
	if indexNum == 0 {
		return attributeIndexKeyLabel
	}
	return fmt.Sprintf("i%d", indexNum)
}

func classIndexListingHeading(indexNum uint) string {
	if indexNum == 0 {
		return attributeIndexKeyLabel
	}
	return fmt.Sprintf("index %d", indexNum)
}

func attributeIndexBracketSuffix(indexNums []uint) string {
	if len(indexNums) == 0 {
		return ""
	}
	sorted := slices.Clone(indexNums)
	slices.Sort(sorted)

	labels := make([]string, 0, len(sorted))
	for _, indexNum := range sorted {
		labels = append(labels, attributeIndexBracketLabel(indexNum))
	}
	return " [" + strings.Join(labels, ",") + "]"
}

// attributeCommentsInvariants renders attribute details and invariants for the attributes table.
func attributeCommentsInvariants(attr model_class.Attribute) string {
	var parts []string
	if details := strings.TrimSpace(attr.Details); details != "" {
		parts = append(parts, details)
	}
	if len(attr.Invariants) > 0 {
		if len(parts) > 0 {
			parts = append(parts, "")
		}
		parts = append(parts, logicListMarkdownHTML(attr.Invariants)...)
	}
	return strings.Join(parts, "<br>")
}

func logicListMarkdownHTML(logics []model_logic.Logic) []string {
	parts := make([]string, 0, len(logics)*3)
	for i, logic := range logics {
		if i > 0 {
			parts = append(parts, "")
		}
		parts = append(parts, logicInvariantMarkdownHTML(logic)...)
	}
	return parts
}

func logicInvariantMarkdownHTML(logic model_logic.Logic) []string {
	var parts []string
	if desc := strings.TrimSpace(logic.Description); desc != "" {
		parts = append(parts, desc)
	}
	if specLine := logicBoldSpecText(logic); specLine != "" {
		parts = append(parts, specLine)
	}
	if logic.Target != "" && logic.TargetTypeSpec != nil {
		if typeSpec := strings.TrimSpace(logic.TargetTypeSpec.Specification); typeSpec != "" {
			parts = append(parts, "Type: "+typeSpec)
		}
	}
	return parts
}

// classAttributeTableName renders the attribute name column, including derivation prefix
// and index membership suffix matching the class UML diagram.

func classAttributeTableName(attr model_class.Attribute) string {
	var name strings.Builder
	if attr.DerivationPolicy != nil {
		name.WriteString("/")
	}
	name.WriteString(attr.Name)
	name.WriteString(attributeIndexBracketSuffix(attr.IndexNums))
	return name.String()
}

func classIndexListings(attributes []model_class.Attribute) []ClassIndexListing {
	indexMap := map[uint][]string{}
	for _, attr := range attributes {
		for _, indexNum := range attr.IndexNums {
			indexMap[indexNum] = append(indexMap[indexNum], attr.Name)
		}
	}
	if len(indexMap) == 0 {
		return nil
	}

	indexNums := make([]uint, 0, len(indexMap))
	for indexNum := range indexMap {
		indexNums = append(indexNums, indexNum)
	}
	slices.Sort(indexNums)

	listings := make([]ClassIndexListing, 0, len(indexNums))
	for _, indexNum := range indexNums {
		names := indexMap[indexNum]
		sort.Strings(names)
		listings = append(listings, ClassIndexListing{
			Heading:    classIndexListingHeading(indexNum),
			Attributes: names,
		})
	}
	return listings
}
