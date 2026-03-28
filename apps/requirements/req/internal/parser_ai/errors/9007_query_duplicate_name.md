# Query Duplicate Name (E9007)

Two query files in the same class have different keys but the same `name` value. This produces duplicate map keys in YAML output.

## How to Fix

Rename one of the queries so each has a unique `name` value within the class.
