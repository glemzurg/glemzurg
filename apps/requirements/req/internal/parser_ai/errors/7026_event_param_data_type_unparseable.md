# Event Parameter Data Type Unparseable (E7026)

The `data_type_rules` value for an event parameter could not be parsed.

## What Went Wrong

The `data_type_rules` field on a state machine event parameter contains a string that does not match any valid data type syntax.

## Valid Data Type Syntax

See E5011 for the complete list of valid data type syntax.

## How to Fix

Check spelling and syntax of `data_type_rules`. Common mistakes:
- Using `integer`/`float`/`number` — there are no primitive numeric types; use a **span** instead, e.g. `[0 .. unconstrained] at 1 count` for integers or `[0 .. unconstrained] at 0.01 dollars` for decimals
- Using `string`/`boolean` (these are TLA+ type specs, not data_type_rules)
- Missing `of` keyword in collections or enums
