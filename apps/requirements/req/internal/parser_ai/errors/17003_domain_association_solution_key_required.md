# Domain Association Solution Key Required (E17003)

A domain association is missing the required `solution_domain_key` field.

## What Went Wrong

The parser found a domain association object that does not contain a `solution_domain_key` property. Every domain association must specify which domain is the solution domain in the relationship.

## Correct Format

```json
{
    "problem_domain_key": "billing",
    "solution_domain_key": "payment_processing"
}
```

## How to Fix

Add the `solution_domain_key` field with the key of the domain that satisfies the requirements:

```json
{
    "problem_domain_key": "billing",
    "solution_domain_key": "payment_processing"
}
```

## Related Errors

- **E17001**: Problem domain key is missing
- **E17004**: Solution domain key is present but empty or whitespace
