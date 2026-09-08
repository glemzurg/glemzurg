[⇦ Development Process](model.md) / [Process](domain-domain.process.md) / [Process](subdomain-domain.process.subdomain.process.md)

# Model Facts — Process

Association multiplicity, association invariant, and index uniqueness constraints for this subdomain.

## Associations

- each Definition::Module Template (follows process) may link to at most one Process; each Process may link to any number of Definition::Module Templates (Process this template is based on, when one is set.).
- each Family::Family (has processes) links to any number of Processes; each Process links to exactly one Family::Family; each Family::Family–Process pairing has the uniqueness → Name, Version, and Version Minor (Versioned processes that belong to this family.).
- each Process (has ancestor) may link to at most one Process; each Process may link to any number of Processes (Earlier process this version replaces, when one exists.).
- each Process (has scripts) links to any number of Scripts; each Script links to exactly one Process; each Process–Script pairing has the uniqueness → Num (Ordered scripts that make up this process.).
- each Project::Core::Project (current subphase) may link to at most one Step; each Step may link to any number of Project::Core::Projects (Planning step currently taking place, when one is set.).
- each Project::Core::Project (follows process) links to exactly one Process; each Process may link to any number of Project::Core::Projects (Process this project follows.).
- each Project::Core::Project Part (current subphase) may link to at most one Step; each Step may link to any number of Project::Core::Project Parts (Planning step currently taking place, when one is set.).
- each Project::Quality::Process Improvement Proposal (on process) links to exactly one Process; each Process may link to any number of Project::Quality::Process Improvement Proposals (Process this proposal is about.).
- each Project::Quality::Process Improvement Proposal (on subphase) links to exactly one Step; each Step may link to any number of Project::Quality::Process Improvement Proposals (Planning step this proposal is about.).
- each Project::Quality::Process Improvement Proposal (resolved in process) links to exactly one Process; each Process may link to any number of Project::Quality::Process Improvement Proposals (Process version that absorbed this proposal.).
- each Script (has steps) links to any number of Steps; each Step links to exactly one Script; each Script–Step pairing has the uniqueness → Num (Ordered steps of this script.).
- each Step (occurs in) links to exactly one Family::Phase; each Family::Phase may link to any number of Steps (Phase of the family skeleton this step is performed in.).

