package database

import (
	"fmt"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_class"

	"github.com/pkg/errors"
)

// Populate a golang struct from a database row.
func scanAssociation(scanner Scanner, association *model_class.Association) (err error) {
	var associationKeyStr string
	var fromClassKeyStr string
	var toClassKeyStr string
	var associationClassKeyPtr *string
	var fromLowerBound, fromHigherBound, toLowerBound, toHigherBound uint

	if err = scanner.Scan(
		&associationKeyStr,
		&association.Name,
		&association.Details,
		&fromClassKeyStr,
		&fromLowerBound,
		&fromHigherBound,
		&toClassKeyStr,
		&toLowerBound,
		&toHigherBound,
		&associationClassKeyPtr,
		&association.UmlComment,
	); err != nil {
		if err.Error() == _POSTGRES_NOT_FOUND {
			err = ErrNotFound
		}
		return err // Do not wrap in stack here. It will be wrapped in the database calls.
	}

	// Parse the association key string into an identity.Key.
	association.Key, err = identity.ParseKey(associationKeyStr)
	if err != nil {
		return err
	}

	// Parse the from class key string into an identity.Key.
	association.FromClassKey, err = identity.ParseKey(fromClassKeyStr)
	if err != nil {
		return err
	}

	// Parse the to class key string into an identity.Key.
	association.ToClassKey, err = identity.ParseKey(toClassKeyStr)
	if err != nil {
		return err
	}

	association.FromMultiplicity = model_class.Multiplicity{LowerBound: fromLowerBound, HigherBound: fromHigherBound}
	association.ToMultiplicity = model_class.Multiplicity{LowerBound: toLowerBound, HigherBound: toHigherBound}

	if associationClassKeyPtr != nil {
		associationClassKey, err := identity.ParseKey(*associationClassKeyPtr)
		if err != nil {
			return err
		}
		association.AssociationClassKey = &associationClassKey
	}

	return nil
}

// LoadAssociation loads a association from the database
func LoadAssociation(dbOrTx DbOrTx, modelKey string, associationKey identity.Key) (association model_class.Association, err error) {

	// Query the database.
	err = dbQueryRow(
		dbOrTx,
		func(scanner Scanner) (err error) {
			if err = scanAssociation(scanner, &association); err != nil {
				return err
			}
			return nil
		},
		`SELECT
			association_key          ,
			name                     ,
			details                  ,
			from_class_key           ,
			from_multiplicity_lower  ,
			from_multiplicity_higher ,
			to_class_key             ,
			to_multiplicity_lower    ,
			to_multiplicity_higher   ,
			association_class_key    ,
			uml_comment
		FROM
			association
		WHERE
			association_key = $2
		AND
			model_key = $1
		ORDER BY association_key`,
		modelKey,
		associationKey.String())
	if err != nil {
		return model_class.Association{}, errors.WithStack(err)
	}

	return association, nil
}

// AddAssociation adds a association to the database.
func AddAssociation(dbOrTx DbOrTx, modelKey string, association model_class.Association) (err error) {
	return AddAssociations(dbOrTx, modelKey, []model_class.Association{association})
}

// UpdateAssociation updates a association in the database.
func UpdateAssociation(dbOrTx DbOrTx, modelKey string, association model_class.Association) (err error) {

	// We may or may not have an association class.
	var associationClassKeyPtr *string
	if association.AssociationClassKey != nil {
		s := association.AssociationClassKey.String()
		associationClassKeyPtr = &s
	}

	// Update the data.
	_, err = dbExec(dbOrTx, `
		UPDATE
			association
		SET
			name                     = $3  ,
			details                  = $4  ,
			from_class_key           = $5  ,
			from_multiplicity_lower  = $6  ,
			from_multiplicity_higher = $7  ,
			to_class_key             = $8  ,
			to_multiplicity_lower    = $9  ,
			to_multiplicity_higher   = $10 ,
			association_class_key    = $11 ,
			uml_comment              = $12
		WHERE
			association_key = $2
		AND
			model_key = $1`,
		modelKey,
		association.Key.String(),
		association.Name,
		association.Details,
		association.FromClassKey.String(),
		association.FromMultiplicity.LowerBound,
		association.FromMultiplicity.HigherBound,
		association.ToClassKey.String(),
		association.ToMultiplicity.LowerBound,
		association.ToMultiplicity.HigherBound,
		associationClassKeyPtr,
		association.UmlComment)
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

// RemoveAssociation deletes a association from the database.
func RemoveAssociation(dbOrTx DbOrTx, modelKey string, associationKey identity.Key) (err error) {

	// Delete the data.
	_, err = dbExec(dbOrTx, `
		DELETE FROM
			association
		WHERE
			association_key = $2
		AND
			model_key = $1`,
		modelKey,
		associationKey.String())
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

// QueryAssociations loads all association from the database
func QueryAssociations(dbOrTx DbOrTx, modelKey string) (associations []model_class.Association, err error) {

	// Query the database.
	err = dbQuery(
		dbOrTx,
		func(scanner Scanner) (err error) {
			var association model_class.Association
			if err = scanAssociation(scanner, &association); err != nil {
				return errors.WithStack(err)
			}
			associations = append(associations, association)
			return nil
		},
		`SELECT
			association_key          ,
			name                     ,
			details                  ,
			from_class_key           ,
			from_multiplicity_lower  ,
			from_multiplicity_higher ,
			to_class_key             ,
			to_multiplicity_lower    ,
			to_multiplicity_higher   ,
			association_class_key    ,
			uml_comment
		FROM
			association
		WHERE
			model_key = $1
		ORDER BY association_key`,
		modelKey)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return associations, nil
}

// AddAssociations adds multiple associations to the database in a single insert.
func AddAssociations(dbOrTx DbOrTx, modelKey string, associations []model_class.Association) (err error) {
	if len(associations) == 0 {
		return nil
	}

	// Build the bulk insert query.
	query := `INSERT INTO association (model_key, association_key, name, details, from_class_key, from_multiplicity_lower, from_multiplicity_higher, to_class_key, to_multiplicity_lower, to_multiplicity_higher, association_class_key, uml_comment) VALUES `
	args := make([]interface{}, 0, len(associations)*12)
	for i, assoc := range associations {
		if i > 0 {
			query += ", "
		}
		base := i * 12
		query += fmt.Sprintf("($%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d)", base+1, base+2, base+3, base+4, base+5, base+6, base+7, base+8, base+9, base+10, base+11, base+12)

		// Handle optional association class key.
		var associationClassKeyPtr *string
		if assoc.AssociationClassKey != nil {
			s := assoc.AssociationClassKey.String()
			associationClassKeyPtr = &s
		}

		args = append(args, modelKey, assoc.Key.String(), assoc.Name, assoc.Details, assoc.FromClassKey.String(), assoc.FromMultiplicity.LowerBound, assoc.FromMultiplicity.HigherBound, assoc.ToClassKey.String(), assoc.ToMultiplicity.LowerBound, assoc.ToMultiplicity.HigherBound, associationClassKeyPtr, assoc.UmlComment)
	}

	_, err = dbExec(dbOrTx, query, args...)
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}
