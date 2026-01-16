package database

import (
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_domain"

	"github.com/pkg/errors"
)

// Populate a golang struct from a database row.
func scanSubdomain(scanner Scanner, domainKeyPtr *string, subdomain *model_domain.Subdomain) (err error) {
	if err = scanner.Scan(
		domainKeyPtr,
		&subdomain.Key,
		&subdomain.Name,
		&subdomain.Details,
		&subdomain.UmlComment,
	); err != nil {
		if err.Error() == _POSTGRES_NOT_FOUND {
			err = ErrNotFound
		}
		return err // Do not wrap in stack here. It will be wrapped in the database calls.
	}

	return nil
}

// LoadSubdomain loads a subdomain from the database
func LoadSubdomain(dbOrTx DbOrTx, modelKey, subdomainKey string) (domainKey string, subdomain model_domain.Subdomain, err error) {

	// Keys should be preened so they collide correctly.
	modelKey, err = identity.PreenKey(modelKey)
	if err != nil {
		return "", model_domain.Subdomain{}, err
	}
	subdomainKey, err = identity.PreenKey(subdomainKey)
	if err != nil {
		return "", model_domain.Subdomain{}, err
	}

	// Query the database.
	err = dbQueryRow(
		dbOrTx,
		func(scanner Scanner) (err error) {
			if err = scanSubdomain(scanner, &domainKey, &subdomain); err != nil {
				return err
			}
			return nil
		},
		`SELECT
			domain_key    ,
			subdomain_key ,
			name          ,
			details       ,
			uml_comment
		FROM
			subdomain
		WHERE
			subdomain_key = $2
		AND
			model_key = $1`,
		modelKey,
		subdomainKey)
	if err != nil {
		return "", model_domain.Subdomain{}, errors.WithStack(err)
	}

	return domainKey, subdomain, nil
}

// AddSubdomain adds a subdomain to the database.
func AddSubdomain(dbOrTx DbOrTx, modelKey, domainKey string, subdomain model_domain.Subdomain) (err error) {

	// Keys should be preened so they collide correctly.
	modelKey, err = identity.PreenKey(modelKey)
	if err != nil {
		return err
	}
	domainKey, err = identity.PreenKey(domainKey)
	if err != nil {
		return err
	}
	subdomainKey, err := identity.PreenKey(subdomain.Key)
	if err != nil {
		return err
	}

	// Add the data.
	_, err = dbExec(dbOrTx, `
			INSERT INTO subdomain
				(
					model_key     ,
					domain_key    ,
					subdomain_key ,
					name          ,
					details       ,
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
		subdomainKey,
		subdomain.Name,
		subdomain.Details,
		subdomain.UmlComment)
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

// UpdateSubdomain updates a subdomain in the database.
func UpdateSubdomain(dbOrTx DbOrTx, modelKey string, subdomain model_domain.Subdomain) (err error) {

	// Keys should be preened so they collide correctly.
	modelKey, err = identity.PreenKey(modelKey)
	if err != nil {
		return err
	}
	subdomainKey, err := identity.PreenKey(subdomain.Key)
	if err != nil {
		return err
	}

	// Update the data.
	_, err = dbExec(dbOrTx, `
		UPDATE
			subdomain
		SET
			name        = $3 ,
			details     = $4 ,
			uml_comment = $5
		WHERE
			model_key = $1
		AND
			subdomain_key = $2`,
		modelKey,
		subdomainKey,
		subdomain.Name,
		subdomain.Details,
		subdomain.UmlComment)
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

// RemoveSubdomain deletes a subdomain from the database.
func RemoveSubdomain(dbOrTx DbOrTx, modelKey, subdomainKey string) (err error) {

	// Keys should be preened so they collide correctly.
	modelKey, err = identity.PreenKey(modelKey)
	if err != nil {
		return err
	}
	subdomainKey, err = identity.PreenKey(subdomainKey)
	if err != nil {
		return err
	}

	// Delete the data.
	_, err = dbExec(dbOrTx, `
			DELETE FROM
				subdomain
			WHERE
				model_key = $1
			AND
				subdomain_key = $2`,
		modelKey,
		subdomainKey)
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

// QuerySubdomains loads all subdomains from the database
func QuerySubdomains(dbOrTx DbOrTx, modelKey string) (subdomains map[string][]model_domain.Subdomain, err error) {

	// Keys should be preened so they collide correctly.
	modelKey, err = identity.PreenKey(modelKey)
	if err != nil {
		return nil, err
	}

	// Query the database.
	err = dbQuery(
		dbOrTx,
		func(scanner Scanner) (err error) {
			var domainKey string
			var subdomain model_domain.Subdomain
			if err = scanSubdomain(scanner, &domainKey, &subdomain); err != nil {
				return errors.WithStack(err)
			}
			if subdomains == nil {
				subdomains = map[string][]model_domain.Subdomain{}
			}
			domainSubdomains := subdomains[domainKey]
			domainSubdomains = append(domainSubdomains, subdomain)
			subdomains[domainKey] = domainSubdomains
			return nil
		},
		`SELECT
				domain_key    ,
				subdomain_key ,
				name          ,
				details       ,
				uml_comment
			FROM
				subdomain
			WHERE
				model_key = $1
			ORDER BY domain_key, subdomain_key`,
		modelKey)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return subdomains, nil
}
