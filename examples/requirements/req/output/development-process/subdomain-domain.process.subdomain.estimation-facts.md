[⇦ Development Process](model.md) / [Process](domain-domain.process.md) / [Estimation](subdomain-domain.process.subdomain.estimation.md)

# Model Facts — Estimation

Association multiplicity, association invariant, and index uniqueness constraints for this subdomain.

## Associations

- each Definition::Family (has estimates) links to any number of Estimates; each Estimate links to exactly one Definition::Family (Size and time estimates recorded in this family.).
- each Estimate (has history) links to any number of Estimate Historics; each Estimate Historic links to exactly one Estimate; each Estimate–Estimate Historic pairing has the uniqueness → Version (Prior versions of this estimate.).
- each Estimate (uses language) links to exactly one Definition::Language; each Definition::Language may link to any number of Estimates (Language this estimate is categorized by. The language must belong to the same family.).
- each Estimate Historic (belongs to family) links to exactly one Definition::Family; each Definition::Family may link to any number of Estimate Historics (Family copied from the estimate this snapshot belongs to.).
- each Estimate Historic (uses language) links to exactly one Definition::Language; each Definition::Language may link to any number of Estimate Historics (Language copied from the estimate this snapshot belongs to.).

