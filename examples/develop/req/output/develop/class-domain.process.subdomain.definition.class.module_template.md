[⇦ Development](model.md) / [Process](domain-domain.process.md) / [Definition](subdomain-domain.process.subdomain.definition.md)

# Module Template

Shared configuration for projects (and project parts) that follow a process.




## Attributes

| Name | Rules | Nullable | TLA+ | Comments / Invariants |
| ---- | ----- | -------- | ---- | --------------------- |
| Name | _(unparsed)_ unconstrained | false |  |  |
| Description | _(unparsed)_ unconstrained | true |  |  |




## Relations

The classes in this diagram.

```mermaid
---
config:
  class:
    hideEmptyMembersBox: true
---
classDiagram
class class_domain_process_subdomain_definition_class_module_template["Module Template"] {
        Name
        Description
    }
namespace Estimate {
class class_domain_process_subdomain_estimate_class_method["Method"] {
            Name [key]
            Description
        }
}
namespace Family {
class class_domain_process_subdomain_family_class_language["Language"] {
            Name
            Description
        }
}
namespace Method {
class class_domain_process_subdomain_method_class_design_method["Design Method"] {
            Name
            Description
        }
}
namespace Process {
class class_domain_process_subdomain_process_class_process["Process"] {
            Name
            Version
            Version Minor
            Purpose
            Entry Criteria
            Exit Criteria
            Script Lock
            Size Unit
            Size K Unit
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
}
namespace Project.Cycle {
class class_domain_project_subdomain_cycle_class_project_cycle_actual["Project Cycle Actual"] {
            Num
            Actual Pct Reuse
        }
class class_domain_project_subdomain_cycle_class_project_cycle_plan["Project Cycle Plan"] {
            Num
            Planned Pct Reuse
        }
}
style class_domain_process_subdomain_definition_class_module_template stroke:#9370DB,stroke-width:3px
class_domain_project_subdomain_core_class_project "*" --> "1" class_domain_process_subdomain_definition_class_module_template : Instantiates
class_domain_project_subdomain_core_class_project_part "*" --> "1" class_domain_process_subdomain_definition_class_module_template : Instantiates
class_domain_project_subdomain_cycle_class_project_cycle_actual "*" --> "1" class_domain_process_subdomain_definition_class_module_template : Instantiates
class_domain_project_subdomain_cycle_class_project_cycle_plan "*" --> "1" class_domain_process_subdomain_definition_class_module_template : Instantiates
class_domain_process_subdomain_definition_class_module_template "*" --> "1" class_domain_process_subdomain_estimate_class_method : Uses Size Estimation Method
class_domain_process_subdomain_definition_class_module_template "*" --> "1" class_domain_process_subdomain_estimate_class_method : Uses Time Estimation Method
class_domain_process_subdomain_definition_class_module_template "*" --> "1" class_domain_process_subdomain_family_class_language : Uses Language
class_domain_process_subdomain_definition_class_module_template "*" --> "1" class_domain_process_subdomain_method_class_design_method : Uses Design Method
class_domain_process_subdomain_definition_class_module_template "*" --> "0..1" class_domain_process_subdomain_process_class_process : Follows Process

```
- **[Method::Design Method](class-domain.process.subdomain.method.class.design_method.md).** A design template used when planning a project or module.
- **[Family::Language](class-domain.process.subdomain.family.class.language.md).** A programming language used when estimating size or time in a process family.
- **[Estimate::Method](class-domain.process.subdomain.estimate.class.method.md).** A programming method used when recording phase statistics for a project.
- **[Module Template](class-domain.process.subdomain.definition.class.module_template.md).** Shared configuration for projects (and project parts) that follow a process.
- **[Process::Process](class-domain.process.subdomain.process.class.process.md).** A versioned process to follow, owned by a family.
- **[Project::Core::Project](class-domain.project.subdomain.core.class.project.md).** Work that follows a process.
- **[Project::Cycle::Project Cycle Actual](class-domain.project.subdomain.cycle.class.project_cycle_actual.md).** Actual recording values for one cycle of a project.
- **[Project::Cycle::Project Cycle Plan](class-domain.project.subdomain.cycle.class.project_cycle_plan.md).** Planned recording values for one cycle of a project.
- **[Project::Core::Project Part](class-domain.project.subdomain.core.class.project_part.md).** A language-specific part of a project.


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
