[⇦ Development Process](model.md) / [Process](domain-domain.process.md) / [Estimation](subdomain-domain.process.subdomain.estimation.md)

# Actual Loc

Actual lines-of-code account for a project.




## Attributes

| Name | Rules | Nullable | TLA+ | Comments / Invariants |
| ---- | ----- | -------- | ---- | --------------------- |
| Cycle | _(unparsed)_ [0 .. unconstrained] at 1 unit | false |  |  |
| Base | _(unparsed)_ [0 .. unconstrained] at 1 loc | false |  |  |
| New | _(unparsed)_ [0 .. unconstrained] at 1 loc | false |  |  |
| Changed | _(unparsed)_ [0 .. unconstrained] at 1 loc | false |  |  |
| Added | _(unparsed)_ [0 .. unconstrained] at 1 loc | false |  |  |
| Modified | _(unparsed)_ [0 .. unconstrained] at 1 loc | false |  |  |
| Deleted | _(unparsed)_ [0 .. unconstrained] at 1 loc | false |  |  |
| Reused | _(unparsed)_ [0 .. unconstrained] at 1 loc | false |  |  |
| Object Loc | _(unparsed)_ [0 .. unconstrained] at 1 loc | false |  |  |




## Relations

The classes in this diagram.

```mermaid
---
config:
  class:
    hideEmptyMembersBox: true
---
classDiagram
class class_domain_process_subdomain_estimation_class_actual_loc["Actual Loc"] {
        Cycle
        Base
        New
        Changed
        Added
        Modified
        Deleted
        Reused
        Object Loc
    }
namespace Definition {
class class_domain_process_subdomain_definition_class_phase["Phase"] {
            Num
            Name
            Description
        }
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
}
style class_domain_process_subdomain_estimation_class_actual_loc stroke:#9370DB,stroke-width:3px
class_domain_process_subdomain_estimation_class_actual_loc "*" --> "1" class_domain_process_subdomain_definition_class_phase : For Phase
class_domain_process_subdomain_project_class_project "1" --> "0..1" class_domain_process_subdomain_estimation_class_actual_loc : Has Actual Loc

```
- **[Actual Loc](class-domain.process.subdomain.estimation.class.actual_loc.md).** Actual lines-of-code account for a project.
- **[Definition::Phase](class-domain.process.subdomain.definition.class.phase.md).** Fundamental phase skeleton for a process family.
- **[Project::Project](class-domain.process.subdomain.project.class.project.md).** Work that follows a process.


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
