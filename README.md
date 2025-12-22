# glemzurg

Tools for modeling and generating software requirements and documentation.

## Environment Variable: GLEMZURG_DATA_PATH

This environment variable must be set on the host for the dev container to launch properly. It specifies the path to the data directory used by the application. If missing you will get a cryptic error about a malformed devcontainer.json.

## Go Built Files

Compiled Go binaries are placed in the `/go/bin` directory within the development environment.

## Database Setup

### PostgreSQL Configuration

In the dev container, PostgreSQL is configured with:
- User: `postgres`
- Password: `postgres`
- Port: `5432`
- Default Database: `postgres`
- Unit Test Database: `unit_test` (same credentials, different database name)

### Main Database for Modeling

To create the default database the first time, execute:

```
./apps/requirements/req/doc.sh
```

This also generates documentation for the current database.

**Each time this is run it will erase the data in the database.**

### Unit Tests Database

To create the unit test database the first time, execute:
```
psql -U postgres -f apps/requirements/req/internal/database/sql/reset_unit_test.sql
```
