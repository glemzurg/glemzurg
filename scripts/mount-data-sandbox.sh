#!/usr/bin/env bash
# Bind-mount the host data directory onto the repo's data_sandbox/ placeholder
# so the dev container sees the data at /workspaces/glemzurg/data_sandbox.
#
# Invoked automatically by the dev container's initializeCommand on the HOST
# before each build/start, so the host mount is always in place when Docker
# creates its own bind mount of the workspace. The script is a no-op when
# GLEMZURG_DATA_PATH is unset, missing, or already mounted, so the dev
# container loads cleanly on every host whether or not data is provided.
#
# The source path comes from $GLEMZURG_DATA_PATH so each host can point to its
# own data location without changing repo files.

set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
TARGET="${REPO_ROOT}/data_sandbox"
SOURCE="${GLEMZURG_DATA_PATH:-}"

if [ -z "${SOURCE}" ]; then
	echo "GLEMZURG_DATA_PATH is not set on the host; leaving data_sandbox/ empty."
	exit 0
fi

if [ ! -d "${SOURCE}" ]; then
	echo "GLEMZURG_DATA_PATH (${SOURCE}) does not exist; leaving data_sandbox/ empty."
	exit 0
fi

if mountpoint -q "${TARGET}"; then
	echo "data_sandbox/ is already a mount point; nothing to do."
	exit 0
fi

sudo mount --bind "${SOURCE}" "${TARGET}"
echo "Bound ${SOURCE} -> ${TARGET}"
