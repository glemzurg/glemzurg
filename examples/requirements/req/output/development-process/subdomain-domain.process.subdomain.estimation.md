[⇦ Development Process](model.md) / [Process](domain-domain.process.md)

# Estimation

Size and time estimates, including prior versions of each estimate.
## Classes

The classes of this subdomain.


```mermaid
---
config:
  class:
    hideEmptyMembersBox: true
---
classDiagram
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
namespace Definition {
class class_domain_process_subdomain_definition_class_family["Family"] {
            Name [key]
            Description
        }
class class_domain_process_subdomain_definition_class_language["Language"] {
            Name
            Description
        }
}
class_domain_process_subdomain_definition_class_family "1" --> "*" class_domain_process_subdomain_estimation_class_estimate : Has Estimates
class_domain_process_subdomain_estimation_class_estimate "*" --> "1" class_domain_process_subdomain_definition_class_language : Uses Language
class_domain_process_subdomain_estimation_class_estimate_historic "*" --> "1" class_domain_process_subdomain_definition_class_family : Belongs To Family
class_domain_process_subdomain_estimation_class_estimate_historic "*" --> "1" class_domain_process_subdomain_definition_class_language : Uses Language
class_domain_process_subdomain_estimation_class_estimate "1" --> "*" class_domain_process_subdomain_estimation_class_estimate_historic : Has History<br/>{unique → Version}

```

- **[Estimate](class-domain.process.subdomain.estimation.class.estimate.md).** An estimate of size or time, categorized by family, language, axis, and scope.
- **[Estimate Historic](class-domain.process.subdomain.estimation.class.estimate_historic.md).** A stored prior iteration of an estimate.
- **[Definition::Family](class-domain.process.subdomain.definition.class.family.md).** Core partitioning of the catalog.
- **[Definition::Language](class-domain.process.subdomain.definition.class.language.md).** A programming language used when estimating size or time in a process family.

[Model facts](subdomain-domain.process.subdomain.estimation-facts.md)


## Use Cases






