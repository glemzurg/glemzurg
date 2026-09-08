[⇦ Development Process](model.md) / [Process](domain-domain.process.md) / [Definition](subdomain-domain.process.subdomain.definition.md)

# Model Facts — Definition

Association multiplicity, association invariant, and index uniqueness constraints for this subdomain.

## Associations

- each Estimation::Estimate (uses language) links to exactly one Language; each Language may link to any number of Estimation::Estimates (Language this estimate is categorized by. The language must belong to the same family.).
- each Estimation::Estimate Historic (belongs to family) links to exactly one Family; each Family may link to any number of Estimation::Estimate Historics (Family copied from the estimate this snapshot belongs to.).
- each Estimation::Estimate Historic (uses language) links to exactly one Language; each Language may link to any number of Estimation::Estimate Historics (Language copied from the estimate this snapshot belongs to.).
- each Family (has defect types) links to any number of Defect Types; each Defect Type links to exactly one Family; each Family–Defect Type pairing has the uniqueness → Num (Defect types classified for this family.).
- each Family (has estimates) links to any number of Estimation::Estimates; each Estimation::Estimate links to exactly one Family (Size and time estimates recorded in this family.).
- each Family (has languages) links to any number of Languages; each Language links to exactly one Family; each Family–Language pairing has the uniqueness → Name (Programming languages used when estimating in this family.).
- each Family (has phases) links to any number of Phases; each Phase links to exactly one Family; each Family–Phase pairing has the uniqueness → Num (Ordered phase skeleton for this family.).
- each Family (has processes) links to any number of Processes; each Process links to exactly one Family; each Family–Process pairing has the uniqueness → Name, Version, and Version Minor (Versioned processes that belong to this family.).
- each Module Template (follows process) may link to at most one Process; each Process may link to any number of Module Templates (Process this template is based on, when one is set.).
- each Module Template (uses design method) links to exactly one Design Method; each Design Method may link to any number of Module Templates (Design template used under this module template.).
- each Module Template (uses language) links to exactly one Language; each Language may link to any number of Module Templates (Language this module template is for.).
- each Module Template (uses size estimation method) links to exactly one Method; each Method may link to any number of Module Templates (Method used to estimate size under this template.).
- each Module Template (uses time estimation method) links to exactly one Method; each Method may link to any number of Module Templates (Method used to estimate time under this template.).
- each Process (has ancestor) may link to at most one Process; each Process may link to any number of Processes (Earlier process this version replaces, when one exists.).
- each Process (has scripts) links to any number of Scripts; each Script links to exactly one Process; each Process–Script pairing has the uniqueness → Num (Ordered scripts that make up this process.).
- each Project::Core::Project (checks phase products) links to any number of Phases; each Phase may link to any number of Project::Core::Projects; each Project::Core::Project–Phase pairing is a Project::Core::Phase Products Check (Whether each phase's products are satisfied for this project.).
- each Project::Core::Project (current phase) may link to at most one Phase; each Phase may link to any number of Project::Core::Projects (Phase whose forms are currently open, when one is set.).
- each Project::Core::Project (current subphase) may link to at most one Step; each Step may link to any number of Project::Core::Projects (Planning step currently taking place, when one is set.).
- each Project::Core::Project (follows process) links to exactly one Process; each Process may link to any number of Project::Core::Projects (Process this project follows.).
- each Project::Core::Project (has probe estimate for) links to any number of Phases; each Phase links to exactly one Project::Core::Project; each Project::Core::Project–Phase pairing is a Project::Estimation::Estimate Probe (PROBE size and time calculation for a phase of this project.).
- each Project::Core::Project (in bucket) may link to at most one Stats Bucket; each Stats Bucket may link to any number of Project::Core::Projects (Stats bucket this project is grouped in, when one is set.).
- each Project::Core::Project (instantiates) links to exactly one Module Template; each Module Template may link to any number of Project::Core::Projects (Module template this project is created from.).
- each Project::Core::Project (uses design method) links to exactly one Design Method; each Design Method may link to any number of Project::Core::Projects (Design template used by this project.).
- each Project::Core::Project (uses language) links to exactly one Language; each Language may link to any number of Project::Core::Projects (Language this project is implemented in.).
- each Project::Core::Project (uses size estimation method) links to exactly one Method; each Method may link to any number of Project::Core::Projects (Method used to estimate size.).
- each Project::Core::Project (uses time estimation method) links to exactly one Method; each Method may link to any number of Project::Core::Projects (Method used to estimate time.).
- each Project::Core::Project Cycle Actual (instantiates) links to exactly one Module Template; each Module Template may link to any number of Project::Core::Project Cycle Actuals (Module template this cycle actual is based on.).
- each Project::Core::Project Cycle Plan (instantiates) links to exactly one Module Template; each Module Template may link to any number of Project::Core::Project Cycle Plans (Module template this cycle plan is based on.).
- each Project::Core::Project Part (current phase) may link to at most one Phase; each Phase may link to any number of Project::Core::Project Parts (Phase whose forms are currently open, when one is set.).
- each Project::Core::Project Part (current subphase) may link to at most one Step; each Step may link to any number of Project::Core::Project Parts (Planning step currently taking place, when one is set.).
- each Project::Core::Project Part (in bucket) may link to at most one Stats Bucket; each Stats Bucket may link to any number of Project::Core::Project Parts (Stats bucket this part is grouped in, when one is set.).
- each Project::Core::Project Part (instantiates) links to exactly one Module Template; each Module Template may link to any number of Project::Core::Project Parts (Module template this part is created from.).
- each Project::Core::Project Part (uses design method) links to exactly one Design Method; each Design Method may link to any number of Project::Core::Project Parts.
- each Project::Core::Project Part (uses language) links to exactly one Language; each Language may link to any number of Project::Core::Project Parts (Language this part is implemented in.).
- each Project::Core::Project Part (uses size estimation method) links to exactly one Method; each Method may link to any number of Project::Core::Project Parts.
- each Project::Core::Project Part (uses time estimation method) links to exactly one Method; each Method may link to any number of Project::Core::Project Parts.
- each Project::Core::Project Stat Phase (for phase) links to exactly one Phase; each Phase may link to any number of Project::Core::Project Stat Phases (Phase these statistics are for.).
- each Project::Core::Project Stat Phase (in bucket) links to exactly one Stats Bucket; each Stats Bucket may link to any number of Project::Core::Project Stat Phases (Stats bucket these statistics are grouped in.).
- each Project::Core::Project Stat Phase (uses method) may link to at most one Method; each Method may link to any number of Project::Core::Project Stat Phases (Programming method for these statistics, when one is set.).
- each Project::Core::Task (occurs in) links to exactly one Phase; each Phase may link to any number of Project::Core::Tasks (Phase this task is performed in.).
- each Project::Core::Time Log (occurs in) links to exactly one Phase; each Phase may link to any number of Project::Core::Time Logs (Phase this time was spent in.).
- each Project::Estimation::Actual Loc (for phase) links to exactly one Phase; each Phase may link to any number of Project::Estimation::Actual Locs (Phase this size account is for.).
- each Project::Estimation::Estimate Loc (for phase) links to exactly one Phase; each Phase may link to any number of Project::Estimation::Estimate Locs (Phase this size account is for.).
- each Project::Estimation::Estimate Probe Add Loc (for phase) links to exactly one Phase; each Phase may link to any number of Project::Estimation::Estimate Probe Add Locs (Phase this added-object line is for.).
- each Project::Estimation::Estimate Probe Add Loc (of size) links to exactly one Probe Object Size; each Probe Object Size may link to any number of Project::Estimation::Estimate Probe Add Locs (Relative size of this added object. SQL column relative_size.).
- each Project::Estimation::Estimate Probe Add Loc (of type) links to exactly one Probe Type; each Probe Type may link to any number of Project::Estimation::Estimate Probe Add Locs (PROBE type of this added object. SQL column type.).
- each Project::Estimation::Estimate Probe Object Loc (for phase) links to exactly one Phase; each Phase may link to any number of Project::Estimation::Estimate Probe Object Locs (Phase this new-object line is for.).
- each Project::Estimation::Estimate Probe Object Loc (of size) links to exactly one Probe Object Size; each Probe Object Size may link to any number of Project::Estimation::Estimate Probe Object Locs (Relative size of this object. SQL column relative_size.).
- each Project::Estimation::Estimate Probe Object Loc (of type) links to exactly one Probe Type; each Probe Type may link to any number of Project::Estimation::Estimate Probe Object Locs (PROBE type of this object. SQL column type.).
- each Project::Estimation::Estimate Probe Object Reused (for phase) links to exactly one Phase; each Phase may link to any number of Project::Estimation::Estimate Probe Object Reuseds (Phase this reused-object line is for.).
- each Project::Quality::Defect (injected in phase) links to exactly one Phase; each Phase may link to any number of Project::Quality::Defects (Phase where this defect was injected.).
- each Project::Quality::Defect (removed in phase) links to exactly one Phase; each Phase may link to any number of Project::Quality::Defects (Phase where this defect was removed.).
- each Project::Quality::Issue (injected in phase) links to exactly one Phase; each Phase may link to any number of Project::Quality::Issues (Phase where this issue was found.).
- each Project::Quality::Process Improvement Proposal (on phase) links to exactly one Phase; each Phase may link to any number of Project::Quality::Process Improvement Proposals (Phase this proposal is about.).
- each Project::Quality::Process Improvement Proposal (on process) links to exactly one Process; each Process may link to any number of Project::Quality::Process Improvement Proposals (Process this proposal is about.).
- each Project::Quality::Process Improvement Proposal (on subphase) links to exactly one Step; each Step may link to any number of Project::Quality::Process Improvement Proposals (Planning step this proposal is about.).
- each Project::Quality::Process Improvement Proposal (resolved in process) links to exactly one Process; each Process may link to any number of Project::Quality::Process Improvement Proposals (Process version that absorbed this proposal.).
- each Script (has steps) links to any number of Steps; each Step links to exactly one Script; each Script–Step pairing has the uniqueness → Num (Ordered steps of this script.).
- each Step (occurs in) links to exactly one Phase; each Phase may link to any number of Steps (Phase of the family skeleton this step is performed in.).

## Indexes

- No Families can share the same Name.
- No Methods can share the same Name.

