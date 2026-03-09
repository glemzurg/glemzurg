# Parameter Data Type Unparseable (E15005)

The `data_type_rules` value for a parameter could not be parsed.

## What Went Wrong

The `data_type_rules` field on an action or query parameter contains a string that does not match any valid data type syntax.

## Valid Data Type Syntax

See E5011 for the complete list of valid data type syntax.

## How to Fix

Check spelling and syntax of `data_type_rules`. Common mistakes:
- Using `string`/`integer`/`boolean` (these are TLA+ type specs, not data_type_rules)
- Missing `of` keyword in collections or enums
