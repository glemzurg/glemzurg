package generate

import (
	"fmt"
	"slices"
	"sort"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_class"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
)

const attributeIndexKeyLabel = "key"

// ClassIndexListing groups attribute names that share one index number.
type ClassIndexListing struct {
	Name       string
	Attributes []string
}

// attributeIndexLabel names an index for markdown and Mermaid display.
// Index 0 is "key"; index N for N >= 1 is "iN".
func attributeIndexLabel(indexNum uint) string {
	if indexNum == 0 {
		return attributeIndexKeyLabel
	}
	return fmt.Sprintf("i%d", indexNum)
}

func classIndexListings(attributes map[identity.Key]model_class.Attribute) []ClassIndexListing {
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
			Name:       attributeIndexLabel(indexNum),
			Attributes: names,
		})
	}
	return listings
}
