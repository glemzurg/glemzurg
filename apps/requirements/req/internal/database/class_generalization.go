package database

import (
	"fmt"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_class"

	"github.com/pkg/errors"
)

// Populate a golang struct from a database row.
func scanGeneralization(scanner Scanner, subdomainKeyPtr *identity.Key, generalization *model_class.Generalization) (err error) {
	var subdomainKeyStr string
	var keyStr string

	if err = scanner.Scan(
		&subdomainKeyStr,
		&keyStr,
		&generalization.Name,
		&generalization.Details,
		&generalization.IsComplete,
		&generalization.IsStatic,
		&generalization.UmlComment,
	); err != nil {
		if err.Error() == _POSTGRES_NOT_FOUND {
			err = ErrNotFound
		}
		return err // Do not wrap in stack here. It will be wrapped in the database calls.
	}

	// Parse the subdomain key string into an identity.Key.
	*subdomainKeyPtr, err = identity.ParseKey(subdomainKeyStr)
	if err != nil {
		return err
	}

	// Parse the key string into an identity.Key.
	generalization.Key, err = identity.ParseKey(keyStr)
	if err != nil {
		return err
	}

	return nil
}

// LoadGeneralization loads a generalization from the database.
func LoadGeneralization(dbOrTx DbOrTx, modelKey string, generalizationKey identity.Key) (subdomainKey identity.Key, generalization model_class.Generalization, err error) {

	// Query the database.
	err = dbQueryRow(
		dbOrTx,
		func(scanner Scanner) (err error) {
			if err = scanGeneralization(scanner, &subdomainKey, &generalization); err != nil {
				return err
			}
			return nil
		},
		`SELECT
			subdomain_key      ,
			generalization_key ,
			name               ,
			details            ,
			is_complete        ,
			is_static          ,
			uml_comment
		FROM
			class_generalization
		WHERE
			generalization_key = $2
		AND
			model_key = $1`,
		modelKey,
		generalizationKey.String())
	if err != nil {
		return identity.Key{}, model_class.Generalization{}, errors.WithStack(err)
	}

	return subdomainKey, generalization, nil
}

// AddGeneralization adds a generalization to the database.
func AddGeneralization(dbOrTx DbOrTx, modelKey string, subdomainKey identity.Key, generalization model_class.Generalization) (err error) {
	return AddGeneralizations(dbOrTx, modelKey, map[identity.Key][]model_class.Generalization{
		subdomainKey: {generalization},
	})
}

// UpdateGeneralization updates a generalization in the database.
func UpdateGeneralization(dbOrTx DbOrTx, modelKey string, generalization model_class.Generalization) (err error) {

	// Update the data.
	_, err = dbExec(dbOrTx, `
		UPDATE
			class_generalization
		SET
			name        = $3 ,
			details     = $4 ,
			is_complete = $5 ,
			is_static   = $6 ,
			uml_comment = $7
		WHERE
			model_key = $1
		AND
			generalization_key = $2`,
		modelKey,
		generalization.Key.String(),
		generalization.Name,
		generalization.Details,
		generalization.IsComplete,
		generalization.IsStatic,
		generalization.UmlComment)
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

// RemoveGeneralization deletes a generalization from the database.
func RemoveGeneralization(dbOrTx DbOrTx, modelKey string, generalizationKey identity.Key) (err error) {

	// Delete the data.
	_, err = dbExec(dbOrTx, `
			DELETE FROM
				class_generalization
			WHERE
				model_key = $1
			AND
				generalization_key = $2`,
		modelKey,
		generalizationKey.String())
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

// QueryGeneralizations loads all generalizations from the database grouped by subdomain key.
func QueryGeneralizations(dbOrTx DbOrTx, modelKey string) (generalizations map[identity.Key][]model_class.Generalization, err error) {

	// Query the database.
	err = dbQuery(
		dbOrTx,
		func(scanner Scanner) (err error) {
			var subdomainKey identity.Key
			var generalization model_class.Generalization
			if err = scanGeneralization(scanner, &subdomainKey, &generalization); err != nil {
				return errors.WithStack(err)
			}
			if generalizations == nil {
				generalizations = map[identity.Key][]model_class.Generalization{}
			}
			generalizations[subdomainKey] = append(generalizations[subdomainKey], generalization)
			return nil
		},
		`SELECT
			subdomain_key      ,
			generalization_key ,
			name               ,
			details            ,
			is_complete        ,
			is_static          ,
			uml_comment
		FROM
			class_generalization
		WHERE
			model_key = $1
		ORDER BY subdomain_key, generalization_key`,
		modelKey)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return generalizations, nil
}

// AddGeneralizations adds multiple generalizations to the database in a single insert.
func AddGeneralizations(dbOrTx DbOrTx, modelKey string, generalizations map[identity.Key][]model_class.Generalization) (err error) {
	// Count total generalizations.
	count := 0
	for _, gens := range generalizations {
		count += len(gens)
	}
	if count == 0 {
		return nil
	}

	// Build the bulk insert query.
	query := `INSERT INTO class_generalization (model_key, subdomain_key, generalization_key, name, details, is_complete, is_static, uml_comment) VALUES `
	args := make([]interface{}, 0, count*8)
	i := 0
	for subdomainKey, gens := range generalizations {
		for _, gen := range gens {
			if i > 0 {
				query += ", "
			}
			base := i * 8
			query += fmt.Sprintf("($%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d)", base+1, base+2, base+3, base+4, base+5, base+6, base+7, base+8)
			args = append(args, modelKey, subdomainKey.String(), gen.Key.String(), gen.Name, gen.Details, gen.IsComplete, gen.IsStatic, gen.UmlComment)
			i++
		}
	}

	_, err = dbExec(dbOrTx, query, args...)
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}
