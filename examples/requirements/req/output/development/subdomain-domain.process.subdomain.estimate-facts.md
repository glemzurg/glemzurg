[⇦ Development Process](model.md) / [Process](domain-domain.process.md) / [Estimate](subdomain-domain.process.subdomain.estimate.md)

# Model Facts — Estimate

Association multiplicity, association invariant, and index uniqueness constraints for this subdomain.

## Associations

- each Definition::Module Template (uses size estimation method) links to exactly one Method; each Method may link to any number of Definition::Module Templates (Method used to estimate size under this template.).
- each Definition::Module Template (uses time estimation method) links to exactly one Method; each Method may link to any number of Definition::Module Templates (Method used to estimate time under this template.).
- each Project::Core::Project (uses size estimation method) links to exactly one Method; each Method may link to any number of Project::Core::Projects (Method used to estimate size.).
- each Project::Core::Project (uses time estimation method) links to exactly one Method; each Method may link to any number of Project::Core::Projects (Method used to estimate time.).
- each Project::Core::Project Part (uses size estimation method) links to exactly one Method; each Method may link to any number of Project::Core::Project Parts.
- each Project::Core::Project Part (uses time estimation method) links to exactly one Method; each Method may link to any number of Project::Core::Project Parts.
- each Project::Core::Project Stat Phase (uses method) may link to at most one Method; each Method may link to any number of Project::Core::Project Stat Phases (Programming method for these statistics, when one is set.).

## Indexes

- No Methods can share the same Name.

