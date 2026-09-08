[⇦ Development Process](model.md) / [Project](domain-domain.project.md) / [Estimation](subdomain-domain.project.subdomain.estimation.md)

# Estimate Probe Object Loc

A new-object line in a PROBE size estimate.




## Attributes

| Name | Rules | Nullable | TLA+ | Comments / Invariants |
| ---- | ----- | -------- | ---- | --------------------- |
| Name | _(unparsed)_ unconstrained | false |  |  |
| Loc Per Method | _(unparsed)_ [0 .. unconstrained] at 1 loc | false |  |  |
| Actual Loc | _(unparsed)_ [0 .. unconstrained] at 1 loc | false |  | SQL column acual_loc. |
| For Reuse | _(unparsed)_ enum of TRUE, FALSE | false |  |  |




## Relations

The classes in this diagram.

```mermaid
---
config:
  class:
    hideEmptyMembersBox: true
---
classDiagram
class class_domain_project_subdomain_estimation_class_estimate_probe_object_loc["Estimate Probe Object Loc"] {
        Name
        Loc Per Method
        Actual Loc
        For Reuse
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
class class_domain_process_subdomain_definition_class_probe_object_size["Probe Object Size"] {
            Number
            Name
            Description
        }
class class_domain_process_subdomain_definition_class_probe_type["Probe Type"] {
            Number
            Name
            Description
        }
}
style class_domain_project_subdomain_estimation_class_estimate_probe_object_loc stroke:#9370DB,stroke-width:3px
class_domain_project_subdomain_estimation_class_estimate_probe_object_loc "*" --> "1" class_domain_process_subdomain_definition_class_phase : For Phase
class_domain_project_subdomain_estimation_class_estimate_probe_object_loc "*" --> "1" class_domain_process_subdomain_definition_class_probe_object_size : Of Size
class_domain_project_subdomain_estimation_class_estimate_probe_object_loc "*" --> "1" class_domain_process_subdomain_definition_class_probe_type : Of Type
class_domain_project_subdomain_core_class_project "1" --> "*" class_domain_project_subdomain_estimation_class_estimate_probe_object_loc : Has Probe Object Loc

```
- **[Estimate Probe Object Loc](class-domain.project.subdomain.estimation.class.estimate_probe_object_loc.md).** A new-object line in a PROBE size estimate.
- **[Process::Definition::Phase](class-domain.process.subdomain.definition.class.phase.md).** Fundamental phase skeleton for a process family.
- **[Process::Definition::Probe Object Size](class-domain.process.subdomain.definition.class.probe_object_size.md).** A relative size category used when estimating objects with PROBE.
- **[Process::Definition::Probe Type](class-domain.process.subdomain.definition.class.probe_type.md).** A PROBE object-type category used when listing added and new objects.
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
