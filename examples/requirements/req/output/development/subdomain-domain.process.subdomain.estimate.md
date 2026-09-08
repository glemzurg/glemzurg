[⇦ Development Process](model.md) / [Process](domain-domain.process.md)

# Estimate

Programming methods used when estimating size and time.
## Classes

The classes of this subdomain.


```mermaid
---
config:
  class:
    hideEmptyMembersBox: true
---
classDiagram
class class_domain_process_subdomain_estimate_class_method["Method"] {
        Name [key]
        Description
    }
namespace Definition {
class class_domain_process_subdomain_definition_class_module_template["Module Template"] {
            Name
            Description
        }
}
namespace Project.Core {
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
class class_domain_project_subdomain_core_class_project_part["Project Part"] {
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
class class_domain_project_subdomain_core_class_project_stat_phase["Project Stat Phase"] {
            Estimate Minute
            Estimate Comment
        }
}
class_domain_project_subdomain_core_class_project "*" --> "1" class_domain_process_subdomain_estimate_class_method : Uses Size Estimation Method
class_domain_project_subdomain_core_class_project "*" --> "1" class_domain_process_subdomain_estimate_class_method : Uses Time Estimation Method
class_domain_project_subdomain_core_class_project_part "*" --> "1" class_domain_process_subdomain_estimate_class_method : Uses Size Estimation Method
class_domain_project_subdomain_core_class_project_part "*" --> "1" class_domain_process_subdomain_estimate_class_method : Uses Time Estimation Method
class_domain_project_subdomain_core_class_project_stat_phase "*" --> "0..1" class_domain_process_subdomain_estimate_class_method : Uses Method
class_domain_process_subdomain_definition_class_module_template "*" --> "1" class_domain_process_subdomain_estimate_class_method : Uses Size Estimation Method
class_domain_process_subdomain_definition_class_module_template "*" --> "1" class_domain_process_subdomain_estimate_class_method : Uses Time Estimation Method

```

- **[Method](class-domain.process.subdomain.estimate.class.method.md).** A programming method used when recording phase statistics for a project.
- **[Definition::Module Template](class-domain.process.subdomain.definition.class.module_template.md).** Shared configuration for projects (and project parts) that follow a process.
- **[Project::Core::Project](class-domain.project.subdomain.core.class.project.md).** Work that follows a process.
- **[Project::Core::Project Part](class-domain.project.subdomain.core.class.project_part.md).** A language-specific part of a project.
- **[Project::Core::Project Stat Phase](class-domain.project.subdomain.core.class.project_stat_phase.md).** A per-phase statistical estimate on a project. stat_phase is Phase; bucket is Stats Bucket.

[Model facts](subdomain-domain.process.subdomain.estimate-facts.md)


## Use Cases






