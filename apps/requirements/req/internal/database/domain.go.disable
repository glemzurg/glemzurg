package database

import (
	"fmt"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_domain"

	"github.com/pkg/errors"
)

// Populate a golang struct from a database row.
func scanDomain(scanner Scanner, domain *model_domain.Domain) (err error) {
	var keyStr string

	if err = scanner.Scan(
		&keyStr,
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

	// Parse the key string into an identity.Key.
	domain.Key, err = identity.ParseKey(keyStr)
	if err != nil {
		return err
	}

	return nil
}

// LoadDomain loads a domain from the database
func LoadDomain(dbOrTx DbOrTx, modelKey string, domainKey identity.Key) (domain model_domain.Domain, err error) {

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
		domainKey.String())
	if err != nil {
		return model_domain.Domain{}, errors.WithStack(err)
	}

	return domain, nil
}

// AddDomain adds a domain to the database.
func AddDomain(dbOrTx DbOrTx, modelKey string, domain model_domain.Domain) (err error) {
	return AddDomains(dbOrTx, modelKey, []model_domain.Domain{domain})
}

// UpdateDomain updates a domain in the database.
func UpdateDomain(dbOrTx DbOrTx, modelKey string, domain model_domain.Domain) (err error) {

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
		domain.Key.String(),
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
func RemoveDomain(dbOrTx DbOrTx, modelKey string, domainKey identity.Key) (err error) {

	// Delete the data.
	_, err = dbExec(dbOrTx, `
			DELETE FROM
				domain
			WHERE
				model_key = $1
			AND
				domain_key = $2`,
		modelKey,
		domainKey.String())
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

// QueryDomains loads all domains from the database
func QueryDomains(dbOrTx DbOrTx, modelKey string) (domains []model_domain.Domain, err error) {

	// Query the database.
	err = dbQuery(
		dbOrTx,
		func(scanner Scanner) (err error) {
			var domain model_domain.Domain
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

// AddDomains adds multiple domains to the database in a single insert.
func AddDomains(dbOrTx DbOrTx, modelKey string, domains []model_domain.Domain) (err error) {
	if len(domains) == 0 {
		return nil
	}

	// Build the bulk insert query.
	query := `INSERT INTO domain (model_key, domain_key, name, details, realized, uml_comment) VALUES `
	args := make([]interface{}, 0, len(domains)*6)
	for i, domain := range domains {
		if i > 0 {
			query += ", "
		}
		base := i * 6
		query += fmt.Sprintf("($%d, $%d, $%d, $%d, $%d, $%d)", base+1, base+2, base+3, base+4, base+5, base+6)
		args = append(args, modelKey, domain.Key.String(), domain.Name, domain.Details, domain.Realized, domain.UmlComment)
	}

	_, err = dbExec(dbOrTx, query, args...)
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}
