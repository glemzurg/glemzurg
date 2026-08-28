[⇦ Development Process](model.md) / [Process](domain-domain.process.md)

# Definition

The catalog of process families: phases, defect types, languages, methods, templates, processes, scripts, and steps.
## Classes

The classes of this subdomain.


```mermaid
---
config:
  class:
    hideEmptyMembersBox: true
---
classDiagram
class class_domain_process_subdomain_definition_class_defect_type["Defect Type"] {
        Num
        Name
        Description
        Base Num
    }
class class_domain_process_subdomain_definition_class_design_method["Design Method"] {
        Name
        Description
    }
class class_domain_process_subdomain_definition_class_family["Family"] {
        Name [key]
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
class class_domain_process_subdomain_definition_class_process["Process"] {
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
class class_domain_process_subdomain_definition_class_script["Script"] {
        Num
        Name
        Task Summary
        Purpose
        Entry Criteria
        Exit Criteria
        Cycle
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
namespace Estimation {
class class_domain_process_subdomain_estimation_class_actual_loc["Actual Loc"] {
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
class class_domain_process_subdomain_estimation_class_estimate["Estimate"] {
            Axis
            Scope
            Version
            Mean
            Variance
            Low
            High
            Actual
            Method
            Prediction Interval
            Comment
            Estimation Time
            Guess Lowest 80 Pred
            Guess Highest 80 Pred
            Sum Count
            Sum Mean
            Sum Variance
            Portion Portion
            Portion Mean
            Portion Variance
        }
class class_domain_process_subdomain_estimation_class_estimate_historic["Estimate Historic"] {
            Axis
            Scope
            Version
            Mean
            Variance
            Low
            High
            Actual
            Method
            Prediction Interval
            Comment
            Estimation Time
            Guess Lowest 80 Pred
            Guess Highest 80 Pred
            Sum Count
            Sum Mean
            Sum Variance
            Portion Portion
            Portion Mean
            Portion Variance
        }
class class_domain_process_subdomain_estimation_class_estimate_loc["Estimate Loc"] {
            Base
            New
            Changed
            Added
            Modified
            Deleted
            Reused
            Object Loc
        }
class class_domain_process_subdomain_estimation_class_estimate_probe["Estimate Probe"] {
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
class class_domain_process_subdomain_estimation_class_estimate_probe_add_loc["Estimate Probe Add Loc"] {
            Name
            Loc
            Actual Loc
        }
class class_domain_process_subdomain_estimation_class_estimate_probe_object_loc["Estimate Probe Object Loc"] {
            Name
            Loc Per Method
            Actual Loc
            For Reuse
        }
class class_domain_process_subdomain_estimation_class_estimate_probe_object_reused["Estimate Probe Object Reused"] {
            Name
            Loc
            Actual Loc
        }
}
namespace Project {
class class_domain_process_subdomain_project_class_phase_products_check["Phase Products Check"] {
            Satisfied
        }
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
class class_domain_process_subdomain_project_class_project_cycle_actual["Project Cycle Actual"] {
            Num
            Actual Pct Reuse
        }
class class_domain_process_subdomain_project_class_project_cycle_plan["Project Cycle Plan"] {
            Num
            Planned Pct Reuse
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
class class_domain_process_subdomain_project_class_project_stat_phase["Project Stat Phase"] {
            Estimate Minute
            Estimate Comment
        }
class class_domain_process_subdomain_project_class_task["Task"] {
            Num
            Name
            Planned Hours
            Pct Complete
        }
class class_domain_process_subdomain_project_class_time_log["Time Log"] {
            Cycle
            Start Time
            Stop Time
            Interruption Minutes
            Comments
        }
}
namespace Quality {
class class_domain_process_subdomain_quality_class_defect["Defect"] {
            Found Time
            Cycle
            Fix Minutes
            Description
        }
class class_domain_process_subdomain_quality_class_issue["Issue"] {
            Found Time
            Cycle
            Description
            Resolution Time
            Resolution
        }
class class_domain_process_subdomain_quality_class_pip["Process Improvement Proposal"] {
            Found Time
            Problem
            Proposal
            Resolved Time
        }
}
class assoc_domain_process_cassociation_subdomain_project_class_project_subdomain_definition_class_phase_checks_phase_products["Checks Phase Products"]
<<association>> assoc_domain_process_cassociation_subdomain_project_class_project_subdomain_definition_class_phase_checks_phase_products
class assoc_domain_process_cassociation_subdomain_project_class_project_subdomain_definition_class_phase_has_probe_estimate_for["Has Probe Estimate For"]
<<association>> assoc_domain_process_cassociation_subdomain_project_class_project_subdomain_definition_class_phase_has_probe_estimate_for
style assoc_domain_process_cassociation_subdomain_project_class_project_subdomain_definition_class_phase_checks_phase_products stroke:#333,stroke-dasharray:5 5
style assoc_domain_process_cassociation_subdomain_project_class_project_subdomain_definition_class_phase_has_probe_estimate_for stroke:#333,stroke-dasharray:5 5
class_domain_process_subdomain_definition_class_family "1" --> "*" class_domain_process_subdomain_estimation_class_estimate : Has Estimates
class_domain_process_subdomain_estimation_class_actual_loc "*" --> "1" class_domain_process_subdomain_definition_class_phase : For Phase
class_domain_process_subdomain_estimation_class_estimate "*" --> "1" class_domain_process_subdomain_definition_class_language : Uses Language
class_domain_process_subdomain_estimation_class_estimate_historic "*" --> "1" class_domain_process_subdomain_definition_class_family : Belongs To Family
class_domain_process_subdomain_estimation_class_estimate_historic "*" --> "1" class_domain_process_subdomain_definition_class_language : Uses Language
class_domain_process_subdomain_estimation_class_estimate_loc "*" --> "1" class_domain_process_subdomain_definition_class_phase : For Phase
class_domain_process_subdomain_estimation_class_estimate_probe_add_loc "*" --> "1" class_domain_process_subdomain_definition_class_phase : For Phase
class_domain_process_subdomain_estimation_class_estimate_probe_add_loc "*" --> "1" class_domain_process_subdomain_definition_class_probe_object_size : Of Size
class_domain_process_subdomain_estimation_class_estimate_probe_add_loc "*" --> "1" class_domain_process_subdomain_definition_class_probe_type : Of Type
class_domain_process_subdomain_estimation_class_estimate_probe_object_loc "*" --> "1" class_domain_process_subdomain_definition_class_phase : For Phase
class_domain_process_subdomain_estimation_class_estimate_probe_object_loc "*" --> "1" class_domain_process_subdomain_definition_class_probe_object_size : Of Size
class_domain_process_subdomain_estimation_class_estimate_probe_object_loc "*" --> "1" class_domain_process_subdomain_definition_class_probe_type : Of Type
class_domain_process_subdomain_estimation_class_estimate_probe_object_reused "*" --> "1" class_domain_process_subdomain_definition_class_phase : For Phase
class_domain_process_subdomain_project_class_project "*" --> "1" class_domain_process_subdomain_definition_class_design_method : Uses Design Method
class_domain_process_subdomain_project_class_project "*" --> "1" class_domain_process_subdomain_definition_class_language : Uses Language
class_domain_process_subdomain_project_class_project "*" --> "1" class_domain_process_subdomain_definition_class_method : Uses Size Estimation Method
class_domain_process_subdomain_project_class_project "*" --> "1" class_domain_process_subdomain_definition_class_method : Uses Time Estimation Method
class_domain_process_subdomain_project_class_project "*" --> "1" class_domain_process_subdomain_definition_class_module_template : Instantiates
class_domain_process_subdomain_project_class_project "*" -- assoc_domain_process_cassociation_subdomain_project_class_project_subdomain_definition_class_phase_checks_phase_products
    assoc_domain_process_cassociation_subdomain_project_class_project_subdomain_definition_class_phase_checks_phase_products --> "*" class_domain_process_subdomain_definition_class_phase
    class_domain_process_subdomain_project_class_phase_products_check .. assoc_domain_process_cassociation_subdomain_project_class_project_subdomain_definition_class_phase_checks_phase_products
class_domain_process_subdomain_project_class_project "*" --> "0..1" class_domain_process_subdomain_definition_class_phase : Current Phase
class_domain_process_subdomain_project_class_project "1" -- assoc_domain_process_cassociation_subdomain_project_class_project_subdomain_definition_class_phase_has_probe_estimate_for
    assoc_domain_process_cassociation_subdomain_project_class_project_subdomain_definition_class_phase_has_probe_estimate_for --> "*" class_domain_process_subdomain_definition_class_phase
    class_domain_process_subdomain_estimation_class_estimate_probe .. assoc_domain_process_cassociation_subdomain_project_class_project_subdomain_definition_class_phase_has_probe_estimate_for
class_domain_process_subdomain_project_class_project "*" --> "1" class_domain_process_subdomain_definition_class_process : Follows Process
class_domain_process_subdomain_project_class_project "*" --> "0..1" class_domain_process_subdomain_definition_class_stats_bucket : In Bucket
class_domain_process_subdomain_project_class_project "*" --> "0..1" class_domain_process_subdomain_definition_class_step : Current Subphase
class_domain_process_subdomain_project_class_project_cycle_actual "*" --> "1" class_domain_process_subdomain_definition_class_module_template : Instantiates
class_domain_process_subdomain_project_class_project_cycle_plan "*" --> "1" class_domain_process_subdomain_definition_class_module_template : Instantiates
class_domain_process_subdomain_project_class_project_part "*" --> "1" class_domain_process_subdomain_definition_class_design_method : Uses Design Method
class_domain_process_subdomain_project_class_project_part "*" --> "1" class_domain_process_subdomain_definition_class_language : Uses Language
class_domain_process_subdomain_project_class_project_part "*" --> "1" class_domain_process_subdomain_definition_class_method : Uses Size Estimation Method
class_domain_process_subdomain_project_class_project_part "*" --> "1" class_domain_process_subdomain_definition_class_method : Uses Time Estimation Method
class_domain_process_subdomain_project_class_project_part "*" --> "1" class_domain_process_subdomain_definition_class_module_template : Instantiates
class_domain_process_subdomain_project_class_project_part "*" --> "0..1" class_domain_process_subdomain_definition_class_phase : Current Phase
class_domain_process_subdomain_project_class_project_part "*" --> "0..1" class_domain_process_subdomain_definition_class_stats_bucket : In Bucket
class_domain_process_subdomain_project_class_project_part "*" --> "0..1" class_domain_process_subdomain_definition_class_step : Current Subphase
class_domain_process_subdomain_project_class_project_stat_phase "*" --> "0..1" class_domain_process_subdomain_definition_class_method : Uses Method
class_domain_process_subdomain_project_class_project_stat_phase "*" --> "1" class_domain_process_subdomain_definition_class_phase : For Phase
class_domain_process_subdomain_project_class_project_stat_phase "*" --> "1" class_domain_process_subdomain_definition_class_stats_bucket : In Bucket
class_domain_process_subdomain_project_class_task "*" --> "1" class_domain_process_subdomain_definition_class_phase : Occurs In
class_domain_process_subdomain_project_class_time_log "*" --> "1" class_domain_process_subdomain_definition_class_phase : Occurs In
class_domain_process_subdomain_quality_class_defect "*" --> "1" class_domain_process_subdomain_definition_class_phase : Injected In Phase
class_domain_process_subdomain_quality_class_defect "*" --> "1" class_domain_process_subdomain_definition_class_phase : Removed In Phase
class_domain_process_subdomain_quality_class_issue "*" --> "1" class_domain_process_subdomain_definition_class_phase : Injected In Phase
class_domain_process_subdomain_quality_class_pip "*" --> "1" class_domain_process_subdomain_definition_class_phase : On Phase
class_domain_process_subdomain_quality_class_pip "*" --> "1" class_domain_process_subdomain_definition_class_process : On Process
class_domain_process_subdomain_quality_class_pip "*" --> "1" class_domain_process_subdomain_definition_class_process : Resolved In Process
class_domain_process_subdomain_quality_class_pip "*" --> "1" class_domain_process_subdomain_definition_class_step : On Subphase
class_domain_process_subdomain_definition_class_family "1" --> "*" class_domain_process_subdomain_definition_class_defect_type : Has Defect Types<br/>{unique → Num}
class_domain_process_subdomain_definition_class_family "1" --> "*" class_domain_process_subdomain_definition_class_language : Has Languages<br/>{unique → Name}
class_domain_process_subdomain_definition_class_family "1" --> "*" class_domain_process_subdomain_definition_class_phase : Has Phases<br/>{unique → Num}
class_domain_process_subdomain_definition_class_family "1" --> "*" class_domain_process_subdomain_definition_class_process : Has Processes<br/>{unique → Name, Version, Version Minor}
class_domain_process_subdomain_definition_class_module_template "*" --> "1" class_domain_process_subdomain_definition_class_design_method : Uses Design Method
class_domain_process_subdomain_definition_class_module_template "*" --> "1" class_domain_process_subdomain_definition_class_language : Uses Language
class_domain_process_subdomain_definition_class_module_template "*" --> "1" class_domain_process_subdomain_definition_class_method : Uses Size Estimation Method
class_domain_process_subdomain_definition_class_module_template "*" --> "1" class_domain_process_subdomain_definition_class_method : Uses Time Estimation Method
class_domain_process_subdomain_definition_class_module_template "*" --> "0..1" class_domain_process_subdomain_definition_class_process : Follows Process
class_domain_process_subdomain_definition_class_process "*" --> "0..1" class_domain_process_subdomain_definition_class_process : Has Ancestor
class_domain_process_subdomain_definition_class_process "1" --> "*" class_domain_process_subdomain_definition_class_script : Has Scripts<br/>{unique → Num}
class_domain_process_subdomain_definition_class_script "1" --> "*" class_domain_process_subdomain_definition_class_step : Has Steps<br/>{unique → Num}
class_domain_process_subdomain_definition_class_step "*" --> "1" class_domain_process_subdomain_definition_class_phase : Occurs In

```

