[⇦ Development Process](model.md) / [Process](domain-domain.process.md) / [Project](subdomain-domain.process.subdomain.project.md)

# Project Cycle Actual

Actual recording values for one cycle of a project.




## Attributes

| Name | Rules | Nullable | TLA+ | Comments / Invariants |
| ---- | ----- | -------- | ---- | --------------------- |
| Num | _(unparsed)_ [0 .. unconstrained] at 1 unit | true |  | Cycle number. Unique among cycle actuals of the same project. |
| Actual Pct Reuse | _(unparsed)_ [0 .. 100] at 1 percent | false |  |  |




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
class class_domain_process_subdomain_project_class_project_cycle_actual["Project Cycle Actual"] {
        Num
        Actual Pct Reuse
    }
namespace Definition {
class class_domain_process_subdomain_definition_class_module_template["Module Template"] {
            Name
            Description
        }
}
style class_domain_process_subdomain_project_class_project_cycle_actual stroke:#9370DB,stroke-width:3px
class_domain_process_subdomain_project_class_project_cycle_actual "*" --> "1" class_domain_process_subdomain_definition_class_module_template : Instantiates
class_domain_process_subdomain_project_class_project "1" --> "*" class_domain_process_subdomain_project_class_project_cycle_actual : Has Cycle Actuals<br/>{unique → Num}

```
- **[Definition::Module Template](class-domain.process.subdomain.definition.class.module_template.md).** Shared configuration for projects (and project parts) that follow a process.
- **[Project](class-domain.process.subdomain.project.class.project.md).** Work that follows a process.
- **[Project Cycle Actual](class-domain.process.subdomain.project.class.project_cycle_actual.md).** Actual recording values for one cycle of a project.


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
