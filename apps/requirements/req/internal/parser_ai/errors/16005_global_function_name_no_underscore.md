# Global Function Name No Underscore (E16005)

The global function name does not start with an underscore character.

## What Went Wrong

Global function names must start with an underscore (`_`) to distinguish them from other identifiers in the model. The parser found a name that does not follow this convention.

## Why Underscore Prefix?

The underscore prefix serves as a visual marker that identifies global functions in expressions throughout the model. When you see `_Max(x, y)` in a specification, you immediately know it refers to a globally-defined function rather than a local variable or state attribute.

## How to Fix

Add an underscore prefix to the function name:

```json
// Wrong:
{ "name": "Max", ... }

// Correct:
{ "name": "_Max", ... }
```

## Examples

| Invalid | Valid |
|---------|-------|
| `Max` | `_Max` |
| `SetOfValues` | `_SetOfValues` |
| `clamp` | `_Clamp` |
| `IS_VALID` | `_IsValid` |

## Related Errors

- **E16001**: Global function name field is missing entirely
- **E16002**: Global function name is whitespace only
