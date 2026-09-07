[⇦ Development Process](model.md) / [Process](domain-domain.process.md) / [Project](subdomain-domain.process.subdomain.project.md)

# Time Log

A recorded interval of work on a project.




## Attributes

| Name | Rules | Nullable | TLA+ | Comments / Invariants |
| ---- | ----- | -------- | ---- | --------------------- |
| Cycle | _(unparsed)_ [0 .. unconstrained] at 1 unit | false |  |  |
| Start Time | _(unparsed)_ datetime | false |  |  |
| Stop Time | _(unparsed)_ datetime | false |  |  |
| Interruption Minutes | _(unparsed)_ [0 .. unconstrained] at 1 minute | false |  |  |
| Comments | _(unparsed)_ unconstrained | false |  |  |




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
class class_domain_process_subdomain_project_class_task["Task"] {
        Num
        Name
        Planned Hours
        Pct Complete
    }
class class_domain_process_subdomain_project_class_time_log["Time Log"] {
        Cycle
        Start Time
        Stop Time
        Interruption Minutes
        Comments
    }
namespace Definition {
class class_domain_process_subdomain_definition_class_phase["Phase"] {
            Num
            Name
            Description
        }
}
style class_domain_process_subdomain_project_class_time_log stroke:#9370DB,stroke-width:3px
class_domain_process_subdomain_project_class_time_log "*" --> "1" class_domain_process_subdomain_definition_class_phase : Occurs In
class_domain_process_subdomain_project_class_project "1" --> "*" class_domain_process_subdomain_project_class_time_log : Has Time Logs
class_domain_process_subdomain_project_class_time_log "*" --> "1" class_domain_process_subdomain_project_class_task : For Task

```
- **[Definition::Phase](class-domain.process.subdomain.definition.class.phase.md).** Fundamental phase skeleton for a process family.
- **[Project](class-domain.process.subdomain.project.class.project.md).** Work that follows a process.
- **[Task](class-domain.process.subdomain.project.class.task.md).** A planned task on a project, assigned to a phase and a schedule week.
- **[Time Log](class-domain.process.subdomain.project.class.time_log.md).** A recorded interval of work on a project.


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
