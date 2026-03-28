# Event Duplicate Name (E7028)

Two events in the same state machine have different keys but the same `name` value. This produces duplicate map keys in YAML output.

## How to Fix

Rename one of the events so each has a unique `name` value within the state machine.
