[⇦ Development](model.md) / [Project](domain-domain.project.md)

# Schedule

The planned calendar for a project, in days or weeks.
## Classes

The classes of this subdomain.


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
class_domain_project_subdomain_core_class_project "1" --> "0..1" class_domain_project_subdomain_schedule_class_schedule : Has Schedule
class_domain_project_subdomain_schedule_class_schedule_week "*" --> "1" class_domain_project_subdomain_core_class_project : For Project
class_domain_project_subdomain_task_class_task "*" --> "1" class_domain_project_subdomain_schedule_class_schedule_week : On Week
class_domain_project_subdomain_schedule_class_schedule "1" --> "*" class_domain_project_subdomain_schedule_class_schedule_week : Has Weeks<br/>{unique → Num}

```

- **[Schedule](class-domain.project.subdomain.schedule.class.schedule.md).** The planned calendar for a project, in days or weeks.
- **[Schedule Week](class-domain.project.subdomain.schedule.class.schedule_week.md).** One week (or day slot) on a project schedule.
- **[Core::Project](class-domain.project.subdomain.core.class.project.md).** Work that follows a process.
- **[Task::Task](class-domain.project.subdomain.task.class.task.md).** A planned task on a project, assigned to a phase and a schedule week.

[Model facts](subdomain-domain.project.subdomain.schedule-facts.md)


## Use Cases






