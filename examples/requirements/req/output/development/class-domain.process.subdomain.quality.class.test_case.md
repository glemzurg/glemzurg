[⇦ Development Process](model.md) / [Process](domain-domain.process.md) / [Quality](subdomain-domain.process.subdomain.quality.md)

# Test Case

A test case defined for a project.




## Attributes

| Name | Rules | Nullable | TLA+ | Comments / Invariants |
| ---- | ----- | -------- | ---- | --------------------- |
| Found Time | _(unparsed)_ datetime | false |  |  |
| Objective | _(unparsed)_ unconstrained | false |  |  |
| Description | _(unparsed)_ unconstrained | false |  |  |
| Conditions | _(unparsed)_ unconstrained | false |  |  |
| Expected | _(unparsed)_ unconstrained | false |  |  |




## Relations

The classes in this diagram.

```mermaid
---
config:
  class:
    hideEmptyMembersBox: true
---
classDiagram
class class_domain_process_subdomain_quality_class_test_case["Test Case"] {
        Found Time
        Objective
        Description
        Conditions
        Expected
    }
class class_domain_process_subdomain_quality_class_test_case_result["Test Case Result"] {
        Run Time
        Actual
    }
namespace Project {
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
}
style class_domain_process_subdomain_quality_class_test_case stroke:#9370DB,stroke-width:3px
class_domain_process_subdomain_quality_class_test_case "*" --> "1" class_domain_process_subdomain_project_class_project : For Project
class_domain_process_subdomain_quality_class_test_case "1" --> "*" class_domain_process_subdomain_quality_class_test_case_result : Has Results

```
- **[Project::Project](class-domain.process.subdomain.project.class.project.md).** Work that follows a process.
- **[Test Case](class-domain.process.subdomain.quality.class.test_case.md).** A test case defined for a project.
- **[Test Case Result](class-domain.process.subdomain.quality.class.test_case_result.md).** A recorded run of a test case.


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
