[⇦ Development](model.md) / [Process](domain-domain.process.md) / [Method](subdomain-domain.process.subdomain.method.md)

# Model Facts — Method

Association multiplicity, association invariant, and index uniqueness constraints for this subdomain.

## Associations

- each Definition::Module Template (uses design method) links to exactly one Design Method; each Design Method may link to any number of Definition::Module Templates (Design template used under this module template.).
- each Project::Core::Project (uses design method) links to exactly one Design Method; each Design Method may link to any number of Project::Core::Projects (Design template used by this project.).
- each Project::Core::Project Part (uses design method) links to exactly one Design Method; each Design Method may link to any number of Project::Core::Project Parts.

