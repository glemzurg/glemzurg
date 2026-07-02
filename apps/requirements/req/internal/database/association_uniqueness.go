package database

import (
	"fmt"
	"strings"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_class"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/pkg/errors"
)

type associationUniquenessRow struct {
	AssociationKey    identity.Key
	FromAttributeKeys []identity.Key
	ToAttributeKeys   []identity.Key
}

// AddAssociationUniqueness inserts uniqueness attribute tuples for associations.
func AddAssociationUniqueness(dbOrTx DbOrTx, modelKey string, associations []model_class.Association) error {
	var attrBuilder strings.Builder
	attrArgs := make([]any, 0)
	attrCount := 0

	for _, assoc := range associations {
		if assoc.Uniqueness == nil {
			continue
		}
		for j, attrKey := range assoc.Uniqueness.FromAttributeKeys {
			if attrCount > 0 {
				attrBuilder.WriteString(", ")
			}
			if attrCount == 0 {
				attrBuilder.WriteString(`INSERT INTO association_uniqueness_attribute (model_key, association_key, end_side, attribute_sort_order, attribute_key) VALUES `)
			}
			base := attrCount * 5
			fmt.Fprintf(&attrBuilder, "($%d, $%d, $%d::association_end, $%d, $%d)", base+1, base+2, base+3, base+4, base+5)
			attrArgs = append(attrArgs, modelKey, assoc.Key.String(), associationEndFrom, j, attrKey.String())
			attrCount++
		}
		for j, attrKey := range assoc.Uniqueness.ToAttributeKeys {
			if attrCount > 0 {
				attrBuilder.WriteString(", ")
			}
			if attrCount == 0 {
				attrBuilder.WriteString(`INSERT INTO association_uniqueness_attribute (model_key, association_key, end_side, attribute_sort_order, attribute_key) VALUES `)
			}
			base := attrCount * 5
			fmt.Fprintf(&attrBuilder, "($%d, $%d, $%d::association_end, $%d, $%d)", base+1, base+2, base+3, base+4, base+5)
			attrArgs = append(attrArgs, modelKey, assoc.Key.String(), associationEndTo, j, attrKey.String())
			attrCount++
		}
	}
	if attrCount == 0 {
		return nil
	}
	return errors.WithStack(dbExec(dbOrTx, attrBuilder.String(), attrArgs...))
}

// QueryAssociationUniqueness loads uniqueness keyed by association.
func QueryAssociationUniqueness(dbOrTx DbOrTx, modelKey string) (map[identity.Key]*model_class.AssociationUniqueness, error) {
	rowsByAssoc := make(map[identity.Key]*associationUniquenessRow)

	err := dbQuery(dbOrTx, func(scanner Scanner) error {
		var associationKeyStr, endSide, attributeKeyStr string
		var attributeSortOrder int
		if err := scanner.Scan(&associationKeyStr, &endSide, &attributeSortOrder, &attributeKeyStr); err != nil {
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
		row := rowsByAssoc[associationKey]
		if row == nil {
			row = &associationUniquenessRow{AssociationKey: associationKey}
			rowsByAssoc[associationKey] = row
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
	}, `SELECT association_key, end_side, attribute_sort_order, attribute_key
		FROM association_uniqueness_attribute
		WHERE model_key = $1
		ORDER BY association_key, end_side, attribute_sort_order`, modelKey)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	result := make(map[identity.Key]*model_class.AssociationUniqueness, len(rowsByAssoc))
	for assocKey, row := range rowsByAssoc {
		uniqueness := model_class.NewAssociationUniqueness(row.FromAttributeKeys, row.ToAttributeKeys)
		result[assocKey] = &uniqueness
	}
	return result, nil
}
