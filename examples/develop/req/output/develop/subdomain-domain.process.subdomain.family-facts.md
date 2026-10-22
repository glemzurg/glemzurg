[⇦ Development](model.md) / [Process](domain-domain.process.md) / [Family](subdomain-domain.process.subdomain.family.md)

# Model Facts — Family

Association multiplicity, association invariant, and index uniqueness constraints for this subdomain.

## Associations

- each Defect Type (subtype of) links to exactly one Defect Type; each Defect Type may link to any number of Defect Types (A defect that is a subtype of another defect.).
- each Definition::Module Template (uses language) links to exactly one Language; each Language may link to any number of Definition::Module Templates (Language this module template is for.).
- each Family (has defect types) links to any number of Defect Types; each Defect Type links to exactly one Family; each Family–Defect Type pairing has the uniqueness → Name and the uniqueness → Num (Defect types classified for this family.).
- each Family (has languages) links to any number of Languages; each Language links to exactly one Family; each Family–Language pairing has the uniqueness → Name (Programming languages used when estimating in this family.).
- each Family (has phases) links to any number of Phases; each Phase links to exactly one Family; each Family–Phase pairing has the uniqueness → Name and the uniqueness → Num (Ordered phase skeleton for this family.).
- each Family (has processes) links to any number of Process::Processes; each Process::Process links to exactly one Family; each Family–Process::Process pairing has the uniqueness → Name, Version, and Version Minor (Versioned processes that belong to this family.).
- each Process::Step (occurs in) links to exactly one Phase; each Phase may link to any number of Process::Steps (Phase of the family skeleton this step is performed in.).
- each Project::Core::Project (uses language) links to exactly one Language; each Language may link to any number of Project::Core::Projects (Language this project is implemented in.).
- each Project::Core::Project Part (current phase) may link to at most one Phase; each Phase may link to any number of Project::Core::Project Parts (Phase whose forms are currently open, when one is set.).
- each Project::Core::Project Part (uses language) links to exactly one Language; each Language may link to any number of Project::Core::Project Parts (Language this part is implemented in.).
- each Project::Core::Project Stat Phase (for phase) links to exactly one Phase; each Phase may link to any number of Project::Core::Project Stat Phases (Phase these statistics are for.).
- each Project::Estimation::Estimate (uses language) links to exactly one Language; each Language may link to any number of Project::Estimation::Estimates (Language this estimate is categorized by. The language must belong to the same family.).
- each Project::Estimation::Estimate Historic (uses language) links to exactly one Language; each Language may link to any number of Project::Estimation::Estimate Historics (Language copied from the estimate this snapshot belongs to.).
- each Project::Quality::Defect (injected in phase) links to exactly one Phase; each Phase may link to any number of Project::Quality::Defects (Phase where this defect was injected.).
- each Project::Quality::Defect (is of type) links to exactly one Defect Type; each Defect Type may link to any number of Project::Quality::Defects (The category of defect this is.).
- each Project::Quality::Defect (removed in phase) links to exactly one Phase; each Phase may link to any number of Project::Quality::Defects (Phase where this defect was removed.).
- each Project::Quality::Issue (injected in phase) links to exactly one Phase; each Phase may link to any number of Project::Quality::Issues (Phase where this issue was found.).

