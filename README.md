# glemzurg

Tools for modeling and generating software requirements and documentation.

## Data Sandbox Directory

The tools in this repo read from and write to `data_sandbox/` at the repo root (for example, requirements documents to process). The directory is committed as an empty placeholder so the dev container always starts cleanly. Working data is supplied per-host by bind-mounting an arbitrary host directory over the placeholder; the contents of `data_sandbox/` (other than `.gitkeep`) are git-ignored.

The dev container's `initializeCommand` runs `scripts/mount-data-sandbox.sh` on the host before each build/start. The script bind-mounts `$GLEMZURG_DATA_PATH` onto `data_sandbox/` when the variable is set and the path exists, and is a no-op otherwise — so the container loads on every host whether or not data is provided. Doing the mount via `initializeCommand` (rather than asking the user to remember a separate step) ensures the host bind mount is always in place before Docker creates the workspace mount, which is required for writes inside the container to propagate to the host data set.

### Linux / macOS host

Set the host environment variable `GLEMZURG_DATA_PATH` to the directory containing your data in the shell from which you launch VS Code:

```bash
export GLEMZURG_DATA_PATH=/absolute/path/to/your/data
```

Then open or rebuild the dev container ("Dev Containers: Rebuild Container"). The `initializeCommand` will run the bind-mount script and may prompt for `sudo` on the host. To detach later:

```bash
sudo umount data_sandbox
```

To go back to the in-repo placeholder, unset the variable, unmount as above, and rebuild the container.

### Windows host

Bind mounts are a Linux kernel feature. On Windows, set `GLEMZURG_DATA_PATH` and open the workspace via the WSL remote so `initializeCommand` runs the script inside WSL against a WSL path; or place your data directly in `data_sandbox/` on the host and leave `GLEMZURG_DATA_PATH` unset.

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
