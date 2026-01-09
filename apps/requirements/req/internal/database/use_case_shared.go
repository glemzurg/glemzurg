package database

import (
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/requirements/model_use_case"

	"github.com/pkg/errors"
)

// Populate a golang struct from a database row.
func scanUseCaseShared(scanner Scanner, seaLevelKeyPtr, mudlevelKeyPtr *string, useCaseShared *model_use_case.UseCaseShared) (err error) {
	if err = scanner.Scan(
		seaLevelKeyPtr,
		mudlevelKeyPtr,
		&useCaseShared.ShareType,
		&useCaseShared.UmlComment,
	); err != nil {
		if err.Error() == _POSTGRES_NOT_FOUND {
			err = ErrNotFound
		}
		return err // Do not wrap in stack here. It will be wrapped in the database calls.
	}

	return nil
}

// LoadUseCaseShared loads a use case from the database
func LoadUseCaseShared(dbOrTx DbOrTx, modelKey, seaLevelKey, mudLevelKey string) (useCaseShared model_use_case.UseCaseShared, err error) {

	// Keys should be preened so they collide correctly.
	modelKey, err = identity.PreenKey(modelKey)
	if err != nil {
		return model_use_case.UseCaseShared{}, err
	}
	seaLevelKey, err = identity.PreenKey(seaLevelKey)
	if err != nil {
		return model_use_case.UseCaseShared{}, err
	}
	mudLevelKey, err = identity.PreenKey(mudLevelKey)
	if err != nil {
		return model_use_case.UseCaseShared{}, err
	}

	// Query the database.
	err = dbQueryRow(
		dbOrTx,
		func(scanner Scanner) (err error) {
			var unusedSeaLevelKey, unusedMudLevelKey string
			if err = scanUseCaseShared(scanner, &unusedSeaLevelKey, &unusedMudLevelKey, &useCaseShared); err != nil {
				return err
			}
			// Not using the keys since this code already has them.
			_, _ = unusedSeaLevelKey, unusedMudLevelKey
			return nil
		},
		`SELECT
			sea_use_case_key ,
			mud_use_case_key ,
			share_type       ,
			uml_comment
		FROM
			use_case_shared
		WHERE
			sea_use_case_key = $2
		AND
			mud_use_case_key = $3
		AND
			model_key = $1`,
		modelKey,
		seaLevelKey,
		mudLevelKey)
	if err != nil {
		return model_use_case.UseCaseShared{}, errors.WithStack(err)
	}

	return useCaseShared, nil
}

// AddUseCaseShared adds a use case to the database.
func AddUseCaseShared(dbOrTx DbOrTx, modelKey, seaLevelKey, mudLevelKey string, useCaseShared model_use_case.UseCaseShared) (err error) {

	// Keys should be preened so they collide correctly.
	modelKey, err = identity.PreenKey(modelKey)
	if err != nil {
		return err
	}
	seaLevelKey, err = identity.PreenKey(seaLevelKey)
	if err != nil {
		return err
	}
	mudLevelKey, err = identity.PreenKey(mudLevelKey)
	if err != nil {
		return err
	}

	// Add the data.
	_, err = dbExec(dbOrTx, `
		INSERT INTO use_case_shared
			(
				model_key        ,
				sea_use_case_key ,
				mud_use_case_key ,
				share_type       ,
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
		seaLevelKey,
		mudLevelKey,
		useCaseShared.ShareType,
		useCaseShared.UmlComment)
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

// UpdateUseCaseShared updates a use case in the database.
func UpdateUseCaseShared(dbOrTx DbOrTx, modelKey, seaLevelKey, mudLevelKey string, useCaseShared model_use_case.UseCaseShared) (err error) {

	// Keys should be preened so they collide correctly.
	modelKey, err = identity.PreenKey(modelKey)
	if err != nil {
		return err
	}
	seaLevelKey, err = identity.PreenKey(seaLevelKey)
	if err != nil {
		return err
	}
	mudLevelKey, err = identity.PreenKey(mudLevelKey)
	if err != nil {
		return err
	}

	// Update the data.
	_, err = dbExec(dbOrTx, `
		UPDATE
			use_case_shared
		SET
			share_type  = $4 ,
			uml_comment = $5
		WHERE
			sea_use_case_key = $2
		AND
			mud_use_case_key = $3
		AND
			model_key = $1`,
		modelKey,
		seaLevelKey,
		mudLevelKey,
		useCaseShared.ShareType,
		useCaseShared.UmlComment)
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

// RemoveUseCaseShared deletes a use case from the database.
func RemoveUseCaseShared(dbOrTx DbOrTx, modelKey, seaLevelKey, mudLevelKey string) (err error) {

	// Keys should be preened so they collide correctly.
	modelKey, err = identity.PreenKey(modelKey)
	if err != nil {
		return err
	}
	seaLevelKey, err = identity.PreenKey(seaLevelKey)
	if err != nil {
		return err
	}
	mudLevelKey, err = identity.PreenKey(mudLevelKey)
	if err != nil {
		return err
	}

	// Delete the data.
	_, err = dbExec(dbOrTx, `
		DELETE FROM
			use_case_shared
		WHERE
			sea_use_case_key = $2
		AND
			mud_use_case_key = $3
		AND
			model_key = $1`,
		modelKey,
		seaLevelKey,
		mudLevelKey)
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

// QueryUseCaseShareds loads all use case from the database
func QueryUseCaseShareds(dbOrTx DbOrTx, modelKey string) (useCaseShareds map[string]map[string]model_use_case.UseCaseShared, err error) {

	// Keys should be preened so they collide correctly.
	modelKey, err = identity.PreenKey(modelKey)
	if err != nil {
		return nil, err
	}

	// Query the database.
	err = dbQuery(
		dbOrTx,
		func(scanner Scanner) (err error) {
			var seaLevelKey, mudLevelKey string
			var useCaseShared model_use_case.UseCaseShared
			if err = scanUseCaseShared(scanner, &seaLevelKey, &mudLevelKey, &useCaseShared); err != nil {
				return errors.WithStack(err)
			}
			if useCaseShareds == nil {
				useCaseShareds = map[string]map[string]model_use_case.UseCaseShared{}
			}
			oneUseCaseShareds := useCaseShareds[seaLevelKey]
			if oneUseCaseShareds == nil {
				oneUseCaseShareds = map[string]model_use_case.UseCaseShared{}
			}
			oneUseCaseShareds[mudLevelKey] = useCaseShared
			useCaseShareds[seaLevelKey] = oneUseCaseShareds
			return nil
		},
		`SELECT
			sea_use_case_key ,
			mud_use_case_key ,
			share_type       ,
			uml_comment
		FROM
			use_case_shared
		WHERE
			model_key = $1
		ORDER BY mud_use_case_key, sea_use_case_key`,
		modelKey)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return useCaseShareds, nil
}
