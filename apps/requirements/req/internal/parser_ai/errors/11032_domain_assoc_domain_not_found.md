# ErrTreeDomainAssocDomainNotFound (11032)

Description

- This error indicates that a domain association file (`*.domain_assoc.json`) references a domain key that does not exist in the model's `domains/` directory.

Cause

- The `problem_domain_key` or `solution_domain_key` field in a domain association points to a domain that is not present in `model.Domains` (i.e. there's no corresponding `domains/<key>/domain.json`).

Why this matters

- Domain associations express relationships across domains. If the referenced domain does not exist, the association cannot be validated or applied and AI guidance relying on that relationship will be incorrect.

How to fix

- Ensure the referenced domain directories exist under `domains/` and include a valid `domain.json` file.
- Fix typos in the `problem_domain_key` or `solution_domain_key` fields in the domain association file.

Example

- Invalid: `problem_domain_key: payments` when only `domains/orders` exists.
- Fix: rename to `problem_domain_key: orders` or create the `domains/payments/domain.json` file.
