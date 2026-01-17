package database

import (
	"fmt"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_domain"

	"github.com/pkg/errors"
)

// Populate a golang struct from a database row.
func scanDomainAssociation(scanner Scanner, association *model_domain.Association) (err error) {
	var keyStr string
	var problemDomainKeyStr string
	var solutionDomainKeyStr string

	if err = scanner.Scan(
		&keyStr,
		&problemDomainKeyStr,
		&solutionDomainKeyStr,
		&association.UmlComment,
	); err != nil {
		if err.Error() == _POSTGRES_NOT_FOUND {
			err = ErrNotFound
		}
		return err // Do not wrap in stack here. It will be wrapped in the database calls.
	}

	// Parse the key strings into identity.Key.
	association.Key, err = identity.ParseKey(keyStr)
	if err != nil {
		return err
	}
	association.ProblemDomainKey, err = identity.ParseKey(problemDomainKeyStr)
	if err != nil {
		return err
	}
	association.SolutionDomainKey, err = identity.ParseKey(solutionDomainKeyStr)
	if err != nil {
		return err
	}

	return nil
}

// LoadDomainAssociation loads a association from the database
func LoadDomainAssociation(dbOrTx DbOrTx, modelKey string, associationKey identity.Key) (association model_domain.Association, err error) {

	// Query the database.
	err = dbQueryRow(
		dbOrTx,
		func(scanner Scanner) (err error) {
			if err = scanDomainAssociation(scanner, &association); err != nil {
				return err
			}
			return nil
		},
		`SELECT
			association_key     ,
			problem_domain_key  ,
			solution_domain_key ,
			uml_comment
		FROM
			domain_association
		WHERE
			association_key = $2
		AND
			model_key = $1
		ORDER BY association_key`,
		modelKey,
		associationKey.String())
	if err != nil {
		return model_domain.Association{}, errors.WithStack(err)
	}

	return association, nil
}

// AddDomainAssociation adds a association to the database.
func AddDomainAssociation(dbOrTx DbOrTx, modelKey string, association model_domain.Association) (err error) {
	return AddDomainAssociations(dbOrTx, modelKey, []model_domain.Association{association})
}

// UpdateDomainAssociation updates a association in the database.
func UpdateDomainAssociation(dbOrTx DbOrTx, modelKey string, association model_domain.Association) (err error) {

	// Update the data.
	_, err = dbExec(dbOrTx, `
		UPDATE
			domain_association
		SET
			problem_domain_key  = $3 ,
			solution_domain_key = $4 ,
			uml_comment         = $5
		WHERE
			association_key = $2
		AND
			model_key = $1`,
		modelKey,
		association.Key.String(),
		association.ProblemDomainKey.String(),
		association.SolutionDomainKey.String(),
		association.UmlComment)
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

// RemoveDomainAssociation deletes a association from the database.
func RemoveDomainAssociation(dbOrTx DbOrTx, modelKey string, associationKey identity.Key) (err error) {

	// Delete the data.
	_, err = dbExec(dbOrTx, `
		DELETE FROM
			domain_association
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

// QueryDomainAssociations loads all association from the database
func QueryDomainAssociations(dbOrTx DbOrTx, modelKey string) (associations []model_domain.Association, err error) {

	// Query the database.
	err = dbQuery(
		dbOrTx,
		func(scanner Scanner) (err error) {
			var association model_domain.Association
			if err = scanDomainAssociation(scanner, &association); err != nil {
				return errors.WithStack(err)
			}
			associations = append(associations, association)
			return nil
		},
		`SELECT
			association_key     ,
			problem_domain_key  ,
			solution_domain_key ,
			uml_comment
		FROM
			domain_association
		WHERE
			model_key = $1
		ORDER BY association_key`,
		modelKey)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return associations, nil
}

// AddDomainAssociations adds multiple domain associations to the database in a single insert.
func AddDomainAssociations(dbOrTx DbOrTx, modelKey string, associations []model_domain.Association) (err error) {
	if len(associations) == 0 {
		return nil
	}

	// Build the bulk insert query.
	query := `INSERT INTO domain_association (model_key, association_key, problem_domain_key, solution_domain_key, uml_comment) VALUES `
	args := make([]interface{}, 0, len(associations)*5)
	for i, assoc := range associations {
		if i > 0 {
			query += ", "
		}
		base := i * 5
		query += fmt.Sprintf("($%d, $%d, $%d, $%d, $%d)", base+1, base+2, base+3, base+4, base+5)
		args = append(args, modelKey, assoc.Key.String(), assoc.ProblemDomainKey.String(), assoc.SolutionDomainKey.String(), assoc.UmlComment)
	}

	_, err = dbExec(dbOrTx, query, args...)
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}
