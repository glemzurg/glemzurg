[⇦ Development Process](model.md) / [Process](domain-domain.process.md) / [Project](subdomain-domain.process.subdomain.project.md)

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
class class_domain_process_subdomain_project_class_schedule["Schedule"] {
        Day Or Week
    }
class class_domain_process_subdomain_project_class_schedule_week["Schedule Week"] {
        Num
        Date Monday
    }
class class_domain_process_subdomain_project_class_task["Task"] {
        Num
        Name
        Planned Hours
        Pct Complete
    }
style class_domain_process_subdomain_project_class_schedule_week stroke:#9370DB,stroke-width:3px
class_domain_process_subdomain_project_class_schedule "1" --> "*" class_domain_process_subdomain_project_class_schedule_week : Has Weeks<br/>{unique → Num}
class_domain_process_subdomain_project_class_schedule_week "*" --> "1" class_domain_process_subdomain_project_class_project : For Project
class_domain_process_subdomain_project_class_task "*" --> "1" class_domain_process_subdomain_project_class_schedule_week : On Week

```
- **[Project](class-domain.process.subdomain.project.class.project.md).** Work that follows a process.
- **[Schedule](class-domain.process.subdomain.project.class.schedule.md).** The planned calendar for a project, in days or weeks.
- **[Schedule Week](class-domain.process.subdomain.project.class.schedule_week.md).** One week (or day slot) on a project schedule.
- **[Task](class-domain.process.subdomain.project.class.task.md).** A planned task on a project, assigned to a phase and a schedule week.


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
