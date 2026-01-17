package database

import (
	"fmt"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_use_case"

	"github.com/pkg/errors"
)

// Populate a golang struct from a database row.
func scanUseCase(scanner Scanner, subdomainKeyPtr *identity.Key, useCase *model_use_case.UseCase) (err error) {
	var subdomainKeyStr string
	var useCaseKeyStr string

	if err = scanner.Scan(
		&subdomainKeyStr,
		&useCaseKeyStr,
		&useCase.Name,
		&useCase.Details,
		&useCase.Level,
		&useCase.ReadOnly,
		&useCase.UmlComment,
	); err != nil {
		if err.Error() == _POSTGRES_NOT_FOUND {
			err = ErrNotFound
		}
		return err // Do not wrap in stack here. It will be wrapped in the database calls.
	}

	// Parse the key strings into identity.Key.
	*subdomainKeyPtr, err = identity.ParseKey(subdomainKeyStr)
	if err != nil {
		return err
	}
	useCase.Key, err = identity.ParseKey(useCaseKeyStr)
	if err != nil {
		return err
	}

	return nil
}

// LoadUseCase loads a use case from the database
func LoadUseCase(dbOrTx DbOrTx, modelKey string, useCaseKey identity.Key) (subdomainKey identity.Key, useCase model_use_case.UseCase, err error) {

	// Query the database.
	err = dbQueryRow(
		dbOrTx,
		func(scanner Scanner) (err error) {
			if err = scanUseCase(scanner, &subdomainKey, &useCase); err != nil {
				return err
			}
			return nil
		},
		`SELECT
			subdomain_key ,
			use_case_key  ,
			name          ,
			details       ,
			level         ,
			read_only     ,
			uml_comment
		FROM
			use_case
		WHERE
			use_case_key = $2
		AND
			model_key = $1`,
		modelKey,
		useCaseKey.String())
	if err != nil {
		return identity.Key{}, model_use_case.UseCase{}, errors.WithStack(err)
	}

	return subdomainKey, useCase, nil
}

// AddUseCase adds a use case to the database.
func AddUseCase(dbOrTx DbOrTx, modelKey string, subdomainKey identity.Key, useCase model_use_case.UseCase) (err error) {

	// Add the data.
	_, err = dbExec(dbOrTx, `
		INSERT INTO use_case
			(
				model_key     ,
				subdomain_key ,
				use_case_key  ,
				name          ,
				details       ,
				level         ,
				read_only     ,
				uml_comment
			)
		VALUES
			(
				$1,
				$2,
				$3,
				$4,
				$5,
				$6,
				$7,
				$8
			)`,
		modelKey,
		subdomainKey.String(),
		useCase.Key.String(),
		useCase.Name,
		useCase.Details,
		useCase.Level,
		useCase.ReadOnly,
		useCase.UmlComment)
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

// UpdateUseCase updates a use case in the database.
func UpdateUseCase(dbOrTx DbOrTx, modelKey string, useCase model_use_case.UseCase) (err error) {

	// Update the data.
	_, err = dbExec(dbOrTx, `
		UPDATE
			use_case
		SET
			name        = $3 ,
			details     = $4 ,
			level       = $5 ,
			read_only   = $6 ,
			uml_comment = $7
		WHERE
			model_key = $1
		AND
			use_case_key = $2`,
		modelKey,
		useCase.Key.String(),
		useCase.Name,
		useCase.Details,
		useCase.Level,
		useCase.ReadOnly,
		useCase.UmlComment)
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

// RemoveUseCase deletes a use case from the database.
func RemoveUseCase(dbOrTx DbOrTx, modelKey string, useCaseKey identity.Key) (err error) {

	// Delete the data.
	_, err = dbExec(dbOrTx, `
		DELETE FROM
			use_case
		WHERE
			model_key = $1
		AND
			use_case_key = $2`,
		modelKey,
		useCaseKey.String())
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

// QueryUseCases loads all use case from the database
func QueryUseCases(dbOrTx DbOrTx, modelKey string) (subdomainKeys map[identity.Key]identity.Key, useCases []model_use_case.UseCase, err error) {

	// Query the database.
	err = dbQuery(
		dbOrTx,
		func(scanner Scanner) (err error) {
			var subdomainKey identity.Key
			var useCase model_use_case.UseCase
			if err = scanUseCase(scanner, &subdomainKey, &useCase); err != nil {
				return errors.WithStack(err)
			}
			if subdomainKeys == nil {
				subdomainKeys = map[identity.Key]identity.Key{}
			}
			subdomainKeys[useCase.Key] = subdomainKey
			useCases = append(useCases, useCase)
			return nil
		},
		`SELECT
			subdomain_key ,
			use_case_key  ,
			name          ,
			details       ,
			level         ,
			read_only     ,
			uml_comment
		FROM
			use_case
		WHERE
			model_key = $1
		ORDER BY subdomain_key, use_case_key`,
		modelKey)
	if err != nil {
		return nil, nil, errors.WithStack(err)
	}

	return subdomainKeys, useCases, nil
}

// AddUseCases adds multiple use cases to the database in a single insert.
// Takes the same format as QueryUseCases returns: a map of useCaseKey -> subdomainKey and a slice of use cases.
func AddUseCases(dbOrTx DbOrTx, modelKey string, subdomainKeys map[identity.Key]identity.Key, useCases []model_use_case.UseCase) (err error) {
	if len(useCases) == 0 {
		return nil
	}

	// Build the bulk insert query.
	query := `INSERT INTO use_case (model_key, subdomain_key, use_case_key, name, details, level, read_only, uml_comment) VALUES `
	args := make([]interface{}, 0, len(useCases)*8)
	for i, uc := range useCases {
		if i > 0 {
			query += ", "
		}
		base := i * 8
		query += fmt.Sprintf("($%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d)", base+1, base+2, base+3, base+4, base+5, base+6, base+7, base+8)
		subdomainKey := subdomainKeys[uc.Key]
		args = append(args, modelKey, subdomainKey.String(), uc.Key.String(), uc.Name, uc.Details, uc.Level, uc.ReadOnly, uc.UmlComment)
	}

	_, err = dbExec(dbOrTx, query, args...)
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}
