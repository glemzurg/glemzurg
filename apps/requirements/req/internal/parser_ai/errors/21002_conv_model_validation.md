# Conversion Model Validation Failed (E21002)

The model failed validation after conversion from parsed input.

## What Went Wrong

After converting all parsed JSON input into the internal model representation, the final model validation check failed. This means the combined model has structural or consistency issues that were not caught by individual entity validation.

## Common Causes

- **Missing required fields**: A name, type, or level field is empty in the converted model
- **Invalid actor type**: Actor type must be "person" or "system"
- **Invalid use case level**: Use case level must be "sky", "sea", or "mud"
- **Invalid share type**: Use case shared share_type must be "include" or "extend"
- **Invalid logic notation**: Logic notation must be "tla_plus" (if specified)
- **Subdomain naming violation**: Single subdomain must be "default"; multiple subdomains cannot include "default"
- **Empty required collections**: Certain entities require non-empty collections (e.g., transitions need from or to state)

## How to Fix

Review the error message for the specific validation failure. The underlying model validation describes exactly which field or constraint was violated. Fix the corresponding JSON input file.

## Validation Rules

The model enforces these rules after conversion:

| Entity | Rule |
|--------|------|
| Actor | Type must be "person" or "system" |
| Use Case | Level must be "sky", "sea", or "mud" |
| Use Case Shared | share_type must be "include" or "extend" |
| Logic | Notation must be "tla_plus" (if provided) |
| Logic | Description is required |
| Global Function | Name must start with underscore `_` |
| Domain | Single subdomain must be named "default" |
| Domain | Multiple subdomains cannot include one named "default" |
| Transition | Must have at least one of from_state_key or to_state_key |
| Multiplicity | Upper bound must be >= lower bound |

## Related Errors

- **E2004**: Actor type invalid
- **E18006**: Use case level invalid
- **E14001**: Logic description required
- **E16005**: Global function name must start with underscore
