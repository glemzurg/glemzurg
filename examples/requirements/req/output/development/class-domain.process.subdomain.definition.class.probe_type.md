[⇦ Development Process](model.md) / [Process](domain-domain.process.md) / [Definition](subdomain-domain.process.subdomain.definition.md)

# Probe Type

A PROBE object-type category used when listing added and new objects.




## Attributes

| Name | Rules | Nullable | TLA+ | Comments / Invariants |
| ---- | ----- | -------- | ---- | --------------------- |
| Number | _(unparsed)_ [0 .. unconstrained] at 1 unit | false |  |  |
| Name | _(unparsed)_ unconstrained | false |  |  |
| Description | _(unparsed)_ unconstrained | false |  |  |




## Relations

The classes in this diagram.

```mermaid
---
config:
  class:
    hideEmptyMembersBox: true
---
classDiagram
class class_domain_process_subdomain_definition_class_probe_type["Probe Type"] {
        Number
        Name
        Description
    }
namespace Estimation {
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
}
style class_domain_process_subdomain_definition_class_probe_type stroke:#9370DB,stroke-width:3px
class_domain_process_subdomain_estimation_class_estimate_probe_add_loc "*" --> "1" class_domain_process_subdomain_definition_class_probe_type : Of Type
class_domain_process_subdomain_estimation_class_estimate_probe_object_loc "*" --> "1" class_domain_process_subdomain_definition_class_probe_type : Of Type

```
- **[Estimation::Estimate Probe Add Loc](class-domain.process.subdomain.estimation.class.estimate_probe_add_loc.md).** An added-object line in a PROBE size estimate.
- **[Estimation::Estimate Probe Object Loc](class-domain.process.subdomain.estimation.class.estimate_probe_object_loc.md).** A new-object line in a PROBE size estimate.
- **[Probe Type](class-domain.process.subdomain.definition.class.probe_type.md).** A PROBE object-type category used when listing added and new objects.


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
