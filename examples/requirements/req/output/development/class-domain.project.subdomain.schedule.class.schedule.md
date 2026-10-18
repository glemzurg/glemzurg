[⇦ Development](model.md) / [Project](domain-domain.project.md) / [Schedule](subdomain-domain.project.subdomain.schedule.md)

# Schedule

The planned calendar for a project, in days or weeks.




## Attributes

| Name | Rules | Nullable | TLA+ | Comments / Invariants |
| ---- | ----- | -------- | ---- | --------------------- |
| Day Or Week | _(unparsed)_ enum of day, week | false |  | Whether the schedule is recorded by day or by week. |




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
style class_domain_project_subdomain_schedule_class_schedule stroke:#9370DB,stroke-width:3px
class_domain_project_subdomain_core_class_project "1" --> "0..1" class_domain_project_subdomain_schedule_class_schedule : Has Schedule
class_domain_project_subdomain_schedule_class_schedule "1" --> "*" class_domain_project_subdomain_schedule_class_schedule_week : Has Weeks<br/>{unique → Num}

```
- **[Core::Project](class-domain.project.subdomain.core.class.project.md).** Work that follows a process.
- **[Schedule](class-domain.project.subdomain.schedule.class.schedule.md).** The planned calendar for a project, in days or weeks.
- **[Schedule Week](class-domain.project.subdomain.schedule.class.schedule_week.md).** One week (or day slot) on a project schedule.


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
