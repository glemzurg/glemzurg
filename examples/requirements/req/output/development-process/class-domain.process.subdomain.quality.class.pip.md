[⇦ Development Process](model.md) / [Process](domain-domain.process.md) / [Quality](subdomain-domain.process.subdomain.quality.md)

# Process Improvement Proposal

A process improvement proposal raised on a project.




## Attributes

| Name | Rules | Nullable | TLA+ | Comments / Invariants |
| ---- | ----- | -------- | ---- | --------------------- |
| Found Time | _(unparsed)_ datetime | false |  |  |
| Problem | _(unparsed)_ unconstrained | false |  |  |
| Proposal | _(unparsed)_ unconstrained | false |  |  |
| Resolved Time | _(unparsed)_ datetime | false |  |  |




## Relations

The classes in this diagram.

```mermaid
---
config:
  class:
    hideEmptyMembersBox: true
---
classDiagram
class class_domain_process_subdomain_quality_class_pip["Process Improvement Proposal"] {
        Found Time
        Problem
        Proposal
        Resolved Time
    }
namespace Definition {
class class_domain_process_subdomain_definition_class_phase["Phase"] {
            Num
            Name
            Description
        }
class class_domain_process_subdomain_definition_class_process["Process"] {
            Name
            Version
            Version Minor
            Purpose
            Entry Criteria
            Exit Criteria
            Script Lock
            Size Unit
            Size K Unit
        }
class class_domain_process_subdomain_definition_class_step["Step"] {
            Num
            Name
            Tasks
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
style class_domain_process_subdomain_quality_class_pip stroke:#9370DB,stroke-width:3px
class_domain_process_subdomain_quality_class_pip "*" --> "1" class_domain_process_subdomain_definition_class_phase : On Phase
class_domain_process_subdomain_quality_class_pip "*" --> "1" class_domain_process_subdomain_definition_class_process : On Process
class_domain_process_subdomain_quality_class_pip "*" --> "1" class_domain_process_subdomain_definition_class_process : Resolved In Process
class_domain_process_subdomain_quality_class_pip "*" --> "1" class_domain_process_subdomain_definition_class_step : On Subphase
class_domain_process_subdomain_quality_class_pip "*" --> "1" class_domain_process_subdomain_project_class_project : On Project

```
- **[Definition::Phase](class-domain.process.subdomain.definition.class.phase.md).** Fundamental phase skeleton for a process family.
- **[Definition::Process](class-domain.process.subdomain.definition.class.process.md).** A versioned process to follow, owned by a family.
- **[Process Improvement Proposal](class-domain.process.subdomain.quality.class.pip.md).** A process improvement proposal raised on a project.
- **[Project::Project](class-domain.process.subdomain.project.class.project.md).** Work that follows a process.
- **[Definition::Step](class-domain.process.subdomain.definition.class.step.md).** A step of a process script.


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
