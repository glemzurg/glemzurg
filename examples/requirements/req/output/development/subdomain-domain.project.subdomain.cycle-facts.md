[⇦ Development Process](model.md) / [Project](domain-domain.project.md) / [Cycle](subdomain-domain.project.subdomain.cycle.md)

# Model Facts — Cycle

Association multiplicity, association invariant, and index uniqueness constraints for this subdomain.

## Associations

- each Core::Project (has cycle actuals) links to any number of Project Cycle Actuals; each Project Cycle Actual links to exactly one Core::Project; each Core::Project–Project Cycle Actual pairing has the uniqueness → Num (Actual reuse values recorded per cycle.).
- each Core::Project (has cycle plans) links to any number of Project Cycle Plans; each Project Cycle Plan links to exactly one Core::Project; each Core::Project–Project Cycle Plan pairing has the uniqueness → Num (Planned reuse values recorded per cycle.).
- each Project Cycle Actual (instantiates) links to exactly one Process::Definition::Module Template; each Process::Definition::Module Template may link to any number of Project Cycle Actuals (Module template this cycle actual is based on.).
- each Project Cycle Plan (instantiates) links to exactly one Process::Definition::Module Template; each Process::Definition::Module Template may link to any number of Project Cycle Plans (Module template this cycle plan is based on.).

