[⇦ Development Process](model.md)

# Statistics

Buckets used to group project statistics.

## Default

Statistics buckets.

## Classes

The classes of this domain.


```mermaid
---
config:
  class:
    hideEmptyMembersBox: true
---
classDiagram
class class_domain_statistics_subdomain_default_class_stats_bucket["Stats Bucket"] {
        Name
        Description
    }
namespace Project.Core {
class class_domain_project_subdomain_core_class_project["Project"] {
            Name
            Description
            Created Time
            Started Time
            Estimate Minute
            Estimate Comment
            Multi Day
            Planned Time
            Actual Time
            Planned Pct Reuse
            Actual Pct Reuse
            Planned Defect Count
            Planned Appraisal Coq
            Planned Failure Coq
        }
class class_domain_project_subdomain_core_class_project_part["Project Part"] {
            Name
            Description
            Multi Day
            Planned Time
            Actual Time
            Planned Pct Reuse
            Actual Pct Reuse
            Planned Defect Count
            Planned Appraisal Coq
            Planned Failure Coq
        }
class class_domain_project_subdomain_core_class_project_stat_phase["Project Stat Phase"] {
            Estimate Minute
            Estimate Comment
        }
}
class_domain_project_subdomain_core_class_project "*" --> "0..1" class_domain_statistics_subdomain_default_class_stats_bucket : In Bucket
class_domain_project_subdomain_core_class_project_part "*" --> "0..1" class_domain_statistics_subdomain_default_class_stats_bucket : In Bucket
class_domain_project_subdomain_core_class_project_stat_phase "*" --> "1" class_domain_statistics_subdomain_default_class_stats_bucket : In Bucket

```

- **[Stats Bucket](class-domain.statistics.subdomain.default.class.stats_bucket.md).** A bucket used to group project statistics.
- **[Project::Core::Project](class-domain.project.subdomain.core.class.project.md).** Work that follows a process.
- **[Project::Core::Project Part](class-domain.project.subdomain.core.class.project_part.md).** A language-specific part of a project.
- **[Project::Core::Project Stat Phase](class-domain.project.subdomain.core.class.project_stat_phase.md).** A per-phase statistical estimate on a project. stat_phase is Phase; bucket is Stats Bucket.

[Model facts](subdomain-domain.statistics.subdomain.default-facts.md)





## Use Cases







