package database

import (
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_class"

	"github.com/pkg/errors"
)

// Populate a golang struct from a database row.
func scanAttribute(scanner Scanner, classKeyPtr *string, attribute *model_class.Attribute) (err error) {
	if err = scanner.Scan(
		classKeyPtr,
		&attribute.Key,
		&attribute.Name,
		&attribute.Details,
		&attribute.DataTypeRules,
		&attribute.DerivationPolicy,
		&attribute.Nullable,
		&attribute.UmlComment,
	); err != nil {
		if err.Error() == _POSTGRES_NOT_FOUND {
			err = ErrNotFound
		}
		return err // Do not wrap in stack here. It will be wrapped in the database calls.
	}

	return nil
}

// LoadAttribute loads a attribute from the database
func LoadAttribute(dbOrTx DbOrTx, modelKey, attributeKey string) (classKey string, attribute model_class.Attribute, err error) {

	// Keys should be preened so they collide correctly.
	modelKey, err = identity.PreenKey(modelKey)
	if err != nil {
		return "", model_class.Attribute{}, err
	}
	attributeKey, err = identity.PreenKey(attributeKey)
	if err != nil {
		return "", model_class.Attribute{}, err
	}

	// Query the database.
	err = dbQueryRow(
		dbOrTx,
		func(scanner Scanner) (err error) {
			if err = scanAttribute(scanner, &classKey, &attribute); err != nil {
				return err
			}
			return nil
		},
		`SELECT
			class_key             ,
			attribute_key         ,
			name                  ,
			details               ,
			data_type_rules       ,
			derivation_policy     ,
			nullable              ,
			uml_comment
		FROM
			attribute
		WHERE
			attribute_key = $2
		AND
			model_key = $1`,
		modelKey,
		attributeKey)
	if err != nil {
		return "", model_class.Attribute{}, errors.WithStack(err)
	}

	return classKey, attribute, nil
}

// AddAttribute adds a attribute to the database.
func AddAttribute(dbOrTx DbOrTx, modelKey, classKey string, attribute model_class.Attribute) (err error) {

	// Keys should be preened so they collide correctly.
	modelKey, err = identity.PreenKey(modelKey)
	if err != nil {
		return err
	}
	classKey, err = identity.PreenKey(classKey)
	if err != nil {
		return err
	}
	attributeKey, err := identity.PreenKey(attribute.Key)
	if err != nil {
		return err
	}

	// Add the data.
	_, err = dbExec(dbOrTx, `
			INSERT INTO attribute
				(
					model_key    ,
					class_key,
					attribute_key,
					name,
					details,
					data_type_rules,
					derivation_policy,
					nullable,
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
					$8,
					$9
				)`,
		modelKey,
		classKey,
		attributeKey,
		attribute.Name,
		attribute.Details,
		attribute.DataTypeRules,
		attribute.DerivationPolicy,
		attribute.Nullable,
		attribute.UmlComment)
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

// UpdateAttribute updates a attribute in the database.
func UpdateAttribute(dbOrTx DbOrTx, modelKey, classKey string, attribute model_class.Attribute) (err error) {

	// Keys should be preened so they collide correctly.
	modelKey, err = identity.PreenKey(modelKey)
	if err != nil {
		return err
	}
	classKey, err = identity.PreenKey(classKey)
	if err != nil {
		return err
	}
	attributeKey, err := identity.PreenKey(attribute.Key)
	if err != nil {
		return err
	}

	// Update the data.
	_, err = dbExec(dbOrTx, `
		UPDATE
			attribute
		SET
			name                  = $4 ,
			details               = $5 ,
			data_type_rules       = $6 ,
			derivation_policy     = $7 ,
			nullable              = $8 ,
			uml_comment           = $9 
		WHERE
			class_key = $2
		AND
			attribute_key = $3
		AND
			model_key = $1`,
		modelKey,
		classKey,
		attributeKey,
		attribute.Name,
		attribute.Details,
		attribute.DataTypeRules,
		attribute.DerivationPolicy,
		attribute.Nullable,
		attribute.UmlComment)
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

// RemoveAttribute deletes a attribute from the database.
func RemoveAttribute(dbOrTx DbOrTx, modelKey, classKey, attributeKey string) (err error) {

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

	// Delete the data.
	_, err = dbExec(dbOrTx, `
		DELETE FROM
			attribute
		WHERE
			class_key = $2
		AND
			attribute_key = $3
		AND
			model_key = $1`,
		modelKey,
		classKey,
		attributeKey)
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

// QueryAttributes loads all attribute from the database
func QueryAttributes(dbOrTx DbOrTx, modelKey string) (attributes map[string][]model_class.Attribute, err error) {

	// Keys should be preened so they collide correctly.
	modelKey, err = identity.PreenKey(modelKey)
	if err != nil {
		return nil, err
	}

	// Query the database.
	err = dbQuery(
		dbOrTx,
		func(scanner Scanner) (err error) {
			var classKey string
			var attribute model_class.Attribute
			if err = scanAttribute(scanner, &classKey, &attribute); err != nil {
				return errors.WithStack(err)
			}
			if attributes == nil {
				attributes = map[string][]model_class.Attribute{}
			}
			classAttributes := attributes[classKey]
			classAttributes = append(classAttributes, attribute)
			attributes[classKey] = classAttributes
			return nil
		},
		`SELECT
			class_key             ,
			attribute_key         ,
			name                  ,
			details               ,
			data_type_rules       ,
			derivation_policy     ,
			nullable              ,
			uml_comment
		FROM
			attribute
		WHERE
			model_key = $1
		ORDER BY class_key, attribute_key`,
		modelKey)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return attributes, nil
}
