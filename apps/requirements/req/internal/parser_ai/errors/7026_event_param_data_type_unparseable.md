# Event Parameter Data Type Unparseable (E7026)

The `data_type_rules` value for an event parameter could not be parsed.

## What Went Wrong

The `data_type_rules` field on a state machine event parameter contains a string that does not match any valid data type syntax.

## How to Fix

See **E5011** for the complete list of valid data type syntax and common mistakes.

Event parameters use the same `data_type_rules` syntax as class attributes. The `data_type_rules` field on an event parameter appears inside the `parameters` array of an event in `state_machine.json`.
