[⇦ Development Process](model.md) / [Statistics](domain-domain.statistics.md)

# Model Facts — Default

Association multiplicity, association invariant, and index uniqueness constraints for this subdomain.

## Associations

- each Project::Core::Project (in bucket) may link to at most one Stats Bucket; each Stats Bucket may link to any number of Project::Core::Projects (Stats bucket this project is grouped in, when one is set.).
- each Project::Core::Project Part (in bucket) may link to at most one Stats Bucket; each Stats Bucket may link to any number of Project::Core::Project Parts (Stats bucket this part is grouped in, when one is set.).
- each Project::Core::Project Stat Phase (in bucket) links to exactly one Stats Bucket; each Stats Bucket may link to any number of Project::Core::Project Stat Phases (Stats bucket these statistics are grouped in.).

