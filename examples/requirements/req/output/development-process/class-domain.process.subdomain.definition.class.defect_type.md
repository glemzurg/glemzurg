[⇦ Development Process](model.md) / [Process](domain-domain.process.md) / [Definition](subdomain-domain.process.subdomain.definition.md)

# Defect Type

A type of defect classified within a process family.




## Attributes

| Name | Rules | Nullable | TLA+ | Comments / Invariants |
| ---- | ----- | -------- | ---- | --------------------- |
| Num | _(unparsed)_ [0 .. unconstrained] at 1 unit | false |  | Order of this defect type within the family. |
| Name | _(unparsed)_ unconstrained | false |  | Unique among defect types of the same family. |
| Description | _(unparsed)_ unconstrained | false |  |  |
| Base Num | _(unparsed)_ [0 .. unconstrained] at 1 unit | false |  | Base classification number for this defect type. |




## Relations

The classes in this diagram.

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
style class_domain_process_subdomain_definition_class_defect_type stroke:#9370DB,stroke-width:3px
class_domain_process_subdomain_definition_class_family "1" --> "*" class_domain_process_subdomain_definition_class_defect_type : Has Defect Types<br/>{unique → Num}

```
- **[Defect Type](class-domain.process.subdomain.definition.class.defect_type.md).** A type of defect classified within a process family.
- **[Family](class-domain.process.subdomain.definition.class.family.md).** Core partitioning of the catalog.


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
