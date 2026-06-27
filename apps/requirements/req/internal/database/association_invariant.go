package database

import (
	"fmt"
	"strings"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/pkg/errors"
)

// AssociationInvariantLink is one association_invariant row with optional to-class anchoring.
type AssociationInvariantLink struct {
	AssociationKey identity.Key
	LogicKey       identity.Key
	// ToClassAnchor is true when the invariant is evaluated on the association to-class (reverse anchor).
	ToClassAnchor bool
}

// QueryAssociationInvariantLinks loads all association invariant join rows.
func QueryAssociationInvariantLinks(dbOrTx DbOrTx, modelKey string) ([]AssociationInvariantLink, error) {
	var links []AssociationInvariantLink

	err := dbQuery(
		dbOrTx,
		func(scanner Scanner) error {
			var associationKeyStr, logicKeyStr string
			var toClassAnchor bool
			if err := scanner.Scan(&associationKeyStr, &logicKeyStr, &toClassAnchor); err != nil {
				return errors.WithStack(err)
			}
			associationKey, err := identity.ParseKey(associationKeyStr)
			if err != nil {
				return errors.WithStack(err)
			}
			logicKey, err := identity.ParseKey(logicKeyStr)
			if err != nil {
				return errors.WithStack(err)
			}
			links = append(links, AssociationInvariantLink{
				AssociationKey: associationKey,
				LogicKey:       logicKey,
				ToClassAnchor:  toClassAnchor,
			})
			return nil
		},
		`SELECT
			ai.association_key,
			ai.logic_key,
			ai.to_class_anchor
		FROM
			association_invariant ai
		JOIN
			logic l ON l.model_key = ai.model_key AND l.logic_key = ai.logic_key
		WHERE
			ai.model_key = $1
		ORDER BY ai.association_key, l.sort_order, ai.logic_key`,
		modelKey,
	)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return links, nil
}

// FromAnchoredInvariantKeysByAssociation groups from-class-anchored invariant logic keys by association.
func FromAnchoredInvariantKeysByAssociation(links []AssociationInvariantLink) map[identity.Key][]identity.Key {
	result := make(map[identity.Key][]identity.Key)
	for _, link := range links {
		if link.ToClassAnchor {
			continue
		}
		result[link.AssociationKey] = append(result[link.AssociationKey], link.LogicKey)
	}
	return result
}

// ToAnchoredAssociationKeyByLogic maps to-class-anchored invariant logic keys to their association keys.
func ToAnchoredAssociationKeyByLogic(links []AssociationInvariantLink) map[identity.Key]identity.Key {
	result := make(map[identity.Key]identity.Key)
	for _, link := range links {
		if !link.ToClassAnchor {
			continue
		}
		result[link.LogicKey] = link.AssociationKey
	}
	return result
}

// AddAssociationInvariantLinks inserts association invariant join rows.
func AddAssociationInvariantLinks(dbOrTx DbOrTx, modelKey string, links []AssociationInvariantLink) error {
	if len(links) == 0 {
		return nil
	}

	var qb strings.Builder
	qb.WriteString(`INSERT INTO association_invariant (model_key, association_key, logic_key, to_class_anchor) VALUES `)
	args := make([]any, 0, len(links)*4)
	for i, link := range links {
		if i > 0 {
			qb.WriteString(", ")
		}
		base := i * 4
		fmt.Fprintf(&qb, "($%d, $%d, $%d, $%d)", base+1, base+2, base+3, base+4)
		args = append(args,
			modelKey,
			link.AssociationKey.String(),
			link.LogicKey.String(),
			link.ToClassAnchor,
		)
	}

	if err := dbExec(dbOrTx, qb.String(), args...); err != nil {
		return errors.WithStack(err)
	}
	return nil
}