- **[Defect Type](class-domain.process.subdomain.definition.class.defect_type.md).** A type of defect classified within a process family.
- **[Design Method](class-domain.process.subdomain.definition.class.design_method.md).** A design template used when planning a project or module.
- **[Family](class-domain.process.subdomain.definition.class.family.md).** Core partitioning of the catalog.
- **[Language](class-domain.process.subdomain.definition.class.language.md).** A programming language used when estimating size or time in a process family.
- **[Method](class-domain.process.subdomain.definition.class.method.md).** A programming method used when recording phase statistics for a project.
- **[Module Template](class-domain.process.subdomain.definition.class.module_template.md).** Shared configuration for projects (and project parts) that follow a process.
- **[Phase](class-domain.process.subdomain.definition.class.phase.md).** Fundamental phase skeleton for a process family.
- **[Probe Object Size](class-domain.process.subdomain.definition.class.probe_object_size.md).** A relative size category used when estimating objects with PROBE.
- **[Probe Object Type](class-domain.process.subdomain.definition.class.probe_object_type.md).** A kind of object used when estimating with PROBE.
- **[Probe Type](class-domain.process.subdomain.definition.class.probe_type.md).** A PROBE object-type category used when listing added and new objects.
- **[Process](class-domain.process.subdomain.definition.class.process.md).** A versioned process to follow, owned by a family.
- **[Script](class-domain.process.subdomain.definition.class.script.md).** A step-by-step process script owned by a process.
- **[Stats Bucket](class-domain.process.subdomain.definition.class.stats_bucket.md).** A bucket used to group project statistics.
- **[Step](class-domain.process.subdomain.definition.class.step.md).** A step of a process script.
- **[Estimation::Actual Loc](class-domain.process.subdomain.estimation.class.actual_loc.md).** Actual lines-of-code account for a project.
- **[Quality::Defect](class-domain.process.subdomain.quality.class.defect.md).** A defect injected and removed in projects and phases.
- **[Estimation::Estimate](class-domain.process.subdomain.estimation.class.estimate.md).** An estimate of size or time, categorized by family, language, axis, and scope.
- **[Estimation::Estimate Historic](class-domain.process.subdomain.estimation.class.estimate_historic.md).** A stored prior iteration of an estimate.
- **[Estimation::Estimate Loc](class-domain.process.subdomain.estimation.class.estimate_loc.md).** Planned lines-of-code account for a project.
- **[Estimation::Estimate Probe](class-domain.process.subdomain.estimation.class.estimate_probe.md).** PROBE size and time calculation for a project in a phase.
- **[Estimation::Estimate Probe Add Loc](class-domain.process.subdomain.estimation.class.estimate_probe_add_loc.md).** An added-object line in a PROBE size estimate.
- **[Estimation::Estimate Probe Object Loc](class-domain.process.subdomain.estimation.class.estimate_probe_object_loc.md).** A new-object line in a PROBE size estimate.
- **[Estimation::Estimate Probe Object Reused](class-domain.process.subdomain.estimation.class.estimate_probe_object_reused.md).** A reused-object line in a PROBE size estimate.
- **[Quality::Issue](class-domain.process.subdomain.quality.class.issue.md).** An issue found in a project phase, with an optional resolution.
- **[Project::Phase Products Check](class-domain.process.subdomain.project.class.phase_products_check.md).** Whether a project's products for a phase are satisfied.
- **[Quality::Process Improvement Proposal](class-domain.process.subdomain.quality.class.pip.md).** A process improvement proposal raised on a project.
- **[Project::Project](class-domain.process.subdomain.project.class.project.md).** Work that follows a process.
- **[Project::Project Cycle Actual](class-domain.process.subdomain.project.class.project_cycle_actual.md).** Actual recording values for one cycle of a project.
- **[Project::Project Cycle Plan](class-domain.process.subdomain.project.class.project_cycle_plan.md).** Planned recording values for one cycle of a project.
- **[Project::Project Part](class-domain.process.subdomain.project.class.project_part.md).** A language-specific part of a project.
- **[Project::Project Stat Phase](class-domain.process.subdomain.project.class.project_stat_phase.md).** A per-phase statistical estimate on a project. stat_phase is Phase; bucket is Stats Bucket.
- **[Project::Task](class-domain.process.subdomain.project.class.task.md).** A planned task on a project, assigned to a phase and a schedule week.
- **[Project::Time Log](class-domain.process.subdomain.project.class.time_log.md).** A recorded interval of work on a project.

[Model facts](subdomain-domain.process.subdomain.definition-facts.md)


## Use Cases






