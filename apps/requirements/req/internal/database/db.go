package database

import (
	"database/sql"
	"fmt"
	"log"
	"sync"

	_ "github.com/lib/pq"
	"github.com/pkg/errors"
)

const (
	_DRIVER   = "postgres"
	_HOST     = "localhost"
	_PORT     = 5432
	_DATABASE = "modeling"
	_USER     = "modeling"
	_PASSWORD = "modeling"
)

const (
	_ROLLBACK_NOT_NECESSARY = "sql: transaction has already been committed or rolled back"
)

var _db *sql.DB // One handle for the database.
var _dbMutex sync.Mutex

// Use values that tests will override.
var (
	_database = _DATABASE
	_user     = _USER
	_password = _PASSWORD
)

// Either a database or a transaction.
type DbOrTx interface {
	Exec(query string, args ...any) (sql.Result, error)
	Query(query string, args ...any) (*sql.Rows, error)
	QueryRow(query string, args ...any) *sql.Row
}

// NewDb gives a consistent connection to all code in the package.
func NewDb() (db *sql.DB, err error) {

	// Avoid race conditions in this code.
	_dbMutex.Lock()
	defer _dbMutex.Unlock()

	// We may need to instantiate the connection.
	if _db == nil {

		// Instantiate the single database connection.
		connStr := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", _HOST, _PORT, _user, _password, _database)
		if _db, err = sql.Open(_DRIVER, connStr); err != nil {
			return nil, errors.WithStack(err)
		}
	}

	return _db, nil
}

// dbExec processes a single sql statement.
func dbExec(dbOrTx DbOrTx, query string, args ...interface{}) (result sql.Result, err error) {

	if result, err = dbOrTx.Exec(query, args...); err != nil {
		return nil, errors.WithStack(err)
	}

	return result, nil
}

// Scanner is any object that can scan a row of a SQL query result (*sql.Row and *sql.Rows).
type Scanner interface {
	Scan(dest ...interface{}) (err error)
}

// RowHandleFunc is a method to run for each row of a query.
type RowHandleFunc func(scanner Scanner) (err error)

// dbQuery runs a multi-row return sql statement and handles each row of the results with the method passed in.
func dbQuery(dbOrTx DbOrTx, rowHandleFunc RowHandleFunc, query string, args ...interface{}) (err error) {

	// Make the query.
	rows, err := dbOrTx.Query(query, args...)
	if err != nil {
		return errors.WithStack(err)
	}
	defer rows.Close()

	// Handle each row.
	for rows.Next() {
		if err = rowHandleFunc(rows); err != nil {
			return errors.WithStack(err)
		}
	}

	return nil
}

// dbQueryRow runs a single-row return sql query statement.
func dbQueryRow(dbOrTx DbOrTx, rowHandleFunc RowHandleFunc, query string, args ...interface{}) (err error) {

	// Query the row.
	row := dbOrTx.QueryRow(query, args...)

	// Handle the row.
	if err = rowHandleFunc(row); err != nil {
		return errors.WithStack(err)
	}

	return nil
}

// dbTransactionFunc is the signature for a lambda passed into a transaction.
type dbTransactionFunc func(tx *sql.Tx) (err error)

// dbTransaction processes the method passed in all in the context of a single SQL transaction.
func dbTransaction(db *sql.DB, transactionFunc dbTransactionFunc) (err error) {

	// Start a transaction.
	tx, err := db.Begin()
	if err != nil {
		return errors.WithStack(err)
	}
	defer func() {
		rollbackErr := tx.Rollback() // If exit in any way before the commit, rollback.
		if rollbackErr != nil {
			// If this is just that the defer happened after a commit, no need to report.
			if rollbackErr.Error() != _ROLLBACK_NOT_NECESSARY {
				log.Println("rollback error:", rollbackErr.Error())
			}
		}
	}()

	// Do the work.
	if err = transactionFunc(tx); err != nil {
		return errors.WithStack(err)
	}

	// If we made it here, we are ready to commit.
	if err = tx.Commit(); err != nil {
		return errors.WithStack(err)
	}

	return nil
}
