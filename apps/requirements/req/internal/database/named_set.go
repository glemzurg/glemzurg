package database

import (
	"fmt"
	"strings"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_logic"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_logic/logic_spec"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"

	"github.com/pkg/errors"
)

// Populate a golang struct from a database row.
func scanNamedSet(scanner Scanner, ns *model_logic.NamedSet) (err error) {
	var keyStr string
	var notation string
	var specification string
	var typeSpecNotation *string
	var typeSpecSpecification *string

	if err = scanner.Scan(
		&keyStr,
		&ns.Name,
		&ns.Description,
		&notation,
		&specification,
		&typeSpecNotation,
		&typeSpecSpecification,
	); err != nil {
		if err.Error() == _POSTGRES_NOT_FOUND {
			err = ErrNotFound
		}
		return err // Do not wrap in stack here. It will be wrapped in the database calls.
	}

	// Parse the key string into an identity.Key.
	ns.Key, err = identity.ParseKey(keyStr)
	if err != nil {
		return err
	}

	// Construct ExpressionSpec via constructor (nil parseFunc — parsing happens at higher layers).
	ns.Spec, err = logic_spec.NewExpressionSpec(notation, specification, nil)
	if err != nil {
		return err
	}

	// Reconstitute TypeSpec if present.
	if typeSpecNotation != nil && *typeSpecNotation != "" {
		spec := ""
		if typeSpecSpecification != nil {
			spec = *typeSpecSpecification
		}
		ts, err := logic_spec.NewTypeSpec(*typeSpecNotation, spec, nil)
		if err != nil {
			return err
		}
		ns.TypeSpec = &ts
	}

	return nil
}

// LoadNamedSet loads a named set from the database.
func LoadNamedSet(dbOrTx DbOrTx, modelKey string, setKey identity.Key) (ns model_logic.NamedSet, err error) {
	err = dbQueryRow(
		dbOrTx,
		func(scanner Scanner) (err error) {
			if err = scanNamedSet(scanner, &ns); err != nil {
				return err
			}
			return nil
		},
		`SELECT
			set_key                ,
			name                   ,
			description            ,
			notation               ,
			specification          ,
			type_spec_notation     ,
			type_spec_specification
		FROM
			named_set
		WHERE
			model_key = $1
		AND
			set_key = $2`,
		modelKey,
		setKey.String())
	if err != nil {
		return model_logic.NamedSet{}, errors.WithStack(err)
	}

	return ns, nil
}

// AddNamedSet adds a named set row to the database.
func AddNamedSet(dbOrTx DbOrTx, modelKey string, ns model_logic.NamedSet) (err error) {
	return AddNamedSets(dbOrTx, modelKey, []model_logic.NamedSet{ns})
}

// RemoveNamedSet deletes a named set row from the database.
func RemoveNamedSet(dbOrTx DbOrTx, modelKey string, setKey identity.Key) (err error) {
	err = dbExec(dbOrTx, `
		DELETE FROM
			named_set
		WHERE
			model_key = $1
		AND
			set_key = $2`,
		modelKey,
		setKey.String())
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

// QueryNamedSets loads all named sets from the database for a given model.
func QueryNamedSets(dbOrTx DbOrTx, modelKey string) (nss []model_logic.NamedSet, err error) {
	err = dbQuery(
		dbOrTx,
		func(scanner Scanner) (err error) {
			var ns model_logic.NamedSet
			if err = scanNamedSet(scanner, &ns); err != nil {
				return errors.WithStack(err)
			}
			nss = append(nss, ns)
			return nil
		},
		`SELECT
			set_key                ,
			name                   ,
			description            ,
			notation               ,
			specification          ,
			type_spec_notation     ,
			type_spec_specification
		FROM
			named_set
		WHERE
			model_key = $1
		ORDER BY set_key`,
		modelKey)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return nss, nil
}

// AddNamedSets adds multiple named set rows to the database in a single insert.
func AddNamedSets(dbOrTx DbOrTx, modelKey string, nss []model_logic.NamedSet) (err error) {
	if len(nss) == 0 {
		return nil
	}

	var qb strings.Builder
	qb.WriteString(`INSERT INTO named_set (model_key, set_key, name, description, notation, specification, type_spec_notation, type_spec_specification) VALUES `)
	args := make([]any, 0, len(nss)*8)
	for i, ns := range nss {
		if i > 0 {
			qb.WriteString(", ")
		}
		base := i * 8
		qb.WriteString(fmt.Sprintf("($%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d)", base+1, base+2, base+3, base+4, base+5, base+6, base+7, base+8))

		var tsNotation *string
		var tsSpecification *string
		if ns.TypeSpec != nil {
			tsNotation = &ns.TypeSpec.Notation
			tsSpecification = &ns.TypeSpec.Specification
		}

		args = append(args,
			modelKey,
			ns.Key.String(),
			ns.Name,
			ns.Description,
			ns.Spec.Notation,
			ns.Spec.Specification,
			tsNotation,
			tsSpecification)
	}

	err = dbExec(dbOrTx, qb.String(), args...)
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}
