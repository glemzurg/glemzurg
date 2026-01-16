package database

import (
	"database/sql"
	"flag"
	"fmt"
	"os"
	"testing"
)

// We only want to run the database tests occassionally because they're slower.
var _runDatabaseTests = flag.Bool("dbtests", false, "Run database tests only if this flag is set")

const (
	// A database used only during test.
	_TEST_DATABASE = "unit_test" // Don't use the database the application uses.
	_TEST_USER     = "postgres"  // Dev containers and CI use the default 'postgres' user.
	_TEST_PASSWORD = "postgres"  // dev containers and CI use the default 'postgres' password.
)

// t_ResetDatabase reset the database between unit tests. If not called the test will use the normal database.
// Return a database that is guaranteed to be a test database.
func t_ResetDatabase(t *testing.T) (db *sql.DB) {

	// Point to test database for any reset.
	// This also sets up the test to work just with the test database.

	_database = _TEST_DATABASE
	_user = _TEST_USER
	_password = _TEST_PASSWORD

	// Make a database of these settings.
	db, err := NewDb()
	if err != nil {
		t.Fatal(err)
	}

	// Get the two sql files to run.

	// Creating the schema.
	schemaContent, err := os.ReadFile("sql/schema.sql")
	if err != nil {
		t.Fatal(err)
	}
	schemaSql := string(schemaContent)

	// Dropping the schema.
	dropContent, err := os.ReadFile("sql/drop_schema.sql")
	if err != nil {
		t.Fatal(err)
	}
	dropSchemaSql := string(dropContent)

	// Drop the schema.
	if _, err = dbExec(db, dropSchemaSql); err != nil {
		// Don't stop. This may be the first time and there is no schema.
		// In addition this method, when custom types change will report a missing type.
		fmt.Println("reset database:", err)
	}

	// Add the schema.
	if _, err = dbExec(db, schemaSql); err != nil {
		t.Fatal(err)
	}

	// Return the test database.
	return db
}
