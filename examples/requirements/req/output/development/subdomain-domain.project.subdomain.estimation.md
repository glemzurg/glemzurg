[⇦ Development Process](model.md) / [Project](domain-domain.project.md)

# Estimation

Lines-of-code accounts and PROBE calculations recorded on a running project.
## Classes

The classes of this subdomain.


```mermaid
---
config:
  class:
    hideEmptyMembersBox: true
---
classDiagram
class class_domain_project_subdomain_estimation_class_actual_loc["Actual Loc"] {
        Cycle
        Base
        New
        Changed
        Added
        Modified
        Deleted
        Reused
        Object Loc
    }
class class_domain_project_subdomain_estimation_class_estimate_loc["Estimate Loc"] {
        Base
        New
        Changed
        Added
        Modified
        Deleted
        Reused
        Object Loc
    }
class class_domain_project_subdomain_estimation_class_estimate_probe["Estimate Probe"] {
        Base Loc
        Deleted Loc
        Modified Loc
        B0 Size
        B1 Size
        B0 Time
        B1 Time
        New Loc
        New Reuse Loc
        Estimated Time Min
        Upper Prediction Interval
        Lower Prediction Interval
        Prediction Interval Percent
        Actual Base Loc
        Actual Deleted Loc
        Actual Modified Loc
        Loc Upper Prediction Interval 70
        Loc Lower Prediction Interval 70
        Time Upper Prediction Interval 70
        Time Lower Prediction Interval 70
    }
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
class class_domain_project_subdomain_estimation_class_estimate_probe_object_reused["Estimate Probe Object Reused"] {
        Name
        Loc
        Actual Loc
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
class class_domain_process_subdomain_definition_class_phase["Phase"] {
            Num
            Name
            Description
        }
class class_domain_process_subdomain_definition_class_probe_object_size["Probe Object Size"] {
            Number
            Name
            Description
        }
class class_domain_process_subdomain_definition_class_probe_type["Probe Type"] {
            Number
            Name
            Description
        }
}
class assoc_cassociation_domain_project_subdomain_core_class_project_domain_process_subdomain_definition_class_phase_has_probe_estimate_for["Has Probe Estimate For"]
<<association>> assoc_cassociation_domain_project_subdomain_core_class_project_domain_process_subdomain_definition_class_phase_has_probe_estimate_for
style assoc_cassociation_domain_project_subdomain_core_class_project_domain_process_subdomain_definition_class_phase_has_probe_estimate_for stroke:#333,stroke-dasharray:5 5
class_domain_project_subdomain_core_class_project "1" -- assoc_cassociation_domain_project_subdomain_core_class_project_domain_process_subdomain_definition_class_phase_has_probe_estimate_for
    assoc_cassociation_domain_project_subdomain_core_class_project_domain_process_subdomain_definition_class_phase_has_probe_estimate_for --> "*" class_domain_process_subdomain_definition_class_phase
    class_domain_project_subdomain_estimation_class_estimate_probe .. assoc_cassociation_domain_project_subdomain_core_class_project_domain_process_subdomain_definition_class_phase_has_probe_estimate_for
class_domain_project_subdomain_estimation_class_actual_loc "*" --> "1" class_domain_process_subdomain_definition_class_phase : For Phase
class_domain_project_subdomain_estimation_class_estimate_loc "*" --> "1" class_domain_process_subdomain_definition_class_phase : For Phase
class_domain_project_subdomain_estimation_class_estimate_probe_add_loc "*" --> "1" class_domain_process_subdomain_definition_class_phase : For Phase
class_domain_project_subdomain_estimation_class_estimate_probe_add_loc "*" --> "1" class_domain_process_subdomain_definition_class_probe_object_size : Of Size
class_domain_project_subdomain_estimation_class_estimate_probe_add_loc "*" --> "1" class_domain_process_subdomain_definition_class_probe_type : Of Type
class_domain_project_subdomain_estimation_class_estimate_probe_object_loc "*" --> "1" class_domain_process_subdomain_definition_class_phase : For Phase
class_domain_project_subdomain_estimation_class_estimate_probe_object_loc "*" --> "1" class_domain_process_subdomain_definition_class_probe_object_size : Of Size
class_domain_project_subdomain_estimation_class_estimate_probe_object_loc "*" --> "1" class_domain_process_subdomain_definition_class_probe_type : Of Type
class_domain_project_subdomain_estimation_class_estimate_probe_object_reused "*" --> "1" class_domain_process_subdomain_definition_class_phase : For Phase
class_domain_project_subdomain_core_class_project "1" --> "0..1" class_domain_project_subdomain_estimation_class_actual_loc : Has Actual Loc
class_domain_project_subdomain_core_class_project "1" --> "0..1" class_domain_project_subdomain_estimation_class_estimate_loc : Has Loc Estimate
class_domain_project_subdomain_core_class_project "1" --> "*" class_domain_project_subdomain_estimation_class_estimate_probe_add_loc : Has Probe Add Loc
class_domain_project_subdomain_core_class_project "1" --> "*" class_domain_project_subdomain_estimation_class_estimate_probe_object_loc : Has Probe Object Loc
class_domain_project_subdomain_core_class_project "1" --> "*" class_domain_project_subdomain_estimation_class_estimate_probe_object_reused : Has Probe Object Reused

```

- **[Actual Loc](class-domain.project.subdomain.estimation.class.actual_loc.md).** Actual lines-of-code account for a project.
- **[Estimate Loc](class-domain.project.subdomain.estimation.class.estimate_loc.md).** Planned lines-of-code account for a project.
- **[Estimate Probe](class-domain.project.subdomain.estimation.class.estimate_probe.md).** PROBE size and time calculation for a project in a phase.
- **[Estimate Probe Add Loc](class-domain.project.subdomain.estimation.class.estimate_probe_add_loc.md).** An added-object line in a PROBE size estimate.
- **[Estimate Probe Object Loc](class-domain.project.subdomain.estimation.class.estimate_probe_object_loc.md).** A new-object line in a PROBE size estimate.
- **[Estimate Probe Object Reused](class-domain.project.subdomain.estimation.class.estimate_probe_object_reused.md).** A reused-object line in a PROBE size estimate.
- **[Process::Definition::Phase](class-domain.process.subdomain.definition.class.phase.md).** Fundamental phase skeleton for a process family.
- **[Process::Definition::Probe Object Size](class-domain.process.subdomain.definition.class.probe_object_size.md).** A relative size category used when estimating objects with PROBE.
- **[Process::Definition::Probe Type](class-domain.process.subdomain.definition.class.probe_type.md).** A PROBE object-type category used when listing added and new objects.
- **[Core::Project](class-domain.project.subdomain.core.class.project.md).** Work that follows a process.

[Model facts](subdomain-domain.project.subdomain.estimation-facts.md)


## Use Cases






