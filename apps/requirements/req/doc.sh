#!/bin/bash
SCRIPT_PATH="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

# This documentation software.
# https://github.com/k1LoW/tbls
# go install github.com/k1LoW/tbls@latest

# UML documentation.
# https://github.com/plantuml/plantuml

cd $SCRIPT_PATH
[ $? -ne 0 ] && exit 1

psql "postgresql://postgres:postgres@localhost:5432/postgres" -f "$SCRIPT_PATH/internal/database/sql/drop_schema.sql"
[ $? -ne 0 ] && exit 1

psql "postgresql://postgres:postgres@localhost:5432/postgres" -f "$SCRIPT_PATH/internal/database/sql/schema.sql"
[ $? -ne 0 ] && exit 1

# Clear it out example uml.
rm -fr $SCRIPT_PATH/docs/dbdoc
[ $? -ne 0 ] && exit 1

# Documentation uses config: .tbls.yml
# Use force to rewrite without removing files, allows the files to update in a reader.
tbls doc --force
[ $? -ne 0 ] && exit 1

exit 0