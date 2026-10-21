# Domain Association References Same Domain (E21117)

A domain association's `problem_domain_key` and `solution_domain_key` reference the same domain.

## What Went Wrong

Domain associations connect two different domains. The problem domain and solution domain must be different.

## How to Fix

Change one of the domain keys so they reference different domains. If both sides of the relationship are in the same domain, a domain association is not appropriate.

## Related Errors

- **E17001**: Domain association problem_key required
- **E17003**: Domain association solution_key required
