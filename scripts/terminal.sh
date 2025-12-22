
# !/bin/bash

# Use docker to open a terminal in the devcontainer, start at the monorepo root.
docker exec -it -w /workspaces/glemzurg \
  $(docker ps --filter "label=devcontainer.local_folder" --format "{{.Names}}" | head -1) \
  bash -c "cd /workspaces/glemzurg && exec bash"