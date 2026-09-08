[⇦ Development Process](model.md) / [Process](domain-domain.process.md) / [Family](subdomain-domain.process.subdomain.family.md)

# Model Facts — Family

Association multiplicity, association invariant, and index uniqueness constraints for this subdomain.

## Associations

- each Definition::Module Template (uses language) links to exactly one Language; each Language may link to any number of Definition::Module Templates (Language this module template is for.).
- each Definition::Step (occurs in) links to exactly one Phase; each Phase may link to any number of Definition::Steps (Phase of the family skeleton this step is performed in.).
- each Estimation::Estimate (uses language) links to exactly one Language; each Language may link to any number of Estimation::Estimates (Language this estimate is categorized by. The language must belong to the same family.).
- each Estimation::Estimate Historic (belongs to family) links to exactly one Family; each Family may link to any number of Estimation::Estimate Historics (Family copied from the estimate this snapshot belongs to.).
- each Estimation::Estimate Historic (uses language) links to exactly one Language; each Language may link to any number of Estimation::Estimate Historics (Language copied from the estimate this snapshot belongs to.).
- each Family (has defect types) links to any number of Defect Types; each Defect Type links to exactly one Family; each Family–Defect Type pairing has the uniqueness → Num (Defect types classified for this family.).
- each Family (has estimates) links to any number of Estimation::Estimates; each Estimation::Estimate links to exactly one Family (Size and time estimates recorded in this family.).
- each Family (has languages) links to any number of Languages; each Language links to exactly one Family; each Family–Language pairing has the uniqueness → Name (Programming languages used when estimating in this family.).
- each Family (has phases) links to any number of Phases; each Phase links to exactly one Family; each Family–Phase pairing has the uniqueness → Num (Ordered phase skeleton for this family.).
- each Family (has processes) links to any number of Definition::Processes; each Definition::Process links to exactly one Family; each Family–Definition::Process pairing has the uniqueness → Name, Version, and Version Minor (Versioned processes that belong to this family.).
- each Project::Core::Project (checks phase products) links to any number of Phases; each Phase may link to any number of Project::Core::Projects; each Project::Core::Project–Phase pairing is a Project::Core::Phase Products Check (Whether each phase's products are satisfied for this project.).
- each Project::Core::Project (current phase) may link to at most one Phase; each Phase may link to any number of Project::Core::Projects (Phase whose forms are currently open, when one is set.).
- each Project::Core::Project (has probe estimate for) links to any number of Phases; each Phase links to exactly one Project::Core::Project; each Project::Core::Project–Phase pairing is a Project::Estimation::Estimate Probe (PROBE size and time calculation for a phase of this project.).
- each Project::Core::Project (uses language) links to exactly one Language; each Language may link to any number of Project::Core::Projects (Language this project is implemented in.).
- each Project::Core::Project Part (current phase) may link to at most one Phase; each Phase may link to any number of Project::Core::Project Parts (Phase whose forms are currently open, when one is set.).
- each Project::Core::Project Part (uses language) links to exactly one Language; each Language may link to any number of Project::Core::Project Parts (Language this part is implemented in.).
- each Project::Core::Project Stat Phase (for phase) links to exactly one Phase; each Phase may link to any number of Project::Core::Project Stat Phases (Phase these statistics are for.).
- each Project::Core::Task (occurs in) links to exactly one Phase; each Phase may link to any number of Project::Core::Tasks (Phase this task is performed in.).
- each Project::Core::Time Log (occurs in) links to exactly one Phase; each Phase may link to any number of Project::Core::Time Logs (Phase this time was spent in.).
- each Project::Estimation::Actual Loc (for phase) links to exactly one Phase; each Phase may link to any number of Project::Estimation::Actual Locs (Phase this size account is for.).
- each Project::Estimation::Estimate Loc (for phase) links to exactly one Phase; each Phase may link to any number of Project::Estimation::Estimate Locs (Phase this size account is for.).
- each Project::Estimation::Estimate Probe Add Loc (for phase) links to exactly one Phase; each Phase may link to any number of Project::Estimation::Estimate Probe Add Locs (Phase this added-object line is for.).
- each Project::Estimation::Estimate Probe Object Loc (for phase) links to exactly one Phase; each Phase may link to any number of Project::Estimation::Estimate Probe Object Locs (Phase this new-object line is for.).
- each Project::Estimation::Estimate Probe Object Reused (for phase) links to exactly one Phase; each Phase may link to any number of Project::Estimation::Estimate Probe Object Reuseds (Phase this reused-object line is for.).
- each Project::Quality::Defect (injected in phase) links to exactly one Phase; each Phase may link to any number of Project::Quality::Defects (Phase where this defect was injected.).
- each Project::Quality::Defect (removed in phase) links to exactly one Phase; each Phase may link to any number of Project::Quality::Defects (Phase where this defect was removed.).
- each Project::Quality::Issue (injected in phase) links to exactly one Phase; each Phase may link to any number of Project::Quality::Issues (Phase where this issue was found.).
- each Project::Quality::Process Improvement Proposal (on phase) links to exactly one Phase; each Phase may link to any number of Project::Quality::Process Improvement Proposals (Phase this proposal is about.).

## Indexes

- No Families can share the same Name.

