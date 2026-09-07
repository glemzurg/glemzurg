[⇦ Development Process](model.md) / [Process](domain-domain.process.md) / [Project](subdomain-domain.process.subdomain.project.md)

# Phase Products Check

Whether a project's products for a phase are satisfied.




## Attributes

| Name | Rules | Nullable | TLA+ | Comments / Invariants |
| ---- | ----- | -------- | ---- | --------------------- |
| Satisfied | _(unparsed)_ enum of TRUE, FALSE | false |  |  |




## Relations

The classes in this diagram.

```mermaid
---
config:
  class:
    hideEmptyMembersBox: true
---
classDiagram
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
namespace Definition {
class class_domain_process_subdomain_definition_class_phase["Phase"] {
            Num
            Name
            Description
        }
}
style class_domain_process_subdomain_project_class_phase_products_check stroke:#9370DB,stroke-width:3px
class assoc_domain_process_cassociation_subdomain_project_class_project_subdomain_definition_class_phase_checks_phase_products["Checks Phase Products"]
<<association>> assoc_domain_process_cassociation_subdomain_project_class_project_subdomain_definition_class_phase_checks_phase_products
style assoc_domain_process_cassociation_subdomain_project_class_project_subdomain_definition_class_phase_checks_phase_products stroke:#333,stroke-dasharray:5 5
class_domain_process_subdomain_project_class_project "*" -- assoc_domain_process_cassociation_subdomain_project_class_project_subdomain_definition_class_phase_checks_phase_products
    assoc_domain_process_cassociation_subdomain_project_class_project_subdomain_definition_class_phase_checks_phase_products --> "*" class_domain_process_subdomain_definition_class_phase
    class_domain_process_subdomain_project_class_phase_products_check .. assoc_domain_process_cassociation_subdomain_project_class_project_subdomain_definition_class_phase_checks_phase_products

```
- **[Definition::Phase](class-domain.process.subdomain.definition.class.phase.md).** Fundamental phase skeleton for a process family.
- **[Phase Products Check](class-domain.process.subdomain.project.class.phase_products_check.md).** Whether a project's products for a phase are satisfied.
- **[Project](class-domain.process.subdomain.project.class.project.md).** Work that follows a process.


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
