# Domain Association Invalid JSON (E17005)

The domain association file contains invalid JSON that cannot be parsed.

## What Went Wrong

The file content is not valid JSON. This could be due to:
- Missing or extra commas
- Unquoted property names
- Single quotes instead of double quotes
- Missing closing braces or brackets
- Trailing commas after the last property
- Comments (JSON does not support comments)

## Valid Domain Association Template

```json
{
    "problem_domain_key": "billing",
    "solution_domain_key": "payment_processing",
    "uml_comment": "optional comment"
}
```

## Related Errors

- **E17006**: JSON is valid but does not match the domain association schema
