[⇦ Development Process](model.md) / [Process](domain-domain.process.md) / [Definition](subdomain-domain.process.subdomain.definition.md)

# Step

A step of a process script. Every step occurs in a phase of the same family as its process.




## Attributes

| Name | Rules | Nullable | TLA+ | Comments / Invariants |
| ---- | ----- | -------- | ---- | --------------------- |
| Num | _(unparsed)_ [0 .. unconstrained] at 1 unit | false |  | Order of this step within the script. |
| Name | _(unparsed)_ unconstrained | false |  | Unique among steps of the same script. |
| Tasks | _(unparsed)_ unconstrained | false |  | Work performed in this step. |

## Invariants

- The phase of a step belongs to the same family as the process that owns the script.
    - **LET script == CHOOSE s ∈ self._HasSteps : TRUE IN LET process == CHOOSE p ∈ script._HasScripts : TRUE IN LET familyFromProcess == CHOOSE f ∈ process._HasProcesses : TRUE IN LET phase == CHOOSE ph ∈ self.OccursIn : TRUE IN LET familyFromPhase == CHOOSE f ∈ phase._HasPhases : TRUE IN familyFromProcess = familyFromPhase**



## Relations

The classes in this diagram.

```mermaid
---
config:
  class:
    hideEmptyMembersBox: true
---
classDiagram
class class_domain_process_subdomain_definition_class_phase["Phase"] {
        Num
        Name
        Description
    }
class class_domain_process_subdomain_definition_class_script["Script"] {
        Num
        Name
        Task Summary
        Purpose
        Entry Criteria
        Exit Criteria
        Cycle
    }
class class_domain_process_subdomain_definition_class_step["Step"] {
        Num
        Name
        Tasks
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
}
namespace Quality {
class class_domain_process_subdomain_quality_class_pip["Process Improvement Proposal"] {
            Found Time
            Problem
            Proposal
            Resolved Time
        }
}
style class_domain_process_subdomain_definition_class_step stroke:#9370DB,stroke-width:3px
class_domain_process_subdomain_project_class_project "*" --> "0..1" class_domain_process_subdomain_definition_class_step : Current Subphase
class_domain_process_subdomain_project_class_project_part "*" --> "0..1" class_domain_process_subdomain_definition_class_step : Current Subphase
class_domain_process_subdomain_quality_class_pip "*" --> "1" class_domain_process_subdomain_definition_class_step : On Subphase
class_domain_process_subdomain_definition_class_script "1" --> "*" class_domain_process_subdomain_definition_class_step : Has Steps<br/>{unique → Num}
class_domain_process_subdomain_definition_class_step "*" --> "1" class_domain_process_subdomain_definition_class_phase : Occurs In

```
- **[Phase](class-domain.process.subdomain.definition.class.phase.md).** Fundamental phase skeleton for a process family.
- **[Quality::Process Improvement Proposal](class-domain.process.subdomain.quality.class.pip.md).** A process improvement proposal raised on a project.
- **[Project::Project](class-domain.process.subdomain.project.class.project.md).** Work that follows a process.
- **[Project::Project Part](class-domain.process.subdomain.project.class.project_part.md).** A language-specific part of a project.
- **[Script](class-domain.process.subdomain.definition.class.script.md).** A step-by-step process script owned by a process.
- **[Step](class-domain.process.subdomain.definition.class.step.md).** A step of a process script.


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
