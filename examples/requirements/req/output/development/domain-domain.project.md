[⇦ Development Process](model.md)

# Project

Data recorded for a running project that follows a process.

## Subdomains

The subdomains of this domain.

```mermaid
graph TD
subdomain_domain_project_subdomain_core["Core"]
subdomain_domain_project_subdomain_estimation["Estimation"]
subdomain_domain_project_subdomain_quality["Quality"]
subdomain_domain_project_subdomain_estimation -.-> subdomain_domain_project_subdomain_core
subdomain_domain_project_subdomain_quality -.-> subdomain_domain_project_subdomain_core

```

- **[Core](subdomain-domain.project.subdomain.core.md).** A running project: parts, cycles, schedule, tasks, and time logs.
- **[Estimation](subdomain-domain.project.subdomain.estimation.md).** Lines-of-code accounts and PROBE calculations recorded on a running project.
- **[Quality](subdomain-domain.project.subdomain.quality.md).** Defects, issues, process improvement proposals, and test cases recorded against projects.


