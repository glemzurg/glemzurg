# Event Duplicate Name (E7028)

Two events in the same state machine have different keys but the same `name` value. This produces duplicate map keys in YAML output.

## How to Fix

Rename one of the events so each has a unique `name` value within the state machine. Do not delete or remove either event to resolve this error — both events exist for a reason and the model needs them. Instead, adjust the `name` field in one of the conflicting event entries to distinguish them.
