[⇦ Development Process](model.md) / [Project](domain-domain.project.md) / [Estimation](subdomain-domain.project.subdomain.estimation.md)

# Model Facts — Estimation

Association multiplicity, association invariant, and index uniqueness constraints for this subdomain.

## Associations

- each Actual Loc (for phase) links to exactly one Process::Definition::Phase; each Process::Definition::Phase may link to any number of Actual Locs (Phase this size account is for.).
- each Core::Project (has actual loc) may link to at most one Actual Loc; each Actual Loc links to exactly one Core::Project (Actual lines-of-code account for this project.).
- each Core::Project (has loc estimate) may link to at most one Estimate Loc; each Estimate Loc links to exactly one Core::Project (Planned lines-of-code account for this project.).
- each Core::Project (has probe add loc) links to any number of Estimate Probe Add Locs; each Estimate Probe Add Loc links to exactly one Core::Project (Added-object LOC lines in the PROBE estimate.).
- each Core::Project (has probe estimate for) links to any number of Process::Definition::Phases; each Process::Definition::Phase links to exactly one Core::Project; each Core::Project–Process::Definition::Phase pairing is a Estimate Probe (PROBE size and time calculation for a phase of this project.).
- each Core::Project (has probe object loc) links to any number of Estimate Probe Object Locs; each Estimate Probe Object Loc links to exactly one Core::Project (New-object LOC lines in the PROBE estimate.).
- each Core::Project (has probe object reused) links to any number of Estimate Probe Object Reuseds; each Estimate Probe Object Reused links to exactly one Core::Project (Reused-object LOC lines in the PROBE estimate.).
- each Estimate Loc (for phase) links to exactly one Process::Definition::Phase; each Process::Definition::Phase may link to any number of Estimate Locs (Phase this size account is for.).
- each Estimate Probe Add Loc (for phase) links to exactly one Process::Definition::Phase; each Process::Definition::Phase may link to any number of Estimate Probe Add Locs (Phase this added-object line is for.).
- each Estimate Probe Add Loc (of size) links to exactly one Process::Definition::Probe Object Size; each Process::Definition::Probe Object Size may link to any number of Estimate Probe Add Locs (Relative size of this added object. SQL column relative_size.).
- each Estimate Probe Add Loc (of type) links to exactly one Process::Definition::Probe Type; each Process::Definition::Probe Type may link to any number of Estimate Probe Add Locs (PROBE type of this added object. SQL column type.).
- each Estimate Probe Object Loc (for phase) links to exactly one Process::Definition::Phase; each Process::Definition::Phase may link to any number of Estimate Probe Object Locs (Phase this new-object line is for.).
- each Estimate Probe Object Loc (of size) links to exactly one Process::Definition::Probe Object Size; each Process::Definition::Probe Object Size may link to any number of Estimate Probe Object Locs (Relative size of this object. SQL column relative_size.).
- each Estimate Probe Object Loc (of type) links to exactly one Process::Definition::Probe Type; each Process::Definition::Probe Type may link to any number of Estimate Probe Object Locs (PROBE type of this object. SQL column type.).
- each Estimate Probe Object Reused (for phase) links to exactly one Process::Definition::Phase; each Process::Definition::Phase may link to any number of Estimate Probe Object Reuseds (Phase this reused-object line is for.).

