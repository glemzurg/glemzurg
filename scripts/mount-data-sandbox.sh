#!/usr/bin/env bash
# Bind-mount the host data directory onto the repo's data_sandbox/ placeholder
# so the dev container sees the data at /workspaces/glemzurg/data_sandbox.
#
# Run this on the HOST (not inside the dev container), from the repo root,
# before opening / starting the dev container. The dev container itself never
# needs this mount: it loads cleanly whether or not the mount is active, and
# data_sandbox/ is simply empty when no host data is bound.
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
