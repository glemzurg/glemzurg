package database

import (
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_use_case"

	"github.com/pkg/errors"
)

// Populate a golang struct from a database row.
func scanUseCase(scanner Scanner, subdomainKeyPtr *string, useCase *model_use_case.UseCase) (err error) {
	if err = scanner.Scan(
		subdomainKeyPtr,
		&useCase.Key,
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

	return nil
}

// LoadUseCase loads a use case from the database
func LoadUseCase(dbOrTx DbOrTx, modelKey, useCaseKey string) (subdomainKey string, useCase model_use_case.UseCase, err error) {

	// Keys should be preened so they collide correctly.
	modelKey, err = identity.PreenKey(modelKey)
	if err != nil {
		return "", model_use_case.UseCase{}, err
	}
	useCaseKey, err = identity.PreenKey(useCaseKey)
	if err != nil {
		return "", model_use_case.UseCase{}, err
	}

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
		useCaseKey)
	if err != nil {
		return "", model_use_case.UseCase{}, errors.WithStack(err)
	}

	return subdomainKey, useCase, nil
}

// AddUseCase adds a use case to the database.
func AddUseCase(dbOrTx DbOrTx, modelKey, subdomainKey string, useCase model_use_case.UseCase) (err error) {

	// Keys should be preened so they collide correctly.
	modelKey, err = identity.PreenKey(modelKey)
	if err != nil {
		return err
	}
	subdomainKey, err = identity.PreenKey(subdomainKey)
	if err != nil {
		return err
	}
	useCaseKey, err := identity.PreenKey(useCase.Key)
	if err != nil {
		return err
	}

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
		subdomainKey,
		useCaseKey,
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

	// Keys should be preened so they collide correctly.
	modelKey, err = identity.PreenKey(modelKey)
	if err != nil {
		return err
	}
	useCaseKey, err := identity.PreenKey(useCase.Key)
	if err != nil {
		return err
	}

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
		useCaseKey,
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
func RemoveUseCase(dbOrTx DbOrTx, modelKey, useCaseKey string) (err error) {

	// Keys should be preened so they collide correctly.
	modelKey, err = identity.PreenKey(modelKey)
	if err != nil {
		return err
	}
	useCaseKey, err = identity.PreenKey(useCaseKey)
	if err != nil {
		return err
	}

	// Delete the data.
	_, err = dbExec(dbOrTx, `
		DELETE FROM
			use_case
		WHERE
			model_key = $1
		AND
			use_case_key = $2`,
		modelKey,
		useCaseKey)
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

// QueryUseCases loads all use case from the database
func QueryUseCases(dbOrTx DbOrTx, modelKey string) (subdomainKeys map[string]string, useCases []model_use_case.UseCase, err error) {

	// Keys should be preened so they collide correctly.
	modelKey, err = identity.PreenKey(modelKey)
	if err != nil {
		return nil, nil, err
	}

	// Query the database.
	err = dbQuery(
		dbOrTx,
		func(scanner Scanner) (err error) {
			var subdomainKey string
			var useCase model_use_case.UseCase
			if err = scanUseCase(scanner, &subdomainKey, &useCase); err != nil {
				return errors.WithStack(err)
			}
			if subdomainKeys == nil {
				subdomainKeys = map[string]string{}
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
