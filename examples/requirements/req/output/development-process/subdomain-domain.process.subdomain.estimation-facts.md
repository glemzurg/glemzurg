[⇦ Development Process](model.md) / [Process](domain-domain.process.md) / [Estimation](subdomain-domain.process.subdomain.estimation.md)

# Model Facts — Estimation

Association multiplicity, association invariant, and index uniqueness constraints for this subdomain.

## Associations

- each Actual Loc (for phase) links to exactly one Definition::Phase; each Definition::Phase may link to any number of Actual Locs (Phase this size account is for.).
- each Definition::Family (has estimates) links to any number of Estimates; each Estimate links to exactly one Definition::Family (Size and time estimates recorded in this family.).
- each Estimate (has history) links to any number of Estimate Historics; each Estimate Historic links to exactly one Estimate; each Estimate–Estimate Historic pairing has the uniqueness → Version (Prior versions of this estimate.).
- each Estimate (uses language) links to exactly one Definition::Language; each Definition::Language may link to any number of Estimates (Language this estimate is categorized by. The language must belong to the same family.).
- each Estimate Historic (belongs to family) links to exactly one Definition::Family; each Definition::Family may link to any number of Estimate Historics (Family copied from the estimate this snapshot belongs to.).
- each Estimate Historic (uses language) links to exactly one Definition::Language; each Definition::Language may link to any number of Estimate Historics (Language copied from the estimate this snapshot belongs to.).
- each Estimate Loc (for phase) links to exactly one Definition::Phase; each Definition::Phase may link to any number of Estimate Locs (Phase this size account is for.).
- each Estimate Probe Add Loc (for phase) links to exactly one Definition::Phase; each Definition::Phase may link to any number of Estimate Probe Add Locs (Phase this added-object line is for.).
- each Estimate Probe Add Loc (of size) links to exactly one Definition::Probe Object Size; each Definition::Probe Object Size may link to any number of Estimate Probe Add Locs (Relative size of this added object. SQL column relative_size.).
- each Estimate Probe Add Loc (of type) links to exactly one Definition::Probe Type; each Definition::Probe Type may link to any number of Estimate Probe Add Locs (PROBE type of this added object. SQL column type.).
- each Estimate Probe Object Loc (for phase) links to exactly one Definition::Phase; each Definition::Phase may link to any number of Estimate Probe Object Locs (Phase this new-object line is for.).
- each Estimate Probe Object Loc (of size) links to exactly one Definition::Probe Object Size; each Definition::Probe Object Size may link to any number of Estimate Probe Object Locs (Relative size of this object. SQL column relative_size.).
- each Estimate Probe Object Loc (of type) links to exactly one Definition::Probe Type; each Definition::Probe Type may link to any number of Estimate Probe Object Locs (PROBE type of this object. SQL column type.).
- each Estimate Probe Object Reused (for phase) links to exactly one Definition::Phase; each Definition::Phase may link to any number of Estimate Probe Object Reuseds (Phase this reused-object line is for.).
- each Project::Project (has actual loc) may link to at most one Actual Loc; each Actual Loc links to exactly one Project::Project (Actual lines-of-code account for this project.).
- each Project::Project (has loc estimate) may link to at most one Estimate Loc; each Estimate Loc links to exactly one Project::Project (Planned lines-of-code account for this project.).
- each Project::Project (has probe add loc) links to any number of Estimate Probe Add Locs; each Estimate Probe Add Loc links to exactly one Project::Project (Added-object LOC lines in the PROBE estimate.).
- each Project::Project (has probe estimate for) links to any number of Definition::Phases; each Definition::Phase links to exactly one Project::Project; each Project::Project–Definition::Phase pairing is a Estimate Probe (PROBE size and time calculation for a phase of this project.).
- each Project::Project (has probe object loc) links to any number of Estimate Probe Object Locs; each Estimate Probe Object Loc links to exactly one Project::Project (New-object LOC lines in the PROBE estimate.).
- each Project::Project (has probe object reused) links to any number of Estimate Probe Object Reuseds; each Estimate Probe Object Reused links to exactly one Project::Project (Reused-object LOC lines in the PROBE estimate.).

