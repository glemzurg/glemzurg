package database

import (
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/requirements"

	"github.com/pkg/errors"
)

// Populate a golang struct from a database row.
func scanDomain(scanner Scanner, domain *requirements.Domain) (err error) {
	if err = scanner.Scan(
		&domain.Key,
		&domain.Name,
		&domain.Details,
		&domain.Realized,
		&domain.UmlComment,
	); err != nil {
		if err.Error() == _POSTGRES_NOT_FOUND {
			err = ErrNotFound
		}
		return err // Do not wrap in stack here. It will be wrapped in the database calls.
	}

	return nil
}

// LoadDomain loads a domain from the database
func LoadDomain(dbOrTx DbOrTx, modelKey, domainKey string) (domain requirements.Domain, err error) {

	// Keys should be preened so they collide correctly.
	modelKey, err = requirements.PreenKey(modelKey)
	if err != nil {
		return requirements.Domain{}, err
	}
	domainKey, err = requirements.PreenKey(domainKey)
	if err != nil {
		return requirements.Domain{}, err
	}

	// Query the database.
	err = dbQueryRow(
		dbOrTx,
		func(scanner Scanner) (err error) {
			if err = scanDomain(scanner, &domain); err != nil {
				return err
			}
			return nil
		},
		`SELECT
			domain_key  ,
			name        ,
			details     ,
			realized    ,
			uml_comment
		FROM
			domain
		WHERE
			domain_key = $2
		AND
			model_key = $1`,
		modelKey,
		domainKey)
	if err != nil {
		return requirements.Domain{}, errors.WithStack(err)
	}

	return domain, nil
}

// AddDomain adds a domain to the database.
func AddDomain(dbOrTx DbOrTx, modelKey string, domain requirements.Domain) (err error) {

	// Keys should be preened so they collide correctly.
	modelKey, err = requirements.PreenKey(modelKey)
	if err != nil {
		return err
	}
	domainKey, err := requirements.PreenKey(domain.Key)
	if err != nil {
		return err
	}

	// Add the data.
	_, err = dbExec(dbOrTx, `
			INSERT INTO domain
				(
					model_key   ,
					domain_key  ,
					name        ,
					details     ,
					realized    ,
					uml_comment
				)
			VALUES
				(
					$1,
					$2,
					$3,
					$4,
					$5,
					$6
				)`,
		modelKey,
		domainKey,
		domain.Name,
		domain.Details,
		domain.Realized,
		domain.UmlComment)
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

// UpdateDomain updates a domain in the database.
func UpdateDomain(dbOrTx DbOrTx, modelKey string, domain requirements.Domain) (err error) {

	// Keys should be preened so they collide correctly.
	modelKey, err = requirements.PreenKey(modelKey)
	if err != nil {
		return err
	}
	domainKey, err := requirements.PreenKey(domain.Key)
	if err != nil {
		return err
	}

	// Update the data.
	_, err = dbExec(dbOrTx, `
		UPDATE
			domain
		SET
			name        = $3 ,
			details     = $4 ,
			realized    = $5 ,
			uml_comment = $6
		WHERE
			model_key = $1
		AND
			domain_key = $2`,
		modelKey,
		domainKey,
		domain.Name,
		domain.Details,
		domain.Realized,
		domain.UmlComment)
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

// RemoveDomain deletes a domain from the database.
func RemoveDomain(dbOrTx DbOrTx, modelKey, domainKey string) (err error) {

	// Keys should be preened so they collide correctly.
	modelKey, err = requirements.PreenKey(modelKey)
	if err != nil {
		return err
	}
	domainKey, err = requirements.PreenKey(domainKey)
	if err != nil {
		return err
	}

	// Delete the data.
	_, err = dbExec(dbOrTx, `
			DELETE FROM
				domain
			WHERE
				model_key = $1
			AND
				domain_key = $2`,
		modelKey,
		domainKey)
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

// QueryDomains loads all domains from the database
func QueryDomains(dbOrTx DbOrTx, modelKey string) (domains []requirements.Domain, err error) {

	// Keys should be preened so they collide correctly.
	modelKey, err = requirements.PreenKey(modelKey)
	if err != nil {
		return nil, err
	}

	// Query the database.
	err = dbQuery(
		dbOrTx,
		func(scanner Scanner) (err error) {
			var domain requirements.Domain
			if err = scanDomain(scanner, &domain); err != nil {
				return errors.WithStack(err)
			}
			domains = append(domains, domain)
			return nil
		},
		`SELECT
				domain_key  ,
				name        ,
				details     ,
				realized    ,
				uml_comment
			FROM
				domain
			WHERE
				model_key = $1
			ORDER BY domain_key`,
		modelKey)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return domains, nil
}
