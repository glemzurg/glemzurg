[⇦ Development Process](model.md) / [Process](domain-domain.process.md)

# Quality

Defects, issues, process improvement proposals, and test cases recorded against projects.
## Classes

The classes of this subdomain.


```mermaid
---
config:
  class:
    hideEmptyMembersBox: true
---
classDiagram
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
namespace Definition {
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
            Size Unit
            Size K Unit
        }
class class_domain_process_subdomain_definition_class_step["Step"] {
            Num
            Name
            Tasks
        }
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
class_domain_process_subdomain_quality_class_defect "*" --> "1" class_domain_process_subdomain_definition_class_phase : Injected In Phase
class_domain_process_subdomain_quality_class_defect "*" --> "1" class_domain_process_subdomain_definition_class_phase : Removed In Phase
class_domain_process_subdomain_quality_class_defect "*" --> "1" class_domain_process_subdomain_project_class_project : Injected In
class_domain_process_subdomain_quality_class_defect "*" --> "1" class_domain_process_subdomain_project_class_project : Removed In
class_domain_process_subdomain_quality_class_issue "*" --> "1" class_domain_process_subdomain_definition_class_phase : Injected In Phase
class_domain_process_subdomain_quality_class_issue "*" --> "1" class_domain_process_subdomain_project_class_project : Injected In
class_domain_process_subdomain_quality_class_pip "*" --> "1" class_domain_process_subdomain_definition_class_phase : On Phase
class_domain_process_subdomain_quality_class_pip "*" --> "1" class_domain_process_subdomain_definition_class_process : On Process
class_domain_process_subdomain_quality_class_pip "*" --> "1" class_domain_process_subdomain_definition_class_process : Resolved In Process
class_domain_process_subdomain_quality_class_pip "*" --> "1" class_domain_process_subdomain_definition_class_step : On Subphase
class_domain_process_subdomain_quality_class_pip "*" --> "1" class_domain_process_subdomain_project_class_project : On Project
class_domain_process_subdomain_quality_class_test_case "*" --> "1" class_domain_process_subdomain_project_class_project : For Project
class_domain_process_subdomain_quality_class_test_case_result "*" --> "1" class_domain_process_subdomain_project_class_project : For Project
class_domain_process_subdomain_quality_class_defect "*" --> "0..1" class_domain_process_subdomain_quality_class_defect : Has Source
class_domain_process_subdomain_quality_class_test_case "1" --> "*" class_domain_process_subdomain_quality_class_test_case_result : Has Results

```

- **[Defect](class-domain.process.subdomain.quality.class.defect.md).** A defect injected and removed in projects and phases.
- **[Issue](class-domain.process.subdomain.quality.class.issue.md).** An issue found in a project phase, with an optional resolution.
- **[Process Improvement Proposal](class-domain.process.subdomain.quality.class.pip.md).** A process improvement proposal raised on a project.
- **[Test Case](class-domain.process.subdomain.quality.class.test_case.md).** A test case defined for a project.
- **[Test Case Result](class-domain.process.subdomain.quality.class.test_case_result.md).** A recorded run of a test case.
- **[Definition::Phase](class-domain.process.subdomain.definition.class.phase.md).** Fundamental phase skeleton for a process family.
- **[Definition::Process](class-domain.process.subdomain.definition.class.process.md).** A versioned process to follow, owned by a family.
- **[Project::Project](class-domain.process.subdomain.project.class.project.md).** Work that follows a process.
- **[Definition::Step](class-domain.process.subdomain.definition.class.step.md).** A step of a process script.

[Model facts](subdomain-domain.process.subdomain.quality-facts.md)


## Use Cases






