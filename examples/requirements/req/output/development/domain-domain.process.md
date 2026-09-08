[⇦ Development Process](model.md)

# Process

The definition of a process: families, scripts, methods, templates, and family-level estimates.

## Subdomains

The subdomains of this domain.

```mermaid
graph TD
subdomain_domain_process_subdomain_definition["Definition"]
subdomain_domain_process_subdomain_estimate["Estimate"]
subdomain_domain_process_subdomain_family["Family"]
subdomain_domain_process_subdomain_method["Method"]
subdomain_domain_process_subdomain_process["Process"]
subdomain_domain_process_subdomain_definition -.-> subdomain_domain_process_subdomain_estimate
subdomain_domain_process_subdomain_definition -.-> subdomain_domain_process_subdomain_family
subdomain_domain_process_subdomain_definition -.-> subdomain_domain_process_subdomain_method
subdomain_domain_process_subdomain_definition -.-> subdomain_domain_process_subdomain_process
subdomain_domain_process_subdomain_process -.-> subdomain_domain_process_subdomain_family

```

- **[Definition](subdomain-domain.process.subdomain.definition.md).** Templates and PROBE catalogs used when following a process.
- **[Estimate](subdomain-domain.process.subdomain.estimate.md).** Programming methods used when estimating size and time.
- **[Family](subdomain-domain.process.subdomain.family.md).** A process family and the phases, defect types, and languages that belong to it.
- **[Method](subdomain-domain.process.subdomain.method.md).** Design templates used when planning a project or module.
- **[Process](subdomain-domain.process.subdomain.process.md).** A versioned process: scripts and steps a family follows.


