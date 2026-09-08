[⇦ Development Process](model.md) / [Project](domain-domain.project.md) / [Estimation](subdomain-domain.project.subdomain.estimation.md)

# Estimate Probe

PROBE size and time calculation for a project in a phase. Reifies Project Has Probe Estimate For.




## Attributes

| Name | Rules | Nullable | TLA+ | Comments / Invariants |
| ---- | ----- | -------- | ---- | --------------------- |
| Base Loc | _(unparsed)_ [0 .. unconstrained] at 1 loc | false |  |  |
| Deleted Loc | _(unparsed)_ [0 .. unconstrained] at 1 loc | false |  |  |
| Modified Loc | _(unparsed)_ [0 .. unconstrained] at 1 loc | false |  |  |
| B0 Size | _(unparsed)_ [0 .. unconstrained] at 1 unit | false |  | Size regression intercept. |
| B1 Size | _(unparsed)_ [0 .. unconstrained] at 1 unit | false |  | Size regression slope. |
| B0 Time | _(unparsed)_ [0 .. unconstrained] at 1 unit | false |  | Time regression intercept. |
| B1 Time | _(unparsed)_ [0 .. unconstrained] at 1 unit | false |  | Time regression slope. |
| New Loc | _(unparsed)_ [0 .. unconstrained] at 1 loc | false |  |  |
| New Reuse Loc | _(unparsed)_ [0 .. unconstrained] at 1 loc | false |  |  |
| Estimated Time Min | _(unparsed)_ [0 .. unconstrained] at 1 minute | false |  |  |
| Upper Prediction Interval | _(unparsed)_ [0 .. unconstrained] at 1 unit | false |  |  |
| Lower Prediction Interval | _(unparsed)_ [0 .. unconstrained] at 1 unit | false |  |  |
| Prediction Interval Percent | _(unparsed)_ [0 .. 100] at 1 percent | false |  |  |
| Actual Base Loc | _(unparsed)_ [0 .. unconstrained] at 1 loc | false |  | SQL column acual_base_loc. |
| Actual Deleted Loc | _(unparsed)_ [0 .. unconstrained] at 1 loc | false |  | SQL column acual_deleted_loc. |
| Actual Modified Loc | _(unparsed)_ [0 .. unconstrained] at 1 loc | false |  | SQL column acual_modified_loc. |
| Loc Upper Prediction Interval 70 | _(unparsed)_ [0 .. unconstrained] at 1 loc | false |  |  |
| Loc Lower Prediction Interval 70 | _(unparsed)_ [0 .. unconstrained] at 1 loc | false |  |  |
| Time Upper Prediction Interval 70 | _(unparsed)_ [0 .. unconstrained] at 1 minute | false |  |  |
| Time Lower Prediction Interval 70 | _(unparsed)_ [0 .. unconstrained] at 1 minute | false |  |  |




## Relations

The classes in this diagram.

```mermaid
---
config:
  class:
    hideEmptyMembersBox: true
---
classDiagram
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
style class_domain_project_subdomain_estimation_class_estimate_probe stroke:#9370DB,stroke-width:3px
class assoc_cassociation_domain_project_subdomain_core_class_project_domain_process_subdomain_family_class_phase_has_probe_estimate_for["Has Probe Estimate For"]
<<association>> assoc_cassociation_domain_project_subdomain_core_class_project_domain_process_subdomain_family_class_phase_has_probe_estimate_for
style assoc_cassociation_domain_project_subdomain_core_class_project_domain_process_subdomain_family_class_phase_has_probe_estimate_for stroke:#333,stroke-dasharray:5 5
class_domain_project_subdomain_core_class_project "1" -- assoc_cassociation_domain_project_subdomain_core_class_project_domain_process_subdomain_family_class_phase_has_probe_estimate_for
    assoc_cassociation_domain_project_subdomain_core_class_project_domain_process_subdomain_family_class_phase_has_probe_estimate_for --> "*" class_domain_process_subdomain_family_class_phase
    class_domain_project_subdomain_estimation_class_estimate_probe .. assoc_cassociation_domain_project_subdomain_core_class_project_domain_process_subdomain_family_class_phase_has_probe_estimate_for

```
- **[Estimate Probe](class-domain.project.subdomain.estimation.class.estimate_probe.md).** PROBE size and time calculation for a project in a phase.
- **[Process::Family::Phase](class-domain.process.subdomain.family.class.phase.md).** Fundamental phase skeleton for a process family.
- **[Core::Project](class-domain.project.subdomain.core.class.project.md).** Work that follows a process.


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
