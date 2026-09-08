[⇦ Development Process](model.md) / [Process](domain-domain.process.md) / [Definition](subdomain-domain.process.subdomain.definition.md)

# Model Facts — Definition

Association multiplicity, association invariant, and index uniqueness constraints for this subdomain.

## Associations

- each Family::Family (has processes) links to any number of Processes; each Process links to exactly one Family::Family; each Family::Family–Process pairing has the uniqueness → Name, Version, and Version Minor (Versioned processes that belong to this family.).
- each Module Template (follows process) may link to at most one Process; each Process may link to any number of Module Templates (Process this template is based on, when one is set.).
- each Module Template (uses design method) links to exactly one Design Method; each Design Method may link to any number of Module Templates (Design template used under this module template.).
- each Module Template (uses language) links to exactly one Family::Language; each Family::Language may link to any number of Module Templates (Language this module template is for.).
- each Module Template (uses size estimation method) links to exactly one Method; each Method may link to any number of Module Templates (Method used to estimate size under this template.).
- each Module Template (uses time estimation method) links to exactly one Method; each Method may link to any number of Module Templates (Method used to estimate time under this template.).
- each Process (has ancestor) may link to at most one Process; each Process may link to any number of Processes (Earlier process this version replaces, when one exists.).
- each Process (has scripts) links to any number of Scripts; each Script links to exactly one Process; each Process–Script pairing has the uniqueness → Num (Ordered scripts that make up this process.).
- each Project::Core::Project (current subphase) may link to at most one Step; each Step may link to any number of Project::Core::Projects (Planning step currently taking place, when one is set.).
- each Project::Core::Project (follows process) links to exactly one Process; each Process may link to any number of Project::Core::Projects (Process this project follows.).
- each Project::Core::Project (in bucket) may link to at most one Stats Bucket; each Stats Bucket may link to any number of Project::Core::Projects (Stats bucket this project is grouped in, when one is set.).
- each Project::Core::Project (instantiates) links to exactly one Module Template; each Module Template may link to any number of Project::Core::Projects (Module template this project is created from.).
- each Project::Core::Project (uses design method) links to exactly one Design Method; each Design Method may link to any number of Project::Core::Projects (Design template used by this project.).
- each Project::Core::Project (uses size estimation method) links to exactly one Method; each Method may link to any number of Project::Core::Projects (Method used to estimate size.).
- each Project::Core::Project (uses time estimation method) links to exactly one Method; each Method may link to any number of Project::Core::Projects (Method used to estimate time.).
- each Project::Core::Project Cycle Actual (instantiates) links to exactly one Module Template; each Module Template may link to any number of Project::Core::Project Cycle Actuals (Module template this cycle actual is based on.).
- each Project::Core::Project Cycle Plan (instantiates) links to exactly one Module Template; each Module Template may link to any number of Project::Core::Project Cycle Plans (Module template this cycle plan is based on.).
- each Project::Core::Project Part (current subphase) may link to at most one Step; each Step may link to any number of Project::Core::Project Parts (Planning step currently taking place, when one is set.).
- each Project::Core::Project Part (in bucket) may link to at most one Stats Bucket; each Stats Bucket may link to any number of Project::Core::Project Parts (Stats bucket this part is grouped in, when one is set.).
- each Project::Core::Project Part (instantiates) links to exactly one Module Template; each Module Template may link to any number of Project::Core::Project Parts (Module template this part is created from.).
- each Project::Core::Project Part (uses design method) links to exactly one Design Method; each Design Method may link to any number of Project::Core::Project Parts.
- each Project::Core::Project Part (uses size estimation method) links to exactly one Method; each Method may link to any number of Project::Core::Project Parts.
- each Project::Core::Project Part (uses time estimation method) links to exactly one Method; each Method may link to any number of Project::Core::Project Parts.
- each Project::Core::Project Stat Phase (in bucket) links to exactly one Stats Bucket; each Stats Bucket may link to any number of Project::Core::Project Stat Phases (Stats bucket these statistics are grouped in.).
- each Project::Core::Project Stat Phase (uses method) may link to at most one Method; each Method may link to any number of Project::Core::Project Stat Phases (Programming method for these statistics, when one is set.).
- each Project::Estimation::Estimate Probe Add Loc (of size) links to exactly one Probe Object Size; each Probe Object Size may link to any number of Project::Estimation::Estimate Probe Add Locs (Relative size of this added object. SQL column relative_size.).
- each Project::Estimation::Estimate Probe Add Loc (of type) links to exactly one Probe Type; each Probe Type may link to any number of Project::Estimation::Estimate Probe Add Locs (PROBE type of this added object. SQL column type.).
- each Project::Estimation::Estimate Probe Object Loc (of size) links to exactly one Probe Object Size; each Probe Object Size may link to any number of Project::Estimation::Estimate Probe Object Locs (Relative size of this object. SQL column relative_size.).
- each Project::Estimation::Estimate Probe Object Loc (of type) links to exactly one Probe Type; each Probe Type may link to any number of Project::Estimation::Estimate Probe Object Locs (PROBE type of this object. SQL column type.).
- each Project::Quality::Process Improvement Proposal (on process) links to exactly one Process; each Process may link to any number of Project::Quality::Process Improvement Proposals (Process this proposal is about.).
- each Project::Quality::Process Improvement Proposal (on subphase) links to exactly one Step; each Step may link to any number of Project::Quality::Process Improvement Proposals (Planning step this proposal is about.).
- each Project::Quality::Process Improvement Proposal (resolved in process) links to exactly one Process; each Process may link to any number of Project::Quality::Process Improvement Proposals (Process version that absorbed this proposal.).
- each Script (has steps) links to any number of Steps; each Step links to exactly one Script; each Script–Step pairing has the uniqueness → Num (Ordered steps of this script.).
- each Step (occurs in) links to exactly one Family::Phase; each Family::Phase may link to any number of Steps (Phase of the family skeleton this step is performed in.).

## Indexes

- No Methods can share the same Name.

