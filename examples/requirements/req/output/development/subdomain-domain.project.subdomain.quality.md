[⇦ Development](model.md) / [Project](domain-domain.project.md)

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
class class_domain_project_subdomain_quality_class_defect["Defect"] {
        Found Time
        Cycle
        Fix Minutes
        Description
        Prevention
        Test Defect
    }
class class_domain_project_subdomain_quality_class_issue["Issue"] {
        Found Time
        Cycle
        Description
        Resolution Time
        Resolution
    }
class class_domain_project_subdomain_quality_class_pip["Process Improvement Proposal"] {
        Found Time
        Problem
        Proposal
        Resolved Time
    }
class class_domain_project_subdomain_quality_class_test_case["Test Case"] {
        Found Time
        Objective
        Description
        Conditions
        Expected
    }
class class_domain_project_subdomain_quality_class_test_case_result["Test Case Result"] {
        Run Time
        Actual
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
namespace Process.Process {
class class_domain_process_subdomain_process_class_process["Process"] {
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
class class_domain_process_subdomain_process_class_step["Step"] {
            Num
            Name
            Tasks
        }
}
class_domain_project_subdomain_quality_class_defect "*" --> "1" class_domain_process_subdomain_family_class_phase : Injected In Phase
class_domain_project_subdomain_quality_class_defect "*" --> "1" class_domain_process_subdomain_family_class_phase : Removed In Phase
class_domain_project_subdomain_quality_class_issue "*" --> "1" class_domain_process_subdomain_family_class_phase : Injected In Phase
class_domain_project_subdomain_quality_class_pip "*" --> "1" class_domain_process_subdomain_family_class_phase : On Phase
class_domain_project_subdomain_quality_class_pip "*" --> "1" class_domain_process_subdomain_process_class_process : On Process
class_domain_project_subdomain_quality_class_pip "*" --> "1" class_domain_process_subdomain_process_class_process : Resolved In Process
class_domain_project_subdomain_quality_class_pip "*" --> "1" class_domain_process_subdomain_process_class_step : On Subphase
class_domain_project_subdomain_quality_class_defect "*" --> "1" class_domain_project_subdomain_core_class_project : Injected In
class_domain_project_subdomain_quality_class_defect "*" --> "1" class_domain_project_subdomain_core_class_project : Removed In
class_domain_project_subdomain_quality_class_issue "*" --> "1" class_domain_project_subdomain_core_class_project : Injected In
class_domain_project_subdomain_quality_class_pip "*" --> "1" class_domain_project_subdomain_core_class_project : On Project
class_domain_project_subdomain_quality_class_test_case "*" --> "1" class_domain_project_subdomain_core_class_project : For Project
class_domain_project_subdomain_quality_class_test_case_result "*" --> "1" class_domain_project_subdomain_core_class_project : For Project
class_domain_project_subdomain_quality_class_defect "*" --> "0..1" class_domain_project_subdomain_quality_class_defect : Has Source
class_domain_project_subdomain_quality_class_test_case "1" --> "*" class_domain_project_subdomain_quality_class_test_case_result : Has Results

```

- **[Defect](class-domain.project.subdomain.quality.class.defect.md).** A defect injected and removed in projects and phases.
- **[Issue](class-domain.project.subdomain.quality.class.issue.md).** An issue found in a project phase, with an optional resolution.
- **[Process Improvement Proposal](class-domain.project.subdomain.quality.class.pip.md).** A process improvement proposal raised on a project.
- **[Test Case](class-domain.project.subdomain.quality.class.test_case.md).** A test case defined for a project.
- **[Test Case Result](class-domain.project.subdomain.quality.class.test_case_result.md).** A recorded run of a test case.
- **[Process::Family::Phase](class-domain.process.subdomain.family.class.phase.md).** Fundamental phase skeleton for all the processes in a family.
- **[Process::Process::Process](class-domain.process.subdomain.process.class.process.md).** A versioned process to follow, owned by a family.
- **[Core::Project](class-domain.project.subdomain.core.class.project.md).** Work that follows a process.
- **[Process::Process::Step](class-domain.process.subdomain.process.class.step.md).** A step of a process script.

[Model facts](subdomain-domain.project.subdomain.quality-facts.md)


## Use Cases






