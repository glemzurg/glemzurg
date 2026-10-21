# Domain Association Solution Key Empty (E17004)

A domain association has a `solution_domain_key` field that contains only whitespace characters.

## What Went Wrong

The parser found a domain association with a `solution_domain_key` that is present but consists entirely of spaces, tabs, or other whitespace. The key must contain visible characters matching an existing domain key.

## Correct Format

```json
{
    "problem_domain_key": "billing",
    "solution_domain_key": "payment_processing"
}
```

## How to Fix

Replace the whitespace-only key with a valid domain key:

```json
{
    "problem_domain_key": "billing",
    "solution_domain_key": "payment_processing"
}
```

## Related Errors

- **E17002**: Problem domain key is empty or whitespace
- **E17003**: Solution domain key field is missing entirely
