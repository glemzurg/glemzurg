[⇦ Development Process](model.md) / [Process](domain-domain.process.md) / [Definition](subdomain-domain.process.subdomain.definition.md)

# Method

A programming method used when recording phase statistics for a project.




## Attributes

| Name | Rules | Nullable | TLA+ | Comments / Invariants |
| ---- | ----- | -------- | ---- | --------------------- |
| Name [key] | _(unparsed)_ unconstrained | false |  | Unique name of this method. |
| Description | _(unparsed)_ unconstrained | false |  | Defaults to empty when omitted. |


### Indexes

- key [Name]





## Relations

The classes in this diagram.

```mermaid
---
config:
  class:
    hideEmptyMembersBox: true
---
classDiagram
class class_domain_process_subdomain_definition_class_method["Method"] {
        Name [key]
        Description
    }
class class_domain_process_subdomain_definition_class_module_template["Module Template"] {
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
style class_domain_process_subdomain_definition_class_method stroke:#9370DB,stroke-width:3px
class_domain_process_subdomain_project_class_project "*" --> "1" class_domain_process_subdomain_definition_class_method : Uses Size Estimation Method
class_domain_process_subdomain_project_class_project "*" --> "1" class_domain_process_subdomain_definition_class_method : Uses Time Estimation Method
class_domain_process_subdomain_project_class_project_part "*" --> "1" class_domain_process_subdomain_definition_class_method : Uses Size Estimation Method
class_domain_process_subdomain_project_class_project_part "*" --> "1" class_domain_process_subdomain_definition_class_method : Uses Time Estimation Method
class_domain_process_subdomain_project_class_project_stat_phase "*" --> "0..1" class_domain_process_subdomain_definition_class_method : Uses Method
class_domain_process_subdomain_definition_class_module_template "*" --> "1" class_domain_process_subdomain_definition_class_method : Uses Size Estimation Method
class_domain_process_subdomain_definition_class_module_template "*" --> "1" class_domain_process_subdomain_definition_class_method : Uses Time Estimation Method

```
- **[Method](class-domain.process.subdomain.definition.class.method.md).** A programming method used when recording phase statistics for a project.
- **[Module Template](class-domain.process.subdomain.definition.class.module_template.md).** Shared configuration for projects (and project parts) that follow a process.
- **[Project::Project](class-domain.process.subdomain.project.class.project.md).** Work that follows a process.
- **[Project::Project Part](class-domain.process.subdomain.project.class.project_part.md).** A language-specific part of a project.
- **[Project::Project Stat Phase](class-domain.process.subdomain.project.class.project_stat_phase.md).** A per-phase statistical estimate on a project. stat_phase is Phase; bucket is Stats Bucket.


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
