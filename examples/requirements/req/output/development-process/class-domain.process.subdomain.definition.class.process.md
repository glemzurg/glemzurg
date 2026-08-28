[⇦ Development Process](model.md) / [Process](domain-domain.process.md) / [Definition](subdomain-domain.process.subdomain.definition.md)

# Process

A versioned process to follow, owned by a family. May replace an earlier process as its ancestor.




## Attributes

| Name | Rules | Nullable | TLA+ | Comments / Invariants |
| ---- | ----- | -------- | ---- | --------------------- |
| Name | _(unparsed)_ unconstrained | false |  | Unique together with version and version minor within the family. |
| Version | _(unparsed)_ [0 .. unconstrained] at 1 unit | false |  | Major version of this process. |
| Version Minor | _(unparsed)_ [0 .. unconstrained] at 1 unit | false |  | Minor version of this process. Defaults to 0. |
| Purpose | _(unparsed)_ unconstrained | false |  |  |
| Entry Criteria | _(unparsed)_ unconstrained | false |  |  |
| Exit Criteria | _(unparsed)_ unconstrained | false |  |  |
| Script Lock | _(unparsed)_ enum of TRUE, FALSE | false |  | When TRUE, scripts of this process are locked. Defaults to FALSE. |

## Invariants

- Script name is unique among scripts of this process.
    - **∀ s ∈ self.HasScripts : _FiniteSets!Cardinality({q ∈ self.HasScripts : q.name = s.name}) = 1**



## Relations

The classes in this diagram.

```mermaid
---
config:
  class:
    hideEmptyMembersBox: true
---
classDiagram
class class_domain_process_subdomain_definition_class_family["Family"] {
        Name [key]
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
style class_domain_process_subdomain_definition_class_process stroke:#9370DB,stroke-width:3px
class_domain_process_subdomain_definition_class_family "1" --> "*" class_domain_process_subdomain_definition_class_process : Has Processes<br/>{unique → Name, Version, Version Minor}
class_domain_process_subdomain_definition_class_process "*" --> "0..1" class_domain_process_subdomain_definition_class_process : Has Ancestor
class_domain_process_subdomain_definition_class_process "1" --> "*" class_domain_process_subdomain_definition_class_script : Has Scripts<br/>{unique → Num}

```
- **[Family](class-domain.process.subdomain.definition.class.family.md).** Core partitioning of the catalog.
- **[Process](class-domain.process.subdomain.definition.class.process.md).** A versioned process to follow, owned by a family.
- **[Script](class-domain.process.subdomain.definition.class.script.md).** A step-by-step process script owned by a process.


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
