[⇦ Development Process](model.md) / [Project](domain-domain.project.md)

# Task

Planned tasks and recorded time logs on a project.
## Classes

The classes of this subdomain.


```mermaid
---
config:
  class:
    hideEmptyMembersBox: true
---
classDiagram
class class_domain_project_subdomain_task_class_task["Task"] {
        Num
        Name
        Planned Hours
        Pct Complete
    }
class class_domain_project_subdomain_task_class_time_log["Time Log"] {
        Cycle
        Start Time
        Stop Time
        Interruption Minutes
        Comments
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
namespace Schedule {
class class_domain_project_subdomain_schedule_class_schedule_week["Schedule Week"] {
            Num
            Date Monday
        }
}
class_domain_project_subdomain_task_class_task "*" --> "1" class_domain_process_subdomain_family_class_phase : Occurs In
class_domain_project_subdomain_task_class_time_log "*" --> "1" class_domain_process_subdomain_family_class_phase : Occurs In
class_domain_project_subdomain_core_class_project "1" --> "*" class_domain_project_subdomain_task_class_task : Has Tasks
class_domain_project_subdomain_core_class_project "1" --> "*" class_domain_project_subdomain_task_class_time_log : Has Time Logs
class_domain_project_subdomain_task_class_task "*" --> "1" class_domain_project_subdomain_schedule_class_schedule_week : On Week
class_domain_project_subdomain_task_class_time_log "*" --> "1" class_domain_project_subdomain_task_class_task : For Task

```

- **[Task](class-domain.project.subdomain.task.class.task.md).** A planned task on a project, assigned to a phase and a schedule week.
- **[Time Log](class-domain.project.subdomain.task.class.time_log.md).** A recorded interval of work on a project.
- **[Process::Family::Phase](class-domain.process.subdomain.family.class.phase.md).** Fundamental phase skeleton for a process family.
- **[Core::Project](class-domain.project.subdomain.core.class.project.md).** Work that follows a process.
- **[Schedule::Schedule Week](class-domain.project.subdomain.schedule.class.schedule_week.md).** One week (or day slot) on a project schedule.

[Model facts](subdomain-domain.project.subdomain.task-facts.md)


## Use Cases






