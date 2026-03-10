# Parameter Data Type Unparseable (E15005)

The `data_type_rules` value for an action or query parameter could not be parsed.

## What Went Wrong

The `data_type_rules` field on a parameter contains a string that does not match any valid data type syntax.

## How to Fix

See **E5011** for the complete list of valid data type syntax and common mistakes.

Parameters use the same `data_type_rules` syntax as class attributes. The `data_type_rules` field on a parameter appears inside the `parameters` array of an action or query JSON file.
