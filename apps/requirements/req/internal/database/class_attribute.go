package database

import (
	"database/sql"
	"fmt"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_class"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_logic"

	"github.com/pkg/errors"
)

// derivationPolicyKey extracts the derivation policy key string for database storage.
// Returns nil if no derivation policy is set.
func derivationPolicyKey(attr model_class.Attribute) *string {
	if attr.DerivationPolicy != nil {
		s := attr.DerivationPolicy.Key.String()
		return &s
	}
	return nil
}

// Populate a golang struct from a database row.
func scanAttribute(scanner Scanner, classKeyPtr *identity.Key, attribute *model_class.Attribute) (err error) {
	var classKeyStr string
	var attributeKeyStr string
	var derivationPolicyKeyStr sql.NullString

	if err = scanner.Scan(
		&classKeyStr,
		&attributeKeyStr,
		&attribute.Name,
		&attribute.Details,
		&attribute.DataTypeRules,
		&derivationPolicyKeyStr,
		&attribute.Nullable,
		&attribute.UmlComment,
	); err != nil {
		if err.Error() == _POSTGRES_NOT_FOUND {
			err = ErrNotFound
		}
		return err // Do not wrap in stack here. It will be wrapped in the database calls.
	}

	// Parse the class key string into an identity.Key.
	*classKeyPtr, err = identity.ParseKey(classKeyStr)
	if err != nil {
		return err
	}

	// Parse the attribute key string into an identity.Key.
	attribute.Key, err = identity.ParseKey(attributeKeyStr)
	if err != nil {
		return err
	}

	// Parse the derivation policy key into a Logic stub if present.
	// The full Logic is stitched in top_level_requirements.go from the logics table.
	if derivationPolicyKeyStr.Valid {
		dpKey, err := identity.ParseKey(derivationPolicyKeyStr.String)
		if err != nil {
			return err
		}
		attribute.DerivationPolicy = &model_logic.Logic{Key: dpKey}
	}

	return nil
}

// LoadAttribute loads a attribute from the database.
// The returned Attribute will have DerivationPolicy as a stub (Key only);
// the full Logic is stitched in top_level_requirements.go.
func LoadAttribute(dbOrTx DbOrTx, modelKey string, attributeKey identity.Key) (classKey identity.Key, attribute model_class.Attribute, err error) {

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
			derivation_policy_key ,
			nullable              ,
			uml_comment
		FROM
			attribute
		WHERE
			attribute_key = $2
		AND
			model_key = $1`,
		modelKey,
		attributeKey.String())
	if err != nil {
		return identity.Key{}, model_class.Attribute{}, errors.WithStack(err)
	}

	return classKey, attribute, nil
}

// AddAttribute adds a attribute to the database.
func AddAttribute(dbOrTx DbOrTx, modelKey string, classKey identity.Key, attribute model_class.Attribute) (err error) {
	return AddAttributes(dbOrTx, modelKey, map[identity.Key][]model_class.Attribute{
		classKey: {attribute},
	})
}

// UpdateAttribute updates a attribute in the database.
func UpdateAttribute(dbOrTx DbOrTx, modelKey string, classKey identity.Key, attribute model_class.Attribute) (err error) {

	// Update the data.
	_, err = dbExec(dbOrTx, `
		UPDATE
			attribute
		SET
			name                  = $4 ,
			details               = $5 ,
			data_type_rules       = $6 ,
			derivation_policy_key = $7 ,
			nullable              = $8 ,
			uml_comment           = $9
		WHERE
			class_key = $2
		AND
			attribute_key = $3
		AND
			model_key = $1`,
		modelKey,
		classKey.String(),
		attribute.Key.String(),
		attribute.Name,
		attribute.Details,
		attribute.DataTypeRules,
		derivationPolicyKey(attribute),
		attribute.Nullable,
		attribute.UmlComment)
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

// RemoveAttribute deletes a attribute from the database.
func RemoveAttribute(dbOrTx DbOrTx, modelKey string, classKey identity.Key, attributeKey identity.Key) (err error) {

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
		classKey.String(),
		attributeKey.String())
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

// QueryAttributes loads all attributes from the database.
// The returned Attributes will have DerivationPolicy as a stub (Key only);
// the full Logic is stitched in top_level_requirements.go.
func QueryAttributes(dbOrTx DbOrTx, modelKey string) (attributes map[identity.Key][]model_class.Attribute, err error) {

	// Query the database.
	err = dbQuery(
		dbOrTx,
		func(scanner Scanner) (err error) {
			var classKey identity.Key
			var attribute model_class.Attribute
			if err = scanAttribute(scanner, &classKey, &attribute); err != nil {
				return errors.WithStack(err)
			}
			if attributes == nil {
				attributes = map[identity.Key][]model_class.Attribute{}
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
			derivation_policy_key ,
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

// AddAttributes adds multiple attributes to the database in a single insert.
func AddAttributes(dbOrTx DbOrTx, modelKey string, attributes map[identity.Key][]model_class.Attribute) (err error) {
	// Count total attributes.
	count := 0
	for _, attrs := range attributes {
		count += len(attrs)
	}
	if count == 0 {
		return nil
	}

	// Build the bulk insert query.
	query := `INSERT INTO attribute (model_key, class_key, attribute_key, name, details, data_type_rules, derivation_policy_key, nullable, uml_comment) VALUES `
	args := make([]interface{}, 0, count*9)
	i := 0
	for classKey, attrList := range attributes {
		for _, attr := range attrList {
			if i > 0 {
				query += ", "
			}
			base := i * 9
			query += fmt.Sprintf("($%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d)", base+1, base+2, base+3, base+4, base+5, base+6, base+7, base+8, base+9)
			args = append(args, modelKey, classKey.String(), attr.Key.String(), attr.Name, attr.Details, attr.DataTypeRules, derivationPolicyKey(attr), attr.Nullable, attr.UmlComment)
			i++
		}
	}

	_, err = dbExec(dbOrTx, query, args...)
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}
