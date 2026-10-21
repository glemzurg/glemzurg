# Domain Association Problem Key Required (E17001)

A domain association is missing the required `problem_domain_key` field.

## What Went Wrong

The parser found a domain association object that does not contain a `problem_domain_key` property. Every domain association must specify which domain is the problem domain in the relationship.

## Where Domain Associations Appear

Domain associations are defined at the model level in `domain_associations/` directory:
```
model/
  domain_associations/
    billing--payment_processing.da.json
```

## Correct Format

```json
{
    "problem_domain_key": "billing",
    "solution_domain_key": "payment_processing"
}
```

## How to Fix

Add the `problem_domain_key` field with the key of the domain that enforces requirements:

```json
{
    "problem_domain_key": "your_problem_domain",
    "solution_domain_key": "your_solution_domain"
}
```

## Related Errors

- **E17002**: Problem domain key is present but empty or whitespace
- **E17003**: Solution domain key is missing
