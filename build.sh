#!/bin/bash
# Build every app under apps/requirements and apps/code.
# Each app keeps its own build.sh. Go apps also keep their own go.mod, because
# a tree can hold apps written in other languages.

set -e

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Go refuses to fetch a module into a download directory the user cannot write.
# The dev container has some of those directories. A private cache is used then.
mod_download="$(go env GOMODCACHE)/cache/download"
if [ -d "$mod_download" ] && find "$mod_download" -type d ! -writable -print -quit | grep -q .; then
	export GOMODCACHE="${TMPDIR:-/tmp}/glemzurg-modcache"
	export GOCACHE="${TMPDIR:-/tmp}/glemzurg-gocache"
	mkdir -p "$GOMODCACHE" "$GOCACHE"
fi

run_tree() {
	local tree="$1"
	local dir
	local found=0
	for dir in "$ROOT/apps/$tree"/*; do
		[ -d "$dir" ] || continue
		if [ -f "$dir/build.sh" ]; then
			found=1
			echo -e "\n======== $tree/$(basename "$dir") ========\n"
			(cd "$dir" && bash ./build.sh)
		fi
	done
	if [ "$found" -eq 0 ]; then
		echo "no build.sh under apps/$tree" >&2
		exit 1
	fi
}

run_tree requirements
run_tree code
