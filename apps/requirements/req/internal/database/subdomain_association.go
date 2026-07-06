package database

import (
	"fmt"
	"strings"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_domain"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"

	"github.com/pkg/errors"
)

func scanSubdomainAssociation(scanner Scanner, association *model_domain.SubdomainAssociation) (err error) {
	var keyStr string
	var problemSubdomainKeyStr string
	var solutionSubdomainKeyStr string

	if err = scanner.Scan(
		&keyStr,
		&problemSubdomainKeyStr,
		&solutionSubdomainKeyStr,
		&association.UmlComment,
	); err != nil {
		if err.Error() == _POSTGRES_NOT_FOUND {
			err = ErrNotFound
		}
		return err
	}

	association.Key, err = identity.ParseKey(keyStr)
	if err != nil {
		return err
	}
	association.ProblemSubdomainKey, err = identity.ParseKey(problemSubdomainKeyStr)
	if err != nil {
		return err
	}
	association.SolutionSubdomainKey, err = identity.ParseKey(solutionSubdomainKeyStr)
	if err != nil {
		return err
	}

	return nil
}

// LoadSubdomainAssociation loads a subdomain association from the database.
func LoadSubdomainAssociation(dbOrTx DbOrTx, modelKey string, associationKey identity.Key) (association model_domain.SubdomainAssociation, err error) {
	err = dbQueryRow(
		dbOrTx,
		func(scanner Scanner) (err error) {
			if err = scanSubdomainAssociation(scanner, &association); err != nil {
				return err
			}
			return nil
		},
		`SELECT
			association_key,
			problem_subdomain_key,
			solution_subdomain_key,
			uml_comment
		FROM
			subdomain_association
		WHERE
			association_key = $2
		AND
			model_key = $1`,
		modelKey,
		associationKey.String())
	if err != nil {
		return model_domain.SubdomainAssociation{}, errors.WithStack(err)
	}

	return association, nil
}

// AddSubdomainAssociation adds a subdomain association to the database.
func AddSubdomainAssociation(dbOrTx DbOrTx, modelKey string, association model_domain.SubdomainAssociation) (err error) {
	return AddSubdomainAssociations(dbOrTx, modelKey, []model_domain.SubdomainAssociation{association})
}

// UpdateSubdomainAssociation updates a subdomain association in the database.
func UpdateSubdomainAssociation(dbOrTx DbOrTx, modelKey string, association model_domain.SubdomainAssociation) (err error) {
	err = dbExec(dbOrTx, `
		UPDATE
			subdomain_association
		SET
			problem_subdomain_key  = $3,
			solution_subdomain_key = $4,
			uml_comment            = $5
		WHERE
			association_key = $2
		AND
			model_key = $1`,
		modelKey,
		association.Key.String(),
		association.ProblemSubdomainKey.String(),
		association.SolutionSubdomainKey.String(),
		association.UmlComment)
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

// RemoveSubdomainAssociation deletes a subdomain association from the database.
func RemoveSubdomainAssociation(dbOrTx DbOrTx, modelKey string, associationKey identity.Key) (err error) {
	err = dbExec(dbOrTx, `
		DELETE FROM
			subdomain_association
		WHERE
			association_key = $2
		AND
			model_key = $1`,
		modelKey,
		associationKey.String())
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

// QuerySubdomainAssociations loads all subdomain associations for a model.
func QuerySubdomainAssociations(dbOrTx DbOrTx, modelKey string) (associations []model_domain.SubdomainAssociation, err error) {
	err = dbQuery(
		dbOrTx,
		func(scanner Scanner) (err error) {
			var association model_domain.SubdomainAssociation
			if err = scanSubdomainAssociation(scanner, &association); err != nil {
				return errors.WithStack(err)
			}
			associations = append(associations, association)
			return nil
		},
		`SELECT
			association_key,
			problem_subdomain_key,
			solution_subdomain_key,
			uml_comment
		FROM
			subdomain_association
		WHERE
			model_key = $1
		ORDER BY association_key`,
		modelKey)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return associations, nil
}

// AddSubdomainAssociations adds multiple subdomain associations to the database in a single insert.
func AddSubdomainAssociations(dbOrTx DbOrTx, modelKey string, associations []model_domain.SubdomainAssociation) (err error) {
	if len(associations) == 0 {
		return nil
	}

	var queryBuilder strings.Builder
	queryBuilder.WriteString(`INSERT INTO subdomain_association (model_key, association_key, problem_subdomain_key, solution_subdomain_key, uml_comment) VALUES `)
	args := make([]any, 0, len(associations)*5)
	for i, assoc := range associations {
		if i > 0 {
			queryBuilder.WriteString(", ")
		}
		base := i * 5
		fmt.Fprintf(&queryBuilder, "($%d, $%d, $%d, $%d, $%d)", base+1, base+2, base+3, base+4, base+5)
		args = append(args, modelKey, assoc.Key.String(), assoc.ProblemSubdomainKey.String(), assoc.SolutionSubdomainKey.String(), assoc.UmlComment)
	}

	err = dbExec(dbOrTx, queryBuilder.String(), args...)
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}
