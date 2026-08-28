[⇦ Development Process](model.md) / [Process](domain-domain.process.md)

# Definition

The catalog of process families: phases, defect types, languages, processes, scripts, and steps.
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
class class_domain_process_subdomain_definition_class_family["Family"] {
        Name [key]
        Description
    }
class class_domain_process_subdomain_definition_class_language["Language"] {
        Name
        Description
    }
class class_domain_process_subdomain_definition_class_phase["Phase"] {
        Num
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
class class_domain_process_subdomain_definition_class_step["Step"] {
        Num
        Name
        Tasks
    }
namespace Estimation {
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
}
class_domain_process_subdomain_definition_class_family "1" --> "*" class_domain_process_subdomain_estimation_class_estimate : Has Estimates
class_domain_process_subdomain_estimation_class_estimate "*" --> "1" class_domain_process_subdomain_definition_class_language : Uses Language
class_domain_process_subdomain_estimation_class_estimate_historic "*" --> "1" class_domain_process_subdomain_definition_class_family : Belongs To Family
class_domain_process_subdomain_estimation_class_estimate_historic "*" --> "1" class_domain_process_subdomain_definition_class_language : Uses Language
class_domain_process_subdomain_definition_class_family "1" --> "*" class_domain_process_subdomain_definition_class_defect_type : Has Defect Types<br/>{unique → Num}
class_domain_process_subdomain_definition_class_family "1" --> "*" class_domain_process_subdomain_definition_class_language : Has Languages<br/>{unique → Name}
class_domain_process_subdomain_definition_class_family "1" --> "*" class_domain_process_subdomain_definition_class_phase : Has Phases<br/>{unique → Num}
class_domain_process_subdomain_definition_class_family "1" --> "*" class_domain_process_subdomain_definition_class_process : Has Processes<br/>{unique → Name, Version, Version Minor}
class_domain_process_subdomain_definition_class_process "*" --> "0..1" class_domain_process_subdomain_definition_class_process : Has Ancestor
class_domain_process_subdomain_definition_class_process "1" --> "*" class_domain_process_subdomain_definition_class_script : Has Scripts<br/>{unique → Num}
class_domain_process_subdomain_definition_class_script "1" --> "*" class_domain_process_subdomain_definition_class_step : Has Steps<br/>{unique → Num}
class_domain_process_subdomain_definition_class_step "*" --> "1" class_domain_process_subdomain_definition_class_phase : Occurs In

```

- **[Defect Type](class-domain.process.subdomain.definition.class.defect_type.md).** A type of defect classified within a process family.
- **[Family](class-domain.process.subdomain.definition.class.family.md).** Core partitioning of the catalog.
- **[Language](class-domain.process.subdomain.definition.class.language.md).** A programming language used when estimating size or time in a process family.
- **[Phase](class-domain.process.subdomain.definition.class.phase.md).** Fundamental phase skeleton for a process family.
- **[Process](class-domain.process.subdomain.definition.class.process.md).** A versioned process to follow, owned by a family.
- **[Script](class-domain.process.subdomain.definition.class.script.md).** A step-by-step process script owned by a process.
- **[Step](class-domain.process.subdomain.definition.class.step.md).** A step of a process script.
- **[Estimation::Estimate](class-domain.process.subdomain.estimation.class.estimate.md).** An estimate of size or time, categorized by family, language, axis, and scope.
- **[Estimation::Estimate Historic](class-domain.process.subdomain.estimation.class.estimate_historic.md).** A stored prior iteration of an estimate.

[Model facts](subdomain-domain.process.subdomain.definition-facts.md)


## Use Cases






