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
- each Process (has ancestor) may link to at most one Process; each Process may link to any number of Processes (Earlier process this version replaces, when one exists.).
- each Process (has scripts) links to any number of Scripts; each Script links to exactly one Process; each Process–Script pairing has the uniqueness → Num (Ordered scripts that make up this process.).
- each Script (has steps) links to any number of Steps; each Step links to exactly one Script; each Script–Step pairing has the uniqueness → Num (Ordered steps of this script.).
- each Step (occurs in) links to exactly one Phase; each Phase may link to any number of Steps (Phase of the family skeleton this step is performed in.).

## Indexes

- No Families can share the same Name.

