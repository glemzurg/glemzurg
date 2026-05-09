# glemzurg

Tools for modeling and generating software requirements and documentation.

## Data Sandbox Directory

The tools in this repo read from and write to `data_sandbox/` at the repo root (for example, requirements documents to process). Rather than committing your working data into the repo, point `data_sandbox/` at a directory of your choice on the host using a symbolic link.

On the **host** (not inside the dev container), from the repo root:

```bash
# Remove the placeholder directory if present (it should be empty).
rmdir data_sandbox

# Create a symbolic link to wherever you keep your data on the host.
ln -s /absolute/path/to/your/data data_sandbox
```

On Windows (PowerShell, run as Administrator or with Developer Mode enabled):

```powershell
Remove-Item data_sandbox
New-Item -ItemType SymbolicLink -Path data_sandbox -Target "C:\absolute\path\to\your\data"
```

The symlink is followed transparently by the dev container because the workspace folder is bind-mounted from the host. The `data_sandbox` symlink is git-ignored so it does not get committed.

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
