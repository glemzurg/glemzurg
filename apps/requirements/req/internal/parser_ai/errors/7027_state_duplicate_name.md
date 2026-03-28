# State Duplicate Name (E7027)

Two states in the same state machine have different keys but the same `name` value. This produces duplicate map keys in YAML output.

## How to Fix

Rename one of the states so each has a unique `name` value within the state machine. Do not delete or remove either state to resolve this error — both states exist for a reason and the model needs them. Instead, adjust the `name` field in one of the conflicting state entries to distinguish them.
