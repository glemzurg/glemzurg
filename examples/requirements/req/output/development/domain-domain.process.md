[⇦ Development Process](model.md)

# Process

Process families, projects that follow them, estimates, and quality records.

## Subdomains

The subdomains of this domain.

```mermaid
graph TD
subdomain_domain_process_subdomain_definition["Definition"]
subdomain_domain_process_subdomain_estimation["Estimation"]
subdomain_domain_process_subdomain_project["Project"]
subdomain_domain_process_subdomain_quality["Quality"]
subdomain_domain_process_subdomain_estimation -.-> subdomain_domain_process_subdomain_definition
subdomain_domain_process_subdomain_estimation -.-> subdomain_domain_process_subdomain_project
subdomain_domain_process_subdomain_project -.-> subdomain_domain_process_subdomain_definition
subdomain_domain_process_subdomain_quality -.-> subdomain_domain_process_subdomain_definition
subdomain_domain_process_subdomain_quality -.-> subdomain_domain_process_subdomain_project

```

- **[Definition](subdomain-domain.process.subdomain.definition.md).** The catalog of process families: phases, defect types, languages, methods, templates, processes, scripts, and steps.
- **[Estimation](subdomain-domain.process.subdomain.estimation.md).** Size and time estimates, PROBE calculations, and lines-of-code accounts.
- **[Project](subdomain-domain.process.subdomain.project.md).** A project following a process: parts, cycles, schedule, tasks, and time logs.
- **[Quality](subdomain-domain.process.subdomain.quality.md).** Defects, issues, process improvement proposals, and test cases recorded against projects.


