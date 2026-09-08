package database

import (
	"fmt"
	"slices"
	"strings"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_class"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/pkg/errors"
)

type associationUniquenessRow struct {
	FromAttributeKeys []identity.Key
	ToAttributeKeys   []identity.Key
}

// AddAssociationUniqueness inserts uniqueness attribute tuples for associations.
func AddAssociationUniqueness(dbOrTx DbOrTx, modelKey string, associations []model_class.Association) error {
	var attrBuilder strings.Builder
	attrArgs := make([]any, 0)
	attrCount := 0

	for _, assoc := range associations {
		for uniquenessSortOrder, uniqueness := range assoc.Uniqueness {
			for j, attrKey := range uniqueness.FromAttributeKeys {
				if attrCount > 0 {
					attrBuilder.WriteString(", ")
				}
				if attrCount == 0 {
					attrBuilder.WriteString(`INSERT INTO association_uniqueness_attribute (model_key, association_key, uniqueness_sort_order, end_side, attribute_sort_order, attribute_key) VALUES `)
				}
				base := attrCount * 6
				fmt.Fprintf(&attrBuilder, "($%d, $%d, $%d, $%d::association_end, $%d, $%d)", base+1, base+2, base+3, base+4, base+5, base+6)
				attrArgs = append(attrArgs, modelKey, assoc.Key.String(), uniquenessSortOrder, associationEndFrom, j, attrKey.String())
				attrCount++
			}
			for j, attrKey := range uniqueness.ToAttributeKeys {
				if attrCount > 0 {
					attrBuilder.WriteString(", ")
				}
				if attrCount == 0 {
					attrBuilder.WriteString(`INSERT INTO association_uniqueness_attribute (model_key, association_key, uniqueness_sort_order, end_side, attribute_sort_order, attribute_key) VALUES `)
				}
				base := attrCount * 6
				fmt.Fprintf(&attrBuilder, "($%d, $%d, $%d, $%d::association_end, $%d, $%d)", base+1, base+2, base+3, base+4, base+5, base+6)
				attrArgs = append(attrArgs, modelKey, assoc.Key.String(), uniquenessSortOrder, associationEndTo, j, attrKey.String())
				attrCount++
			}
		}
	}
	if attrCount == 0 {
		return nil
	}
	return errors.WithStack(dbExec(dbOrTx, attrBuilder.String(), attrArgs...))
}

// QueryAssociationUniqueness loads uniqueness keyed by association.
func QueryAssociationUniqueness(dbOrTx DbOrTx, modelKey string) (map[identity.Key][]model_class.AssociationUniqueness, error) {
	rowsByAssoc := make(map[identity.Key]map[int]*associationUniquenessRow)

	err := dbQuery(dbOrTx, func(scanner Scanner) error {
		var associationKeyStr, endSide, attributeKeyStr string
		var uniquenessSortOrder, attributeSortOrder int
		if err := scanner.Scan(&associationKeyStr, &uniquenessSortOrder, &endSide, &attributeSortOrder, &attributeKeyStr); err != nil {
			return errors.WithStack(err)
		}
		associationKey, err := identity.ParseKey(associationKeyStr)
		if err != nil {
			return err
		}
		attributeKey, err := identity.ParseKey(attributeKeyStr)
		if err != nil {
			return err
		}
		bySort := rowsByAssoc[associationKey]
		if bySort == nil {
			bySort = make(map[int]*associationUniquenessRow)
			rowsByAssoc[associationKey] = bySort
		}
		row := bySort[uniquenessSortOrder]
		if row == nil {
			row = &associationUniquenessRow{}
			bySort[uniquenessSortOrder] = row
		}
		switch associationEnd(endSide) {
		case associationEndFrom:
			row.FromAttributeKeys = append(row.FromAttributeKeys, attributeKey)
		case associationEndTo:
			row.ToAttributeKeys = append(row.ToAttributeKeys, attributeKey)
		default:
			return errors.Errorf("invalid association_end %q", endSide)
		}
		return nil
	}, `SELECT association_key, uniqueness_sort_order, end_side, attribute_sort_order, attribute_key
		FROM association_uniqueness_attribute
		WHERE model_key = $1
		ORDER BY association_key, uniqueness_sort_order, end_side, attribute_sort_order`, modelKey)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	result := make(map[identity.Key][]model_class.AssociationUniqueness, len(rowsByAssoc))
	for assocKey, bySort := range rowsByAssoc {
		sorts := make([]int, 0, len(bySort))
		for uniquenessSortOrder := range bySort {
			sorts = append(sorts, uniquenessSortOrder)
		}
		slices.Sort(sorts)
		constraints := make([]model_class.AssociationUniqueness, 0, len(sorts))
		for _, uniquenessSortOrder := range sorts {
			row := bySort[uniquenessSortOrder]
			constraints = append(constraints, model_class.NewAssociationUniqueness(row.FromAttributeKeys, row.ToAttributeKeys))
		}
		result[assocKey] = constraints
	}
	return result, nil
}
