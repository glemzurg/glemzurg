[⇦ Development](model.md) / [Project](domain-domain.project.md) / [Cycle](subdomain-domain.project.subdomain.cycle.md)

# Project Cycle Plan

Planned recording values for one cycle of a project.




## Attributes

| Name | Rules | Nullable | TLA+ | Comments / Invariants |
| ---- | ----- | -------- | ---- | --------------------- |
| Num | _(unparsed)_ [0 .. unconstrained] at 1 unit | true |  | Cycle number. Unique among cycle plans of the same project. |
| Planned Pct Reuse | _(unparsed)_ [0 .. 100] at 1 percent | false |  |  |




## Relations

The classes in this diagram.

```mermaid
---
config:
  class:
    hideEmptyMembersBox: true
---
classDiagram
class class_domain_project_subdomain_cycle_class_project_cycle_plan["Project Cycle Plan"] {
        Num
        Planned Pct Reuse
    }
namespace Core {
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
}
namespace Process.Definition {
class class_domain_process_subdomain_definition_class_module_template["Module Template"] {
            Name
            Description
        }
}
style class_domain_project_subdomain_cycle_class_project_cycle_plan stroke:#9370DB,stroke-width:3px
class_domain_project_subdomain_cycle_class_project_cycle_plan "*" --> "1" class_domain_process_subdomain_definition_class_module_template : Instantiates
class_domain_project_subdomain_core_class_project "1" --> "*" class_domain_project_subdomain_cycle_class_project_cycle_plan : Has Cycle Plans<br/>{unique → Num}

```
- **[Process::Definition::Module Template](class-domain.process.subdomain.definition.class.module_template.md).** Shared configuration for projects (and project parts) that follow a process.
- **[Core::Project](class-domain.project.subdomain.core.class.project.md).** Work that follows a process.
- **[Project Cycle Plan](class-domain.project.subdomain.cycle.class.project_cycle_plan.md).** Planned recording values for one cycle of a project.


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
