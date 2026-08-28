[⇦ Development Process](model.md) / [Process](domain-domain.process.md) / [Project](subdomain-domain.process.subdomain.project.md)

# Project Part

A language-specific part of a project.




## Attributes

| Name | Rules | Nullable | TLA+ | Comments / Invariants |
| ---- | ----- | -------- | ---- | --------------------- |
| Name | _(unparsed)_ unconstrained | false |  |  |
| Description | _(unparsed)_ unconstrained | false |  |  |
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
namespace Definition {
class class_domain_process_subdomain_definition_class_design_method["Design Method"] {
            Name
            Description
        }
class class_domain_process_subdomain_definition_class_language["Language"] {
            Name
            Description
        }
class class_domain_process_subdomain_definition_class_method["Method"] {
            Name [key]
            Description
        }
class class_domain_process_subdomain_definition_class_module_template["Module Template"] {
            Name
            Description
        }
class class_domain_process_subdomain_definition_class_phase["Phase"] {
            Num
            Name
            Description
        }
class class_domain_process_subdomain_definition_class_stats_bucket["Stats Bucket"] {
            Name
            Description
        }
class class_domain_process_subdomain_definition_class_step["Step"] {
            Num
            Name
            Tasks
        }
}
style class_domain_process_subdomain_project_class_project_part stroke:#9370DB,stroke-width:3px
class_domain_process_subdomain_project_class_project_part "*" --> "1" class_domain_process_subdomain_definition_class_design_method : Uses Design Method
class_domain_process_subdomain_project_class_project_part "*" --> "1" class_domain_process_subdomain_definition_class_language : Uses Language
class_domain_process_subdomain_project_class_project_part "*" --> "1" class_domain_process_subdomain_definition_class_method : Uses Size Estimation Method
class_domain_process_subdomain_project_class_project_part "*" --> "1" class_domain_process_subdomain_definition_class_method : Uses Time Estimation Method
class_domain_process_subdomain_project_class_project_part "*" --> "1" class_domain_process_subdomain_definition_class_module_template : Instantiates
class_domain_process_subdomain_project_class_project_part "*" --> "0..1" class_domain_process_subdomain_definition_class_phase : Current Phase
class_domain_process_subdomain_project_class_project_part "*" --> "0..1" class_domain_process_subdomain_definition_class_stats_bucket : In Bucket
class_domain_process_subdomain_project_class_project_part "*" --> "0..1" class_domain_process_subdomain_definition_class_step : Current Subphase
class_domain_process_subdomain_project_class_project "1" --> "*" class_domain_process_subdomain_project_class_project_part : Has Parts

```
- **[Definition::Design Method](class-domain.process.subdomain.definition.class.design_method.md).** A design template used when planning a project or module.
- **[Definition::Language](class-domain.process.subdomain.definition.class.language.md).** A programming language used when estimating size or time in a process family.
- **[Definition::Method](class-domain.process.subdomain.definition.class.method.md).** A programming method used when recording phase statistics for a project.
- **[Definition::Module Template](class-domain.process.subdomain.definition.class.module_template.md).** Shared configuration for projects (and project parts) that follow a process.
- **[Definition::Phase](class-domain.process.subdomain.definition.class.phase.md).** Fundamental phase skeleton for a process family.
- **[Project](class-domain.process.subdomain.project.class.project.md).** Work that follows a process.
- **[Project Part](class-domain.process.subdomain.project.class.project_part.md).** A language-specific part of a project.
- **[Definition::Stats Bucket](class-domain.process.subdomain.definition.class.stats_bucket.md).** A bucket used to group project statistics.
- **[Definition::Step](class-domain.process.subdomain.definition.class.step.md).** A step of a process script.


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
