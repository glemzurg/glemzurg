[⇦ Development](model.md) / [Project](domain-domain.project.md) / [Schedule](subdomain-domain.project.subdomain.schedule.md)

# Schedule Week

One week (or day slot) on a project schedule.




## Attributes

| Name | Rules | Nullable | TLA+ | Comments / Invariants |
| ---- | ----- | -------- | ---- | --------------------- |
| Num | _(unparsed)_ [0 .. unconstrained] at 1 unit | false |  | Order of this week on the schedule. |
| Date Monday | _(unparsed)_ datetime | false |  | Monday of this schedule week. |

## Invariants

- The project of a schedule week is the project of its schedule.
    - **LET schedule == CHOOSE s ∈ self._HasWeeks : TRUE IN self.ForProject = schedule._HasSchedule**



## Relations

The classes in this diagram.

```mermaid
---
config:
  class:
    hideEmptyMembersBox: true
---
classDiagram
class class_domain_project_subdomain_schedule_class_schedule["Schedule"] {
        Day Or Week
    }
class class_domain_project_subdomain_schedule_class_schedule_week["Schedule Week"] {
        Num
        Date Monday
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
namespace Task {
class class_domain_project_subdomain_task_class_task["Task"] {
            Num
            Name
            Planned Hours
            Pct Complete
        }
}
style class_domain_project_subdomain_schedule_class_schedule_week stroke:#9370DB,stroke-width:3px
class_domain_project_subdomain_schedule_class_schedule_week "*" --> "1" class_domain_project_subdomain_core_class_project : For Project
class_domain_project_subdomain_task_class_task "*" --> "1" class_domain_project_subdomain_schedule_class_schedule_week : On Week
class_domain_project_subdomain_schedule_class_schedule "1" --> "*" class_domain_project_subdomain_schedule_class_schedule_week : Has Weeks<br/>{unique → Num}

```
- **[Core::Project](class-domain.project.subdomain.core.class.project.md).** Work that follows a process.
- **[Schedule](class-domain.project.subdomain.schedule.class.schedule.md).** The planned calendar for a project, in days or weeks.
- **[Schedule Week](class-domain.project.subdomain.schedule.class.schedule_week.md).** One week (or day slot) on a project schedule.
- **[Task::Task](class-domain.project.subdomain.task.class.task.md).** A planned task on a project, assigned to a phase and a schedule week.


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
