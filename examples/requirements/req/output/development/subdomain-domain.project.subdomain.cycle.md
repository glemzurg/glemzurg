[⇦ Development Process](model.md) / [Project](domain-domain.project.md)

# Cycle

Planned and actual values recorded per project cycle.
## Classes

The classes of this subdomain.


```mermaid
---
config:
  class:
    hideEmptyMembersBox: true
---
classDiagram
class class_domain_project_subdomain_cycle_class_project_cycle_actual["Project Cycle Actual"] {
        Num
        Actual Pct Reuse
    }
class class_domain_project_subdomain_cycle_class_project_cycle_plan["Project Cycle Plan"] {
        Num
        Planned Pct Reuse
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
class class_domain_process_subdomain_definition_class_module_template["Module Template"] {
            Name
            Description
        }
}
class_domain_project_subdomain_cycle_class_project_cycle_actual "*" --> "1" class_domain_process_subdomain_definition_class_module_template : Instantiates
class_domain_project_subdomain_cycle_class_project_cycle_plan "*" --> "1" class_domain_process_subdomain_definition_class_module_template : Instantiates
class_domain_project_subdomain_core_class_project "1" --> "*" class_domain_project_subdomain_cycle_class_project_cycle_actual : Has Cycle Actuals<br/>{unique → Num}
class_domain_project_subdomain_core_class_project "1" --> "*" class_domain_project_subdomain_cycle_class_project_cycle_plan : Has Cycle Plans<br/>{unique → Num}

```

- **[Project Cycle Actual](class-domain.project.subdomain.cycle.class.project_cycle_actual.md).** Actual recording values for one cycle of a project.
- **[Project Cycle Plan](class-domain.project.subdomain.cycle.class.project_cycle_plan.md).** Planned recording values for one cycle of a project.
- **[Process::Definition::Module Template](class-domain.process.subdomain.definition.class.module_template.md).** Shared configuration for projects (and project parts) that follow a process.
- **[Core::Project](class-domain.project.subdomain.core.class.project.md).** Work that follows a process.

[Model facts](subdomain-domain.project.subdomain.cycle-facts.md)


## Use Cases






