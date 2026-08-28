[⇦ Development Process](model.md) / [Process](domain-domain.process.md) / [Project](subdomain-domain.process.subdomain.project.md)

# Model Facts — Project

Association multiplicity, association invariant, and index uniqueness constraints for this subdomain.

## Associations

- each Project (checks phase products) links to any number of Definition::Phases; each Definition::Phase may link to any number of Projects; each Project–Definition::Phase pairing is a Phase Products Check (Whether each phase's products are satisfied for this project.).
- each Project (current phase) may link to at most one Definition::Phase; each Definition::Phase may link to any number of Projects (Phase whose forms are currently open, when one is set.).
- each Project (current subphase) may link to at most one Definition::Step; each Definition::Step may link to any number of Projects (Planning step currently taking place, when one is set.).
- each Project (follows process) links to exactly one Definition::Process; each Definition::Process may link to any number of Projects (Process this project follows.).
- each Project (has actual loc) may link to at most one Estimation::Actual Loc; each Estimation::Actual Loc links to exactly one Project (Actual lines-of-code account for this project.).
- each Project (has cycle actuals) links to any number of Project Cycle Actuals; each Project Cycle Actual links to exactly one Project; each Project–Project Cycle Actual pairing has the uniqueness → Num (Actual reuse values recorded per cycle.).
- each Project (has cycle plans) links to any number of Project Cycle Plans; each Project Cycle Plan links to exactly one Project; each Project–Project Cycle Plan pairing has the uniqueness → Num (Planned reuse values recorded per cycle.).
- each Project (has loc estimate) may link to at most one Estimation::Estimate Loc; each Estimation::Estimate Loc links to exactly one Project (Planned lines-of-code account for this project.).
- each Project (has parts) links to any number of Project Parts; each Project Part links to exactly one Project (Language-specific parts of this project.).
- each Project (has probe add loc) links to any number of Estimation::Estimate Probe Add Locs; each Estimation::Estimate Probe Add Loc links to exactly one Project (Added-object LOC lines in the PROBE estimate.).
- each Project (has probe estimate for) links to any number of Definition::Phases; each Definition::Phase links to exactly one Project; each Project–Definition::Phase pairing is a Estimation::Estimate Probe (PROBE size and time calculation for a phase of this project.).
- each Project (has probe object loc) links to any number of Estimation::Estimate Probe Object Locs; each Estimation::Estimate Probe Object Loc links to exactly one Project (New-object LOC lines in the PROBE estimate.).
- each Project (has probe object reused) links to any number of Estimation::Estimate Probe Object Reuseds; each Estimation::Estimate Probe Object Reused links to exactly one Project (Reused-object LOC lines in the PROBE estimate.).
- each Project (has schedule) may link to at most one Schedule; each Schedule links to exactly one Project (Schedule for recording planned work.).
- each Project (has stat phases) links to any number of Project Stat Phases; each Project Stat Phase links to exactly one Project (Per-phase statistical estimates for this project.).
- each Project (has tasks) links to any number of Tasks; each Task links to exactly one Project (Planned tasks on this project.).
- each Project (has time logs) links to any number of Time Logs; each Time Log links to exactly one Project (Recorded time log entries.).
- each Project (in bucket) may link to at most one Definition::Stats Bucket; each Definition::Stats Bucket may link to any number of Projects (Stats bucket this project is grouped in, when one is set.).
- each Project (instantiates) links to exactly one Definition::Module Template; each Definition::Module Template may link to any number of Projects (Module template this project is created from.).
- each Project (uses design method) links to exactly one Definition::Design Method; each Definition::Design Method may link to any number of Projects (Design template used by this project.).
- each Project (uses language) links to exactly one Definition::Language; each Definition::Language may link to any number of Projects (Language this project is implemented in.).
- each Project (uses size estimation method) links to exactly one Definition::Method; each Definition::Method may link to any number of Projects (Method used to estimate size.).
- each Project (uses time estimation method) links to exactly one Definition::Method; each Definition::Method may link to any number of Projects (Method used to estimate time.).
- each Project Cycle Actual (instantiates) links to exactly one Definition::Module Template; each Definition::Module Template may link to any number of Project Cycle Actuals (Module template this cycle actual is based on.).
- each Project Cycle Plan (instantiates) links to exactly one Definition::Module Template; each Definition::Module Template may link to any number of Project Cycle Plans (Module template this cycle plan is based on.).
- each Project Part (current phase) may link to at most one Definition::Phase; each Definition::Phase may link to any number of Project Parts (Phase whose forms are currently open, when one is set.).
- each Project Part (current subphase) may link to at most one Definition::Step; each Definition::Step may link to any number of Project Parts (Planning step currently taking place, when one is set.).
- each Project Part (in bucket) may link to at most one Definition::Stats Bucket; each Definition::Stats Bucket may link to any number of Project Parts (Stats bucket this part is grouped in, when one is set.).
- each Project Part (instantiates) links to exactly one Definition::Module Template; each Definition::Module Template may link to any number of Project Parts (Module template this part is created from.).
- each Project Part (uses design method) links to exactly one Definition::Design Method; each Definition::Design Method may link to any number of Project Parts.
- each Project Part (uses language) links to exactly one Definition::Language; each Definition::Language may link to any number of Project Parts (Language this part is implemented in.).
- each Project Part (uses size estimation method) links to exactly one Definition::Method; each Definition::Method may link to any number of Project Parts.
- each Project Part (uses time estimation method) links to exactly one Definition::Method; each Definition::Method may link to any number of Project Parts.
- each Project Stat Phase (for phase) links to exactly one Definition::Phase; each Definition::Phase may link to any number of Project Stat Phases (Phase these statistics are for.).
- each Project Stat Phase (in bucket) links to exactly one Definition::Stats Bucket; each Definition::Stats Bucket may link to any number of Project Stat Phases (Stats bucket these statistics are grouped in.).
- each Project Stat Phase (uses method) may link to at most one Definition::Method; each Definition::Method may link to any number of Project Stat Phases (Programming method for these statistics, when one is set.).
- each Quality::Defect (injected in) links to exactly one Project; each Project may link to any number of Quality::Defects (Project where this defect was injected.).
- each Quality::Defect (removed in) links to exactly one Project; each Project may link to any number of Quality::Defects (Project where this defect was removed.).
- each Quality::Issue (injected in) links to exactly one Project; each Project may link to any number of Quality::Issues (Project where this issue was found.).
- each Quality::Process Improvement Proposal (on project) links to exactly one Project; each Project may link to any number of Quality::Process Improvement Proposals (Project this proposal was raised on.).
- each Quality::Test Case (for project) links to exactly one Project; each Project may link to any number of Quality::Test Cases (Project this test case belongs to.).
- each Quality::Test Case Result (for project) links to exactly one Project; each Project may link to any number of Quality::Test Case Results (Project this result was recorded against.).
- each Schedule (has weeks) links to any number of Schedule Weeks; each Schedule Week links to exactly one Schedule; each Schedule–Schedule Week pairing has the uniqueness → Num (Ordered weeks (or day slots) on this schedule.).
- each Schedule Week (for project) links to exactly one Project; each Project may link to any number of Schedule Weeks (Project this week belongs to, copied from the schedule.).
- each Task (occurs in) links to exactly one Definition::Phase; each Definition::Phase may link to any number of Tasks (Phase this task is performed in.).
- each Task (on week) links to exactly one Schedule Week; each Schedule Week may link to any number of Tasks (Schedule week this task is planned for.).
- each Time Log (for task) links to exactly one Task; each Task may link to any number of Time Logs (Task this time was spent on.).
- each Time Log (occurs in) links to exactly one Definition::Phase; each Definition::Phase may link to any number of Time Logs (Phase this time was spent in.).

