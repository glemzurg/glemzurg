[⇦ Development Process](model.md) / [Process](domain-domain.process.md) / [Project](subdomain-domain.process.subdomain.project.md)

# Project Stat Phase

A per-phase statistical estimate on a project. stat_phase is Phase; bucket is Stats Bucket.




## Attributes

| Name | Rules | Nullable | TLA+ | Comments / Invariants |
| ---- | ----- | -------- | ---- | --------------------- |
| Estimate Minute | _(unparsed)_ [0 .. unconstrained] at 1 minute | false |  | Defaults to 0. |
| Estimate Comment | _(unparsed)_ unconstrained | false |  | Defaults to empty. |




## Relations

The classes in this diagram.

```mermaid
---
config:
  class:
    hideEmptyMembersBox: true
---
classDiagram
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
class class_domain_process_subdomain_project_class_project_stat_phase["Project Stat Phase"] {
        Estimate Minute
        Estimate Comment
    }
namespace Definition {
class class_domain_process_subdomain_definition_class_method["Method"] {
            Name [key]
            Description
        }
class class_domain_process_subdomain_definition_class_phase["Phase"] {
            Num
            Name
            Description
        }
class class_domain_process_subdomain_definition_class_stats_bucket["Stats Bucket"] {
            Name
            Description
        }
}
style class_domain_process_subdomain_project_class_project_stat_phase stroke:#9370DB,stroke-width:3px
class_domain_process_subdomain_project_class_project_stat_phase "*" --> "0..1" class_domain_process_subdomain_definition_class_method : Uses Method
class_domain_process_subdomain_project_class_project_stat_phase "*" --> "1" class_domain_process_subdomain_definition_class_phase : For Phase
class_domain_process_subdomain_project_class_project_stat_phase "*" --> "1" class_domain_process_subdomain_definition_class_stats_bucket : In Bucket
class_domain_process_subdomain_project_class_project "1" --> "*" class_domain_process_subdomain_project_class_project_stat_phase : Has Stat Phases

```
- **[Definition::Method](class-domain.process.subdomain.definition.class.method.md).** A programming method used when recording phase statistics for a project.
- **[Definition::Phase](class-domain.process.subdomain.definition.class.phase.md).** Fundamental phase skeleton for a process family.
- **[Project](class-domain.process.subdomain.project.class.project.md).** Work that follows a process.
- **[Project Stat Phase](class-domain.process.subdomain.project.class.project_stat_phase.md).** A per-phase statistical estimate on a project. stat_phase is Phase; bucket is Stats Bucket.
- **[Definition::Stats Bucket](class-domain.process.subdomain.definition.class.stats_bucket.md).** A bucket used to group project statistics.


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
