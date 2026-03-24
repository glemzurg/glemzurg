# Domain Structure Invalid (E21110)

A domain structural rule was violated.

## What Went Wrong

Domain structural rules enforce naming conventions for subdomains:

- A domain with a **single subdomain** must name it `default`
- A domain with **multiple subdomains** must not have one named `default`

## How to Fix

Check the error message for the specific rule violated, then rename the subdomain directory accordingly.

## Related Errors

- **E11030**: Single subdomain must be named "default"
- **E11031**: Multiple subdomains cannot include one named "default"
