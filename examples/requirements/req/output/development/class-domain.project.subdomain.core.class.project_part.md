[⇦ Development](model.md) / [Project](domain-domain.project.md) / [Core](subdomain-domain.project.subdomain.core.md)

# Project Part

A language-specific part of a project.




## Attributes

| Name | Rules | Nullable | TLA+ | Comments / Invariants |
| ---- | ----- | -------- | ---- | --------------------- |
| Name | _(unparsed)_ unconstrained | false |  |  |
| Description | _(unparsed)_ unconstrained | true |  |  |
| Multi Day | _(unparsed)_ enum of TRUE, FALSE | false |  | SQL column mutli_day. |
| Planned Time | _(unparsed)_ [0 .. unconstrained] at 1 unit | false |  |  |
| Actual Time | _(unparsed)_ [0 .. unconstrained] at 1 unit | false |  |  |
| Planned Pct Reuse | _(unparsed)_ [0 .. 100] at 1 percent | false |  |  |
| Actual Pct Reuse | _(unparsed)_ [0 .. 100] at 1 percent | false |  |  |
| Planned Defect Count | _(unparsed)_ [0 .. unconstrained] at 1 unit | false |  |  |
| Planned Appraisal Coq | _(unparsed)_ [0 .. unconstrained] at 1 unit | false |  |  |
| Planned Failure Coq | _(unparsed)_ [0 .. unconstrained] at 1 unit | false |  |  |




## Relations

The classes in this diagram.

```mermaid
---
config:
  class:
    hideEmptyMembersBox: true
---
classDiagram
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
namespace Process.Definition {
class class_domain_process_subdomain_definition_class_module_template["Module Template"] {
            Name
            Description
        }
}
namespace Process.Estimate {
class class_domain_process_subdomain_estimate_class_method["Method"] {
            Name [key]
            Description
        }
}
namespace Process.Family {
class class_domain_process_subdomain_family_class_language["Language"] {
            Name
            Description
        }
class class_domain_process_subdomain_family_class_phase["Phase"] {
            Num
            Name
            Description
        }
}
namespace Process.Method {
class class_domain_process_subdomain_method_class_design_method["Design Method"] {
            Name
            Description
        }
}
namespace Process.Process {
class class_domain_process_subdomain_process_class_step["Step"] {
            Num
            Name
            Tasks
        }
}
namespace Statistics {
class class_domain_statistics_subdomain_default_class_stats_bucket["Stats Bucket"] {
            Name
            Description
        }
}
style class_domain_project_subdomain_core_class_project_part stroke:#9370DB,stroke-width:3px
class_domain_project_subdomain_core_class_project_part "*" --> "1" class_domain_process_subdomain_definition_class_module_template : Instantiates
class_domain_project_subdomain_core_class_project_part "*" --> "1" class_domain_process_subdomain_estimate_class_method : Uses Size Estimation Method
class_domain_project_subdomain_core_class_project_part "*" --> "1" class_domain_process_subdomain_estimate_class_method : Uses Time Estimation Method
class_domain_project_subdomain_core_class_project_part "*" --> "1" class_domain_process_subdomain_family_class_language : Uses Language
class_domain_project_subdomain_core_class_project_part "*" --> "0..1" class_domain_process_subdomain_family_class_phase : Current Phase
class_domain_project_subdomain_core_class_project_part "*" --> "1" class_domain_process_subdomain_method_class_design_method : Uses Design Method
class_domain_project_subdomain_core_class_project_part "*" --> "0..1" class_domain_process_subdomain_process_class_step : Current Subphase
class_domain_project_subdomain_core_class_project_part "*" --> "0..1" class_domain_statistics_subdomain_default_class_stats_bucket : In Bucket
class_domain_project_subdomain_core_class_project "1" --> "*" class_domain_project_subdomain_core_class_project_part : Has Parts

```
- **[Process::Method::Design Method](class-domain.process.subdomain.method.class.design_method.md).** A design template used when planning a project or module.
- **[Process::Family::Language](class-domain.process.subdomain.family.class.language.md).** A programming language used when estimating size or time in a process family.
- **[Process::Estimate::Method](class-domain.process.subdomain.estimate.class.method.md).** A programming method used when recording phase statistics for a project.
- **[Process::Definition::Module Template](class-domain.process.subdomain.definition.class.module_template.md).** Shared configuration for projects (and project parts) that follow a process.
- **[Process::Family::Phase](class-domain.process.subdomain.family.class.phase.md).** Fundamental phase skeleton for all the processes in a family.
- **[Project](class-domain.project.subdomain.core.class.project.md).** Work that follows a process.
- **[Project Part](class-domain.project.subdomain.core.class.project_part.md).** A language-specific part of a project.
- **[Statistics::Stats Bucket](class-domain.statistics.subdomain.default.class.stats_bucket.md).** A bucket used to group project statistics.
- **[Process::Process::Step](class-domain.process.subdomain.process.class.step.md).** A step of a process script.


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
