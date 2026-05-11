# glemzurg

Tools for modeling and generating software requirements and documentation.

## Data Sandbox Directory

The tools in this repo read from and write to `data_sandbox/` at the repo root (for example, requirements documents to process). The directory is committed as an empty placeholder so the dev container always starts cleanly. Working data is supplied per-host by bind-mounting an arbitrary host directory over the placeholder; the contents of `data_sandbox/` (other than `.gitkeep`) are git-ignored.

The dev container itself does not declare any optional mount, so it loads on every host whether or not data is provided.

### Linux / macOS host

Set the host environment variable `GLEMZURG_DATA_PATH` to the directory containing your data, then run the helper from the repo root **on the host** (not inside the dev container) before starting the container:

```bash
export GLEMZURG_DATA_PATH=/absolute/path/to/your/data
./scripts/mount-data-sandbox.sh
```

The script bind-mounts `$GLEMZURG_DATA_PATH` onto `data_sandbox/` if the path exists and is not already mounted, and otherwise leaves the placeholder empty. To detach later:

```bash
sudo umount data_sandbox
```

### Windows host

Bind mounts are a Linux kernel feature. On Windows, run the helper above from inside WSL against a WSL path, or place your data directly in `data_sandbox/` on the host.

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
