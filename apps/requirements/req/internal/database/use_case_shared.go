package database

import (
	"fmt"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_use_case"

	"github.com/pkg/errors"
)

// Populate a golang struct from a database row.
func scanUseCaseShared(scanner Scanner, seaLevelKeyPtr, mudlevelKeyPtr *identity.Key, useCaseShared *model_use_case.UseCaseShared) (err error) {
	var seaLevelKeyStr string
	var mudLevelKeyStr string

	if err = scanner.Scan(
		&seaLevelKeyStr,
		&mudLevelKeyStr,
		&useCaseShared.ShareType,
		&useCaseShared.UmlComment,
	); err != nil {
		if err.Error() == _POSTGRES_NOT_FOUND {
			err = ErrNotFound
		}
		return err // Do not wrap in stack here. It will be wrapped in the database calls.
	}

	// Parse the sea level key string into an identity.Key.
	*seaLevelKeyPtr, err = identity.ParseKey(seaLevelKeyStr)
	if err != nil {
		return err
	}

	// Parse the mud level key string into an identity.Key.
	*mudlevelKeyPtr, err = identity.ParseKey(mudLevelKeyStr)
	if err != nil {
		return err
	}

	return nil
}

// LoadUseCaseShared loads a use case from the database
func LoadUseCaseShared(dbOrTx DbOrTx, modelKey string, seaLevelKey identity.Key, mudLevelKey identity.Key) (useCaseShared model_use_case.UseCaseShared, err error) {

	// Query the database.
	err = dbQueryRow(
		dbOrTx,
		func(scanner Scanner) (err error) {
			var unusedSeaLevelKey, unusedMudLevelKey identity.Key
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
		seaLevelKey.String(),
		mudLevelKey.String())
	if err != nil {
		return model_use_case.UseCaseShared{}, errors.WithStack(err)
	}

	return useCaseShared, nil
}

// AddUseCaseShared adds a use case to the database.
func AddUseCaseShared(dbOrTx DbOrTx, modelKey string, seaLevelKey identity.Key, mudLevelKey identity.Key, useCaseShared model_use_case.UseCaseShared) (err error) {
	return AddUseCaseShareds(dbOrTx, modelKey, map[identity.Key]map[identity.Key]model_use_case.UseCaseShared{
		seaLevelKey: {mudLevelKey: useCaseShared},
	})
}

// UpdateUseCaseShared updates a use case in the database.
func UpdateUseCaseShared(dbOrTx DbOrTx, modelKey string, seaLevelKey identity.Key, mudLevelKey identity.Key, useCaseShared model_use_case.UseCaseShared) (err error) {

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
		seaLevelKey.String(),
		mudLevelKey.String(),
		useCaseShared.ShareType,
		useCaseShared.UmlComment)
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

// RemoveUseCaseShared deletes a use case from the database.
func RemoveUseCaseShared(dbOrTx DbOrTx, modelKey string, seaLevelKey identity.Key, mudLevelKey identity.Key) (err error) {

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
		seaLevelKey.String(),
		mudLevelKey.String())
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

// QueryUseCaseShareds loads all use case from the database
func QueryUseCaseShareds(dbOrTx DbOrTx, modelKey string) (useCaseShareds map[identity.Key]map[identity.Key]model_use_case.UseCaseShared, err error) {

	useCaseShareds = make(map[identity.Key]map[identity.Key]model_use_case.UseCaseShared)

	// Query the database.
	err = dbQuery(
		dbOrTx,
		func(scanner Scanner) (err error) {
			var seaLevelKey, mudLevelKey identity.Key
			var useCaseShared model_use_case.UseCaseShared
			if err = scanUseCaseShared(scanner, &seaLevelKey, &mudLevelKey, &useCaseShared); err != nil {
				return errors.WithStack(err)
			}
			oneUseCaseShareds := useCaseShareds[seaLevelKey]
			if oneUseCaseShareds == nil {
				oneUseCaseShareds = map[identity.Key]model_use_case.UseCaseShared{}
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

// AddUseCaseShareds adds multiple use case shared entries to the database in a single insert.
func AddUseCaseShareds(dbOrTx DbOrTx, modelKey string, useCaseShareds map[identity.Key]map[identity.Key]model_use_case.UseCaseShared) (err error) {
	// Count total entries.
	count := 0
	for _, sharedMap := range useCaseShareds {
		count += len(sharedMap)
	}
	if count == 0 {
		return nil
	}

	// Build the bulk insert query.
	query := `INSERT INTO use_case_shared (model_key, sea_use_case_key, mud_use_case_key, share_type, uml_comment) VALUES `
	args := make([]interface{}, 0, count*5)
	i := 0
	for seaLevelKey, sharedMap := range useCaseShareds {
		for mudLevelKey, shared := range sharedMap {
			if i > 0 {
				query += ", "
			}
			base := i * 5
			query += fmt.Sprintf("($%d, $%d, $%d, $%d, $%d)", base+1, base+2, base+3, base+4, base+5)
			args = append(args, modelKey, seaLevelKey.String(), mudLevelKey.String(), shared.ShareType, shared.UmlComment)
			i++
		}
	}

	_, err = dbExec(dbOrTx, query, args...)
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}
