# Source Model Validation Failed (E21008)

The source model failed validation before conversion to input format.

## What Went Wrong

When converting an internal model back to the input format (for serialization or export), the source model must first pass validation. The model has structural or consistency issues that prevent conversion.

## Common Causes

This error typically indicates that the model was programmatically constructed with invalid data. The validation rules are the same as E21002.

## How to Fix

Ensure the model being converted passes all validation rules:

- All required names and types are set
- Actor types are "person" or "system"
- Use case levels are "sky", "sea", or "mud"
- Logic notations are "tla_plus" (if specified)
- Subdomain naming rules are followed
- All cross-references point to existing entities

## Related Errors

- **E21002**: Model validation failed after conversion to model
