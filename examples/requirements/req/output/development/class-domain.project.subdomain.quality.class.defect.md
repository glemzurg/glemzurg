[⇦ Development Process](model.md) / [Project](domain-domain.project.md) / [Quality](subdomain-domain.project.subdomain.quality.md)

# Defect

A defect injected and removed in projects and phases.




## Attributes

| Name | Rules | Nullable | TLA+ | Comments / Invariants |
| ---- | ----- | -------- | ---- | --------------------- |
| Found Time | _(unparsed)_ datetime | false |  |  |
| Cycle | _(unparsed)_ [0 .. unconstrained] at 1 unit | false |  |  |
| Fix Minutes | _(unparsed)_ [0 .. unconstrained] at 1 minute | false |  |  |
| Description | _(unparsed)_ unconstrained | false |  |  |
| Test Defect | _(unparsed)_ unconstrained | false |  | A defect in a test itself. |




## Relations

The classes in this diagram.

```mermaid
---
config:
  class:
    hideEmptyMembersBox: true
---
classDiagram
class class_domain_project_subdomain_quality_class_defect["Defect"] {
        Found Time
        Cycle
        Fix Minutes
        Description
        Test Defect
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
style class_domain_project_subdomain_quality_class_defect stroke:#9370DB,stroke-width:3px
class_domain_project_subdomain_quality_class_defect "*" --> "1" class_domain_process_subdomain_family_class_phase : Injected In Phase
class_domain_project_subdomain_quality_class_defect "*" --> "1" class_domain_process_subdomain_family_class_phase : Removed In Phase
class_domain_project_subdomain_quality_class_defect "*" --> "1" class_domain_project_subdomain_core_class_project : Injected In
class_domain_project_subdomain_quality_class_defect "*" --> "1" class_domain_project_subdomain_core_class_project : Removed In
class_domain_project_subdomain_quality_class_defect "*" --> "0..1" class_domain_project_subdomain_quality_class_defect : Has Source

```
- **[Defect](class-domain.project.subdomain.quality.class.defect.md).** A defect injected and removed in projects and phases.
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
