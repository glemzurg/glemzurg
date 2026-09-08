[⇦ Development Process](model.md) / [Process](domain-domain.process.md)

# Process

A versioned process: scripts and steps a family follows.
## Classes

The classes of this subdomain.


```mermaid
---
config:
  class:
    hideEmptyMembersBox: true
---
classDiagram
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
class class_domain_process_subdomain_process_class_script["Script"] {
        Num
        Name
        Task Summary
        Purpose
        Entry Criteria
        Exit Criteria
        Cycle
    }
class class_domain_process_subdomain_process_class_step["Step"] {
        Num
        Name
        Tasks
    }
namespace Definition {
class class_domain_process_subdomain_definition_class_module_template["Module Template"] {
            Name
            Description
        }
}
namespace Family {
class class_domain_process_subdomain_family_class_family["Family"] {
            Name [key]
            Description
        }
class class_domain_process_subdomain_family_class_phase["Phase"] {
            Num
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
}
namespace Project.Quality {
class class_domain_project_subdomain_quality_class_pip["Process Improvement Proposal"] {
            Found Time
            Problem
            Proposal
            Resolved Time
        }
}
class_domain_project_subdomain_core_class_project "*" --> "1" class_domain_process_subdomain_process_class_process : Follows Process
class_domain_project_subdomain_core_class_project "*" --> "0..1" class_domain_process_subdomain_process_class_step : Current Subphase
class_domain_project_subdomain_core_class_project_part "*" --> "0..1" class_domain_process_subdomain_process_class_step : Current Subphase
class_domain_project_subdomain_quality_class_pip "*" --> "1" class_domain_process_subdomain_process_class_process : On Process
class_domain_project_subdomain_quality_class_pip "*" --> "1" class_domain_process_subdomain_process_class_process : Resolved In Process
class_domain_project_subdomain_quality_class_pip "*" --> "1" class_domain_process_subdomain_process_class_step : On Subphase
class_domain_process_subdomain_definition_class_module_template "*" --> "0..1" class_domain_process_subdomain_process_class_process : Follows Process
class_domain_process_subdomain_family_class_family "1" --> "*" class_domain_process_subdomain_process_class_process : Has Processes<br/>{unique → Name, Version, Version Minor}
class_domain_process_subdomain_process_class_step "*" --> "1" class_domain_process_subdomain_family_class_phase : Occurs In
class_domain_process_subdomain_process_class_process "*" --> "0..1" class_domain_process_subdomain_process_class_process : Has Ancestor
class_domain_process_subdomain_process_class_process "1" --> "*" class_domain_process_subdomain_process_class_script : Has Scripts<br/>{unique → Num}
class_domain_process_subdomain_process_class_script "1" --> "*" class_domain_process_subdomain_process_class_step : Has Steps<br/>{unique → Num}

```

- **[Process](class-domain.process.subdomain.process.class.process.md).** A versioned process to follow, owned by a family.
- **[Script](class-domain.process.subdomain.process.class.script.md).** A step-by-step process script owned by a process.
- **[Step](class-domain.process.subdomain.process.class.step.md).** A step of a process script.
- **[Family::Family](class-domain.process.subdomain.family.class.family.md).** Core partitioning of the catalog.
- **[Definition::Module Template](class-domain.process.subdomain.definition.class.module_template.md).** Shared configuration for projects (and project parts) that follow a process.
- **[Family::Phase](class-domain.process.subdomain.family.class.phase.md).** Fundamental phase skeleton for a process family.
- **[Project::Quality::Process Improvement Proposal](class-domain.project.subdomain.quality.class.pip.md).** A process improvement proposal raised on a project.
- **[Project::Core::Project](class-domain.project.subdomain.core.class.project.md).** Work that follows a process.
- **[Project::Core::Project Part](class-domain.project.subdomain.core.class.project_part.md).** A language-specific part of a project.

[Model facts](subdomain-domain.process.subdomain.process-facts.md)


## Use Cases






