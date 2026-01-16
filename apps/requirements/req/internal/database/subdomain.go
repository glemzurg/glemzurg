package database

import (
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_domain"

	"github.com/pkg/errors"
)

// Populate a golang struct from a database row.
func scanSubdomain(scanner Scanner, domainKeyPtr *identity.Key, subdomain *model_domain.Subdomain) (err error) {
	var domainKeyStr string
	var subdomainKeyStr string

	if err = scanner.Scan(
		&domainKeyStr,
		&subdomainKeyStr,
		&subdomain.Name,
		&subdomain.Details,
		&subdomain.UmlComment,
	); err != nil {
		if err.Error() == _POSTGRES_NOT_FOUND {
			err = ErrNotFound
		}
		return err // Do not wrap in stack here. It will be wrapped in the database calls.
	}

	// Parse the domain key string into an identity.Key.
	*domainKeyPtr, err = identity.ParseKey(domainKeyStr)
	if err != nil {
		return err
	}

	// Parse the subdomain key string into an identity.Key.
	subdomain.Key, err = identity.ParseKey(subdomainKeyStr)
	if err != nil {
		return err
	}

	return nil
}

// LoadSubdomain loads a subdomain from the database
func LoadSubdomain(dbOrTx DbOrTx, modelKey string, subdomainKey identity.Key) (domainKey identity.Key, subdomain model_domain.Subdomain, err error) {

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
		subdomainKey.String())
	if err != nil {
		return identity.Key{}, model_domain.Subdomain{}, errors.WithStack(err)
	}

	return domainKey, subdomain, nil
}

// AddSubdomain adds a subdomain to the database.
func AddSubdomain(dbOrTx DbOrTx, modelKey string, domainKey identity.Key, subdomain model_domain.Subdomain) (err error) {

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
		domainKey.String(),
		subdomain.Key.String(),
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
		subdomain.Key.String(),
		subdomain.Name,
		subdomain.Details,
		subdomain.UmlComment)
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

// RemoveSubdomain deletes a subdomain from the database.
func RemoveSubdomain(dbOrTx DbOrTx, modelKey string, subdomainKey identity.Key) (err error) {

	// Delete the data.
	_, err = dbExec(dbOrTx, `
			DELETE FROM
				subdomain
			WHERE
				model_key = $1
			AND
				subdomain_key = $2`,
		modelKey,
		subdomainKey.String())
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

// QuerySubdomains loads all subdomains from the database
func QuerySubdomains(dbOrTx DbOrTx, modelKey string) (subdomains map[identity.Key][]model_domain.Subdomain, err error) {

	// Query the database.
	err = dbQuery(
		dbOrTx,
		func(scanner Scanner) (err error) {
			var domainKey identity.Key
			var subdomain model_domain.Subdomain
			if err = scanSubdomain(scanner, &domainKey, &subdomain); err != nil {
				return errors.WithStack(err)
			}
			if subdomains == nil {
				subdomains = map[identity.Key][]model_domain.Subdomain{}
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
