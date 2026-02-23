# Global Function Name Required (E16001)

A global function object is missing the required `name` field.

## What Went Wrong

The parser found a global function object that does not contain a `name` property. Every global function must have a name that starts with an underscore (e.g., `_Max`, `_SetOfValues`).

## Where Global Functions Appear

Global functions are defined at the model level in `global_functions/` directory:
```
model/
  global_functions/
    _max.gf.json          <- Global function file
    _set_of_values.gf.json
```

## Correct Format

A global function must have a `name`, and a `logic` object with a `description`:

```json
{
    "name": "_Max",
    "parameters": ["x", "y"],
    "logic": {
        "description": "Returns the maximum of two values",
        "notation": "tla_plus",
        "specification": "IF x > y THEN x ELSE y"
    }
}
```

## How to Fix

Add the `name` field with a descriptive name starting with underscore:

```json
{
    "name": "_YourFunctionName",
    "logic": {
        "description": "What this function does"
    }
}
```

## Naming Conventions

- Global function names **must** start with an underscore (`_`)
- Use PascalCase after the underscore (e.g., `_Max`, `_SetOfValues`, `_Clamp`)
- The name should describe what the function computes or represents

## Related Errors

- **E16002**: Global function name is present but empty or whitespace
- **E16005**: Global function name does not start with underscore
