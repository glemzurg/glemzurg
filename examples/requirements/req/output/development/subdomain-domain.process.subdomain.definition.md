[⇦ Development Process](model.md) / [Process](domain-domain.process.md)

# Definition

Templates and PROBE catalogs used when following a process.
## Classes

The classes of this subdomain.


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
class class_domain_process_subdomain_definition_class_probe_object_size["Probe Object Size"] {
        Number
        Name
        Description
    }
class class_domain_process_subdomain_definition_class_probe_object_type["Probe Object Type"] {
        Number
        Name
        Description
    }
class class_domain_process_subdomain_definition_class_probe_type["Probe Type"] {
        Number
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
namespace Project.Estimation {
class class_domain_project_subdomain_estimation_class_estimate_probe_add_loc["Estimate Probe Add Loc"] {
            Name
            Loc
            Actual Loc
        }
class class_domain_project_subdomain_estimation_class_estimate_probe_object_loc["Estimate Probe Object Loc"] {
            Name
            Loc Per Method
            Actual Loc
            For Reuse
        }
}
class_domain_project_subdomain_core_class_project "*" --> "1" class_domain_process_subdomain_definition_class_module_template : Instantiates
class_domain_project_subdomain_core_class_project_part "*" --> "1" class_domain_process_subdomain_definition_class_module_template : Instantiates
class_domain_project_subdomain_cycle_class_project_cycle_actual "*" --> "1" class_domain_process_subdomain_definition_class_module_template : Instantiates
class_domain_project_subdomain_cycle_class_project_cycle_plan "*" --> "1" class_domain_process_subdomain_definition_class_module_template : Instantiates
class_domain_project_subdomain_estimation_class_estimate_probe_add_loc "*" --> "1" class_domain_process_subdomain_definition_class_probe_object_size : Of Size
class_domain_project_subdomain_estimation_class_estimate_probe_add_loc "*" --> "1" class_domain_process_subdomain_definition_class_probe_type : Of Type
class_domain_project_subdomain_estimation_class_estimate_probe_object_loc "*" --> "1" class_domain_process_subdomain_definition_class_probe_object_size : Of Size
class_domain_project_subdomain_estimation_class_estimate_probe_object_loc "*" --> "1" class_domain_process_subdomain_definition_class_probe_type : Of Type
class_domain_process_subdomain_definition_class_module_template "*" --> "1" class_domain_process_subdomain_estimate_class_method : Uses Size Estimation Method
class_domain_process_subdomain_definition_class_module_template "*" --> "1" class_domain_process_subdomain_estimate_class_method : Uses Time Estimation Method
class_domain_process_subdomain_definition_class_module_template "*" --> "1" class_domain_process_subdomain_family_class_language : Uses Language
class_domain_process_subdomain_definition_class_module_template "*" --> "1" class_domain_process_subdomain_method_class_design_method : Uses Design Method
class_domain_process_subdomain_definition_class_module_template "*" --> "0..1" class_domain_process_subdomain_process_class_process : Follows Process

```

- **[Module Template](class-domain.process.subdomain.definition.class.module_template.md).** Shared configuration for projects (and project parts) that follow a process.
- **[Probe Object Size](class-domain.process.subdomain.definition.class.probe_object_size.md).** A relative size category used when estimating objects with PROBE.
- **[Probe Object Type](class-domain.process.subdomain.definition.class.probe_object_type.md).** A kind of object used when estimating with PROBE.
- **[Probe Type](class-domain.process.subdomain.definition.class.probe_type.md).** A PROBE object-type category used when listing added and new objects.
- **[Method::Design Method](class-domain.process.subdomain.method.class.design_method.md).** A design template used when planning a project or module.
- **[Project::Estimation::Estimate Probe Add Loc](class-domain.project.subdomain.estimation.class.estimate_probe_add_loc.md).** An added-object line in a PROBE size estimate.
- **[Project::Estimation::Estimate Probe Object Loc](class-domain.project.subdomain.estimation.class.estimate_probe_object_loc.md).** A new-object line in a PROBE size estimate.
- **[Family::Language](class-domain.process.subdomain.family.class.language.md).** A programming language used when estimating size or time in a process family.
- **[Estimate::Method](class-domain.process.subdomain.estimate.class.method.md).** A programming method used when recording phase statistics for a project.
- **[Process::Process](class-domain.process.subdomain.process.class.process.md).** A versioned process to follow, owned by a family.
- **[Project::Core::Project](class-domain.project.subdomain.core.class.project.md).** Work that follows a process.
- **[Project::Cycle::Project Cycle Actual](class-domain.project.subdomain.cycle.class.project_cycle_actual.md).** Actual recording values for one cycle of a project.
- **[Project::Cycle::Project Cycle Plan](class-domain.project.subdomain.cycle.class.project_cycle_plan.md).** Planned recording values for one cycle of a project.
- **[Project::Core::Project Part](class-domain.project.subdomain.core.class.project_part.md).** A language-specific part of a project.

[Model facts](subdomain-domain.process.subdomain.definition-facts.md)


## Use Cases






