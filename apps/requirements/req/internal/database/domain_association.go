package database

import (
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/requirements"

	"github.com/pkg/errors"
)

// Populate a golang struct from a database row.
func scanDomainAssociation(scanner Scanner, association *requirements.DomainAssociation) (err error) {
	if err = scanner.Scan(
		&association.Key,
		&association.ProblemDomainKey,
		&association.SolutionDomainKey,
		&association.UmlComment,
	); err != nil {
		if err.Error() == _POSTGRES_NOT_FOUND {
			err = ErrNotFound
		}
		return err // Do not wrap in stack here. It will be wrapped in the database calls.
	}

	return nil
}

// LoadDomainAssociation loads a association from the database
func LoadDomainAssociation(dbOrTx DbOrTx, modelKey, associationKey string) (association requirements.DomainAssociation, err error) {

	// Keys should be preened so they collide correctly.
	modelKey, err = requirements.PreenKey(modelKey)
	if err != nil {
		return requirements.DomainAssociation{}, err
	}
	associationKey, err = requirements.PreenKey(associationKey)
	if err != nil {
		return requirements.DomainAssociation{}, err
	}

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
		associationKey)
	if err != nil {
		return requirements.DomainAssociation{}, errors.WithStack(err)
	}

	return association, nil
}

// AddDomainAssociation adds a association to the database.
func AddDomainAssociation(dbOrTx DbOrTx, modelKey string, association requirements.DomainAssociation) (err error) {

	// Keys should be preened so they collide correctly.
	modelKey, err = requirements.PreenKey(modelKey)
	if err != nil {
		return err
	}
	associationKey, err := requirements.PreenKey(association.Key)
	if err != nil {
		return err
	}
	problemDomainKey, err := requirements.PreenKey(association.ProblemDomainKey)
	if err != nil {
		return err
	}
	solutionDomainKey, err := requirements.PreenKey(association.SolutionDomainKey)
	if err != nil {
		return err
	}

	// Add the data.
	_, err = dbExec(dbOrTx, `
			INSERT INTO domain_association
				(
					model_key,
					association_key,
					problem_domain_key,
					solution_domain_key,
					uml_comment
				)
			VALUES
				(
					$1,
					$2,
					$3,
					$4,
					$5
				)`,
		modelKey,
		associationKey,
		problemDomainKey,
		solutionDomainKey,
		association.UmlComment)
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

// UpdateDomainAssociation updates a association in the database.
func UpdateDomainAssociation(dbOrTx DbOrTx, modelKey string, association requirements.DomainAssociation) (err error) {

	// Keys should be preened so they collide correctly.
	modelKey, err = requirements.PreenKey(modelKey)
	if err != nil {
		return err
	}
	associationKey, err := requirements.PreenKey(association.Key)
	if err != nil {
		return err
	}
	problemDomainKey, err := requirements.PreenKey(association.ProblemDomainKey)
	if err != nil {
		return err
	}
	solutionDomainKey, err := requirements.PreenKey(association.SolutionDomainKey)
	if err != nil {
		return err
	}

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
		associationKey,
		problemDomainKey,
		solutionDomainKey,
		association.UmlComment)
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

// RemoveDomainAssociation deletes a association from the database.
func RemoveDomainAssociation(dbOrTx DbOrTx, modelKey, associationKey string) (err error) {

	// Keys should be preened so they collide correctly.
	modelKey, err = requirements.PreenKey(modelKey)
	if err != nil {
		return err
	}
	associationKey, err = requirements.PreenKey(associationKey)
	if err != nil {
		return err
	}

	// Delete the data.
	_, err = dbExec(dbOrTx, `
		DELETE FROM
			domain_association
		WHERE
			association_key = $2
		AND
			model_key = $1`,
		modelKey,
		associationKey)
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

// QueryDomainAssociations loads all association from the database
func QueryDomainAssociations(dbOrTx DbOrTx, modelKey string) (associations []requirements.DomainAssociation, err error) {

	// Keys should be preened so they collide correctly.
	modelKey, err = requirements.PreenKey(modelKey)
	if err != nil {
		return nil, err
	}

	// Query the database.
	err = dbQuery(
		dbOrTx,
		func(scanner Scanner) (err error) {
			var association requirements.DomainAssociation
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
