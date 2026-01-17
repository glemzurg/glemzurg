package database

import (
	"fmt"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_class"

	"github.com/pkg/errors"
)

// Populate a golang struct from a database row.
func scanClass(scanner Scanner, subdomainKeyPtr *identity.Key, class *model_class.Class) (err error) {
	var subdomainKeyStr string
	var classKeyStr string
	var actorKeyPtr, superclassOfKeyPtr, subclassOfKeyPtr *string

	if err = scanner.Scan(
		&subdomainKeyStr,
		&classKeyStr,
		&class.Name,
		&class.Details,
		&actorKeyPtr,
		&superclassOfKeyPtr,
		&subclassOfKeyPtr,
		&class.UmlComment,
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

	// Parse the class key string into an identity.Key.
	class.Key, err = identity.ParseKey(classKeyStr)
	if err != nil {
		return err
	}

	// Parse optional key pointers.
	if actorKeyPtr != nil {
		actorKey, err := identity.ParseKey(*actorKeyPtr)
		if err != nil {
			return err
		}
		class.ActorKey = &actorKey
	}
	if superclassOfKeyPtr != nil {
		superclassOfKey, err := identity.ParseKey(*superclassOfKeyPtr)
		if err != nil {
			return err
		}
		class.SuperclassOfKey = &superclassOfKey
	}
	if subclassOfKeyPtr != nil {
		subclassOfKey, err := identity.ParseKey(*subclassOfKeyPtr)
		if err != nil {
			return err
		}
		class.SubclassOfKey = &subclassOfKey
	}

	return nil
}

// LoadClass loads a class from the database
func LoadClass(dbOrTx DbOrTx, modelKey string, classKey identity.Key) (subdomainKey identity.Key, class model_class.Class, err error) {

	// Query the database.
	err = dbQueryRow(
		dbOrTx,
		func(scanner Scanner) (err error) {
			if err = scanClass(scanner, &subdomainKey, &class); err != nil {
				return err
			}
			return nil
		},
		`SELECT
			subdomain_key     ,
			class_key         ,
			name              ,
			details           ,
			actor_key         ,
			superclass_of_key ,
			subclass_of_key   ,
			uml_comment
		FROM
			class
		WHERE
			class_key = $2
		AND
			model_key = $1`,
		modelKey,
		classKey.String())
	if err != nil {
		return identity.Key{}, model_class.Class{}, errors.WithStack(err)
	}

	return subdomainKey, class, nil
}

// AddClass adds a class to the database.
func AddClass(dbOrTx DbOrTx, modelKey string, subdomainKey identity.Key, class model_class.Class) (err error) {

	// We may or may not have optional key pointers.
	var actorKeyPtr *string
	if class.ActorKey != nil {
		actorKeyStr := class.ActorKey.String()
		actorKeyPtr = &actorKeyStr
	}
	var superclassOfKeyPtr *string
	if class.SuperclassOfKey != nil {
		superclassOfKeyStr := class.SuperclassOfKey.String()
		superclassOfKeyPtr = &superclassOfKeyStr
	}
	var subclassOfKeyPtr *string
	if class.SubclassOfKey != nil {
		subclassOfKeyStr := class.SubclassOfKey.String()
		subclassOfKeyPtr = &subclassOfKeyStr
	}

	// Add the data.
	_, err = dbExec(dbOrTx, `
		INSERT INTO class
			(
				model_key         ,
				subdomain_key     ,
				class_key         ,
				name              ,
				details           ,
				actor_key         ,
				superclass_of_key ,
				subclass_of_key   ,
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
		subdomainKey.String(),
		class.Key.String(),
		class.Name,
		class.Details,
		actorKeyPtr,
		superclassOfKeyPtr,
		subclassOfKeyPtr,
		class.UmlComment)
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

// UpdateClass updates a class in the database.
func UpdateClass(dbOrTx DbOrTx, modelKey string, class model_class.Class) (err error) {

	// We may or may not have optional key pointers.
	var actorKeyPtr *string
	if class.ActorKey != nil {
		actorKeyStr := class.ActorKey.String()
		actorKeyPtr = &actorKeyStr
	}
	var superclassOfKeyPtr *string
	if class.SuperclassOfKey != nil {
		superclassOfKeyStr := class.SuperclassOfKey.String()
		superclassOfKeyPtr = &superclassOfKeyStr
	}
	var subclassOfKeyPtr *string
	if class.SubclassOfKey != nil {
		subclassOfKeyStr := class.SubclassOfKey.String()
		subclassOfKeyPtr = &subclassOfKeyStr
	}

	// Update the data.
	_, err = dbExec(dbOrTx, `
		UPDATE
			class
		SET
			name              = $3 ,
			details           = $4 ,
			actor_key         = $5 ,
			superclass_of_key = $6 ,
			subclass_of_key   = $7 ,
			uml_comment       = $8
		WHERE
			model_key = $1
		AND
			class_key = $2`,
		modelKey,
		class.Key.String(),
		class.Name,
		class.Details,
		actorKeyPtr,
		superclassOfKeyPtr,
		subclassOfKeyPtr,
		class.UmlComment)
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

// RemoveClass deletes a class from the database.
func RemoveClass(dbOrTx DbOrTx, modelKey string, classKey identity.Key) (err error) {

	// Delete the data.
	_, err = dbExec(dbOrTx, `
		DELETE FROM
			class
		WHERE
			model_key = $1
		AND
			class_key = $2`,
		modelKey,
		classKey.String())
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

// QueryClasses loads all classes from the database
func QueryClasses(dbOrTx DbOrTx, modelKey string) (classes map[identity.Key][]model_class.Class, err error) {

	// Query the database.
	err = dbQuery(
		dbOrTx,
		func(scanner Scanner) (err error) {
			var subdomainKey identity.Key
			var class model_class.Class
			if err = scanClass(scanner, &subdomainKey, &class); err != nil {
				return errors.WithStack(err)
			}
			if classes == nil {
				classes = map[identity.Key][]model_class.Class{}
			}
			subdomainClasses := classes[subdomainKey]
			subdomainClasses = append(subdomainClasses, class)
			classes[subdomainKey] = subdomainClasses

			return nil
		},
		`SELECT
			subdomain_key     ,
			class_key         ,
			name              ,
			details           ,
			actor_key         ,
			superclass_of_key ,
			subclass_of_key   ,
			uml_comment
		FROM
			class
		WHERE
			model_key = $1
		ORDER BY subdomain_key, class_key`,
		modelKey)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return classes, nil
}

// AddClasses adds multiple classes to the database in a single insert.
func AddClasses(dbOrTx DbOrTx, modelKey string, classes map[identity.Key][]model_class.Class) (err error) {
	// Count total classes.
	count := 0
	for _, cls := range classes {
		count += len(cls)
	}
	if count == 0 {
		return nil
	}

	// Build the bulk insert query.
	query := `INSERT INTO class (model_key, subdomain_key, class_key, name, details, actor_key, superclass_of_key, subclass_of_key, uml_comment) VALUES `
	args := make([]interface{}, 0, count*9)
	i := 0
	for subdomainKey, classList := range classes {
		for _, class := range classList {
			if i > 0 {
				query += ", "
			}
			base := i * 9
			query += fmt.Sprintf("($%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d)", base+1, base+2, base+3, base+4, base+5, base+6, base+7, base+8, base+9)

			// Handle optional key pointers.
			var actorKeyPtr, superclassOfKeyPtr, subclassOfKeyPtr *string
			if class.ActorKey != nil {
				s := class.ActorKey.String()
				actorKeyPtr = &s
			}
			if class.SuperclassOfKey != nil {
				s := class.SuperclassOfKey.String()
				superclassOfKeyPtr = &s
			}
			if class.SubclassOfKey != nil {
				s := class.SubclassOfKey.String()
				subclassOfKeyPtr = &s
			}

			args = append(args, modelKey, subdomainKey.String(), class.Key.String(), class.Name, class.Details, actorKeyPtr, superclassOfKeyPtr, subclassOfKeyPtr, class.UmlComment)
			i++
		}
	}

	_, err = dbExec(dbOrTx, query, args...)
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}
