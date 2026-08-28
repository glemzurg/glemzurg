[⇦ Development Process](model.md) / [Process](domain-domain.process.md) / [Quality](subdomain-domain.process.subdomain.quality.md)

# Model Facts — Quality

Association multiplicity, association invariant, and index uniqueness constraints for this subdomain.

## Associations

- each Defect (has source) may link to at most one Defect; each Defect may link to any number of Defects (Defect this one was cloned from, when one exists.).
- each Defect (injected in phase) links to exactly one Definition::Phase; each Definition::Phase may link to any number of Defects (Phase where this defect was injected.).
- each Defect (injected in) links to exactly one Project::Project; each Project::Project may link to any number of Defects (Project where this defect was injected.).
- each Defect (removed in phase) links to exactly one Definition::Phase; each Definition::Phase may link to any number of Defects (Phase where this defect was removed.).
- each Defect (removed in) links to exactly one Project::Project; each Project::Project may link to any number of Defects (Project where this defect was removed.).
- each Issue (injected in phase) links to exactly one Definition::Phase; each Definition::Phase may link to any number of Issues (Phase where this issue was found.).
- each Issue (injected in) links to exactly one Project::Project; each Project::Project may link to any number of Issues (Project where this issue was found.).
- each Process Improvement Proposal (on phase) links to exactly one Definition::Phase; each Definition::Phase may link to any number of Process Improvement Proposals (Phase this proposal is about.).
- each Process Improvement Proposal (on process) links to exactly one Definition::Process; each Definition::Process may link to any number of Process Improvement Proposals (Process this proposal is about.).
- each Process Improvement Proposal (on project) links to exactly one Project::Project; each Project::Project may link to any number of Process Improvement Proposals (Project this proposal was raised on.).
- each Process Improvement Proposal (on subphase) links to exactly one Definition::Step; each Definition::Step may link to any number of Process Improvement Proposals (Planning step this proposal is about.).
- each Process Improvement Proposal (resolved in process) links to exactly one Definition::Process; each Definition::Process may link to any number of Process Improvement Proposals (Process version that absorbed this proposal.).
- each Test Case (for project) links to exactly one Project::Project; each Project::Project may link to any number of Test Cases (Project this test case belongs to.).
- each Test Case (has results) links to any number of Test Case Results; each Test Case Result links to exactly one Test Case (Recorded runs of this test case.).
- each Test Case Result (for project) links to exactly one Project::Project; each Project::Project may link to any number of Test Case Results (Project this result was recorded against.).

