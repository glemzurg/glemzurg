package database

import (
	"fmt"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_state"

	"github.com/lib/pq"
	"github.com/pkg/errors"
)

// Populate a golang struct from a database row.
func scanQuery(scanner Scanner, classKeyPtr *identity.Key, query *model_state.Query) (err error) {
	var classKeyStr string
	var queryKeyStr string

	if err = scanner.Scan(
		&classKeyStr,
		&queryKeyStr,
		&query.Name,
		&query.Details,
		pq.Array(&query.Requires),
		pq.Array(&query.Guarantees),
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

	// Parse the query key string into an identity.Key.
	query.Key, err = identity.ParseKey(queryKeyStr)
	if err != nil {
		return err
	}

	return nil
}

// LoadQuery loads a query from the database
func LoadQuery(dbOrTx DbOrTx, modelKey string, queryKey identity.Key) (classKey identity.Key, query model_state.Query, err error) {

	// Query the database.
	err = dbQueryRow(
		dbOrTx,
		func(scanner Scanner) (err error) {
			if err = scanQuery(scanner, &classKey, &query); err != nil {
				return err
			}
			return nil
		},
		`SELECT
			class_key  ,
			query_key ,
			name       ,
			details    ,
			requires   ,
			guarantees
		FROM
			query
		WHERE
			query_key = $2
		AND
			model_key = $1`,
		modelKey,
		queryKey.String())
	if err != nil {
		return identity.Key{}, model_state.Query{}, errors.WithStack(err)
	}

	return classKey, query, nil
}

// AddQuery adds a query to the database.
func AddQuery(dbOrTx DbOrTx, modelKey string, classKey identity.Key, query model_state.Query) (err error) {
	return AddQueries(dbOrTx, modelKey, map[identity.Key][]model_state.Query{
		classKey: {query},
	})
}

// UpdateQuery updates a query in the database.
func UpdateQuery(dbOrTx DbOrTx, modelKey string, classKey identity.Key, query model_state.Query) (err error) {

	// Update the data.
	_, err = dbExec(dbOrTx, `
		UPDATE
			query
		SET
			name       = $4 ,
			details    = $5 ,
			requires   = $6 ,
			guarantees = $7
		WHERE
			class_key = $2
		AND
			query_key = $3
		AND
			model_key = $1`,
		modelKey,
		classKey.String(),
		query.Key.String(),
		query.Name,
		query.Details,
		pq.Array(query.Requires),
		pq.Array(query.Guarantees))
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

// RemoveQuery deletes a query from the database.
func RemoveQuery(dbOrTx DbOrTx, modelKey string, classKey identity.Key, queryKey identity.Key) (err error) {

	// Delete the data.
	_, err = dbExec(dbOrTx, `
		DELETE FROM
			query
		WHERE
			class_key = $2
		AND
			query_key = $3
		AND
			model_key = $1`,
		modelKey,
		classKey.String(),
		queryKey.String())
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

// QueryQueries loads all query from the database
func QueryQueries(dbOrTx DbOrTx, modelKey string) (queries map[identity.Key][]model_state.Query, err error) {

	// Query the database.
	err = dbQuery(
		dbOrTx,
		func(scanner Scanner) (err error) {
			var classKey identity.Key
			var query model_state.Query
			if err = scanQuery(scanner, &classKey, &query); err != nil {
				return errors.WithStack(err)
			}
			if queries == nil {
				queries = map[identity.Key][]model_state.Query{}
			}
			classQueries := queries[classKey]
			classQueries = append(classQueries, query)
			queries[classKey] = classQueries
			return nil
		},
		`SELECT
			class_key  ,
			query_key ,
			name       ,
			details    ,
			requires   ,
			guarantees
		FROM
			query
		WHERE
			model_key = $1
		ORDER BY class_key, query_key`,
		modelKey)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return queries, nil
}

// AddQueries adds multiple queries to the database in a single insert.
func AddQueries(dbOrTx DbOrTx, modelKey string, queries map[identity.Key][]model_state.Query) (err error) {
	// Count total queries.
	count := 0
	for _, qrys := range queries {
		count += len(qrys)
	}
	if count == 0 {
		return nil
	}

	// Build the bulk insert query.
	sqlQuery := `INSERT INTO query (model_key, class_key, query_key, name, details, requires, guarantees) VALUES `
	args := make([]interface{}, 0, count*7)
	i := 0
	for classKey, queryList := range queries {
		for _, query := range queryList {
			if i > 0 {
				sqlQuery += ", "
			}
			base := i * 7
			sqlQuery += fmt.Sprintf("($%d, $%d, $%d, $%d, $%d, $%d, $%d)", base+1, base+2, base+3, base+4, base+5, base+6, base+7)
			args = append(args, modelKey, classKey.String(), query.Key.String(), query.Name, query.Details, pq.Array(query.Requires), pq.Array(query.Guarantees))
			i++
		}
	}

	_, err = dbExec(dbOrTx, sqlQuery, args...)
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}
