[⇦ Development Process](model.md) / [Project](domain-domain.project.md) / [Estimation](subdomain-domain.project.subdomain.estimation.md)

# Estimate Probe Object Reused

A reused-object line in a PROBE size estimate.




## Attributes

| Name | Rules | Nullable | TLA+ | Comments / Invariants |
| ---- | ----- | -------- | ---- | --------------------- |
| Name | _(unparsed)_ unconstrained | false |  |  |
| Loc | _(unparsed)_ [0 .. unconstrained] at 1 loc | false |  |  |
| Actual Loc | _(unparsed)_ [0 .. unconstrained] at 1 loc | false |  | SQL column acual_loc. |




## Relations

The classes in this diagram.

```mermaid
---
config:
  class:
    hideEmptyMembersBox: true
---
classDiagram
class class_domain_project_subdomain_estimation_class_estimate_probe_object_reused["Estimate Probe Object Reused"] {
        Name
        Loc
        Actual Loc
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
class class_domain_process_subdomain_definition_class_phase["Phase"] {
            Num
            Name
            Description
        }
}
style class_domain_project_subdomain_estimation_class_estimate_probe_object_reused stroke:#9370DB,stroke-width:3px
class_domain_project_subdomain_estimation_class_estimate_probe_object_reused "*" --> "1" class_domain_process_subdomain_definition_class_phase : For Phase
class_domain_project_subdomain_core_class_project "1" --> "*" class_domain_project_subdomain_estimation_class_estimate_probe_object_reused : Has Probe Object Reused

```
- **[Estimate Probe Object Reused](class-domain.project.subdomain.estimation.class.estimate_probe_object_reused.md).** A reused-object line in a PROBE size estimate.
- **[Process::Definition::Phase](class-domain.process.subdomain.definition.class.phase.md).** Fundamental phase skeleton for a process family.
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
