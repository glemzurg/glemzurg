# State Duplicate Name (E7027)

Two states in the same state machine have different keys but the same `name` value. This produces duplicate map keys in YAML output.

## How to Fix

Rename one of the states so each has a unique `name` value within the state machine.
