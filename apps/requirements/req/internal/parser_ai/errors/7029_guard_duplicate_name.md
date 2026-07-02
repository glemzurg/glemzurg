# Guard Duplicate Name (E7029)

Two guards in the same state machine have different keys but the same `name` value. This produces duplicate map keys in YAML output.

## How to Fix

Rename one of the guards so each has a unique `name` value within the state machine. Do not delete or remove either guard to resolve this error — both guards exist for a reason and the model needs them. Instead, adjust the `name` field in one of the conflicting guard entries to distinguish them.
