[⇦ Development Process](model.md) / [Process](domain-domain.process.md) / [Definition](subdomain-domain.process.subdomain.definition.md)

# Stats Bucket

A bucket used to group project statistics.




## Attributes

| Name | Rules | Nullable | TLA+ | Comments / Invariants |
| ---- | ----- | -------- | ---- | --------------------- |
| Name | _(unparsed)_ unconstrained | false |  |  |
| Description | _(unparsed)_ unconstrained | false |  |  |




## Relations

The classes in this diagram.

```mermaid
---
config:
  class:
    hideEmptyMembersBox: true
---
classDiagram
class class_domain_process_subdomain_definition_class_stats_bucket["Stats Bucket"] {
        Name
        Description
    }
namespace Project {
class class_domain_process_subdomain_project_class_project["Project"] {
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
class class_domain_process_subdomain_project_class_project_part["Project Part"] {
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
class class_domain_process_subdomain_project_class_project_stat_phase["Project Stat Phase"] {
            Estimate Minute
            Estimate Comment
        }
}
style class_domain_process_subdomain_definition_class_stats_bucket stroke:#9370DB,stroke-width:3px
class_domain_process_subdomain_project_class_project "*" --> "0..1" class_domain_process_subdomain_definition_class_stats_bucket : In Bucket
class_domain_process_subdomain_project_class_project_part "*" --> "0..1" class_domain_process_subdomain_definition_class_stats_bucket : In Bucket
class_domain_process_subdomain_project_class_project_stat_phase "*" --> "1" class_domain_process_subdomain_definition_class_stats_bucket : In Bucket

```
- **[Project::Project](class-domain.process.subdomain.project.class.project.md).** Work that follows a process.
- **[Project::Project Part](class-domain.process.subdomain.project.class.project_part.md).** A language-specific part of a project.
- **[Project::Project Stat Phase](class-domain.process.subdomain.project.class.project_stat_phase.md).** A per-phase statistical estimate on a project. stat_phase is Phase; bucket is Stats Bucket.
- **[Stats Bucket](class-domain.process.subdomain.definition.class.stats_bucket.md).** A bucket used to group project statistics.


# State Machine


## State and Event Descriptions

The states for this class.

*None*

The events for this class.

*None*



## Action Specifications

The actions for this class.

*None*

## Query Specifications

The queries for this class.

*None*
