package database

import (
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/pkg/errors"
)

// Populate a golang struct from a database row.
func scanClassIndex(scanner Scanner, indexNumPtr *uint) (err error) {
	if err = scanner.Scan(
		indexNumPtr,
	); err != nil {
		if err.Error() == _POSTGRES_NOT_FOUND {
			err = ErrNotFound
		}
		return err // Do not wrap in stack here. It will be wrapped in the database calls.
	}

	return nil
}

// LoadClassAttributeIndexes loads the indexes on a specific attribute from the database
func LoadClassAttributeIndexes(dbOrTx DbOrTx, modelKey, classKey, attributeKey string) (indexNums []uint, err error) {

	// Keys should be preened so they collide correctly.
	modelKey, err = identity.PreenKey(modelKey)
	if err != nil {
		return nil, err
	}
	classKey, err = identity.PreenKey(classKey)
	if err != nil {
		return nil, err
	}
	attributeKey, err = identity.PreenKey(attributeKey)
	if err != nil {
		return nil, err
	}

	// Query the database.
	err = dbQuery(
		dbOrTx,
		func(scanner Scanner) (err error) {
			var indexNum uint
			if err = scanClassIndex(scanner, &indexNum); err != nil {
				return errors.WithStack(err)
			}
			indexNums = append(indexNums, indexNum)
			return nil
		},
		`SELECT
				index_num
			FROM
				class_index
			WHERE
				model_key = $1
			AND
				class_key = $2
			AND
				attribute_key = $3
			ORDER BY index_num`,
		modelKey,
		classKey,
		attributeKey)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return indexNums, nil
}

// AddClassIndex adds a attribute to the database.
func AddClassIndex(dbOrTx DbOrTx, modelKey, classKey, attributeKey string, indexNum uint) (err error) {

	// Keys should be preened so they collide correctly.
	modelKey, err = identity.PreenKey(modelKey)
	if err != nil {
		return err
	}
	classKey, err = identity.PreenKey(classKey)
	if err != nil {
		return err
	}
	attributeKey, err = identity.PreenKey(attributeKey)
	if err != nil {
		return err
	}

	// Add the data.
	_, err = dbExec(dbOrTx, `
			INSERT INTO class_index
				(
					model_key     ,
					class_key     ,
					index_num     ,
					attribute_key
				)
			VALUES
				(
					$1,
					$2,
					$3,
					$4
				)`,
		modelKey,
		classKey,
		indexNum,
		attributeKey)
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

// RemoveClassIndex deletes a attribute from the database.
func RemoveClassIndex(dbOrTx DbOrTx, modelKey, classKey, attributeKey string, indexNum uint) (err error) {

	// Keys should be preened so they collide correctly.
	modelKey, err = identity.PreenKey(modelKey)
	if err != nil {
		return err
	}
	classKey, err = identity.PreenKey(classKey)
	if err != nil {
		return err
	}
	attributeKey, err = identity.PreenKey(attributeKey)
	if err != nil {
		return err
	}

	// Add the data.
	_, err = dbExec(dbOrTx, `
			DELETE FROM class_index
			WHERE
				model_key = $1
			AND
				class_key = $2
			AND
				index_num = $3
			AND
				attribute_key = $4`,
		modelKey,
		classKey,
		indexNum,
		attributeKey)
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}
