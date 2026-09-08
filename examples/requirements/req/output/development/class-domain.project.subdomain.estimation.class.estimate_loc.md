[⇦ Development Process](model.md) / [Project](domain-domain.project.md) / [Estimation](subdomain-domain.project.subdomain.estimation.md)

# Estimate Loc

Planned lines-of-code account for a project.




## Attributes

| Name | Rules | Nullable | TLA+ | Comments / Invariants |
| ---- | ----- | -------- | ---- | --------------------- |
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
class class_domain_project_subdomain_estimation_class_estimate_loc["Estimate Loc"] {
        Base
        New
        Changed
        Added
        Modified
        Deleted
        Reused
        Object Loc
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
namespace Process.Family {
class class_domain_process_subdomain_family_class_phase["Phase"] {
            Num
            Name
            Description
        }
}
style class_domain_project_subdomain_estimation_class_estimate_loc stroke:#9370DB,stroke-width:3px
class_domain_project_subdomain_estimation_class_estimate_loc "*" --> "1" class_domain_process_subdomain_family_class_phase : For Phase
class_domain_project_subdomain_core_class_project "1" --> "0..1" class_domain_project_subdomain_estimation_class_estimate_loc : Has Loc Estimate

```
- **[Estimate Loc](class-domain.project.subdomain.estimation.class.estimate_loc.md).** Planned lines-of-code account for a project.
- **[Process::Family::Phase](class-domain.process.subdomain.family.class.phase.md).** Fundamental phase skeleton for a process family.
- **[Core::Project](class-domain.project.subdomain.core.class.project.md).** Work that follows a process.


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
