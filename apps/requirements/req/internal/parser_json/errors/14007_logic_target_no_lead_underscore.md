# Logic Target No Leading Underscore (E14007)

The `target` field value starts with an underscore, which is not allowed.

## How to Fix

Target must be a valid attribute key (lowercase snake_case, no leading underscore):

```json
{
    "target": "status"
}
```
