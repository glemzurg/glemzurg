[⇦ Development Process](model.md) / [Process](domain-domain.process.md) / [Definition](subdomain-domain.process.subdomain.definition.md)

# Model Facts — Definition

Association multiplicity, association invariant, and index uniqueness constraints for this subdomain.

## Associations

- each Module Template (follows process) may link to at most one Process::Process; each Process::Process may link to any number of Module Templates (Process this template is based on, when one is set.).
- each Module Template (uses design method) links to exactly one Method::Design Method; each Method::Design Method may link to any number of Module Templates (Design template used under this module template.).
- each Module Template (uses language) links to exactly one Family::Language; each Family::Language may link to any number of Module Templates (Language this module template is for.).
- each Module Template (uses size estimation method) links to exactly one Estimate::Method; each Estimate::Method may link to any number of Module Templates (Method used to estimate size under this template.).
- each Module Template (uses time estimation method) links to exactly one Estimate::Method; each Estimate::Method may link to any number of Module Templates (Method used to estimate time under this template.).
- each Project::Core::Project (instantiates) links to exactly one Module Template; each Module Template may link to any number of Project::Core::Projects (Module template this project is created from.).
- each Project::Core::Project Part (instantiates) links to exactly one Module Template; each Module Template may link to any number of Project::Core::Project Parts (Module template this part is created from.).
- each Project::Cycle::Project Cycle Actual (instantiates) links to exactly one Module Template; each Module Template may link to any number of Project::Cycle::Project Cycle Actuals (Module template this cycle actual is based on.).
- each Project::Cycle::Project Cycle Plan (instantiates) links to exactly one Module Template; each Module Template may link to any number of Project::Cycle::Project Cycle Plans (Module template this cycle plan is based on.).
- each Project::Estimation::Estimate Probe Add Loc (of size) links to exactly one Probe Object Size; each Probe Object Size may link to any number of Project::Estimation::Estimate Probe Add Locs (Relative size of this added object. SQL column relative_size.).
- each Project::Estimation::Estimate Probe Add Loc (of type) links to exactly one Probe Type; each Probe Type may link to any number of Project::Estimation::Estimate Probe Add Locs (PROBE type of this added object. SQL column type.).
- each Project::Estimation::Estimate Probe Object Loc (of size) links to exactly one Probe Object Size; each Probe Object Size may link to any number of Project::Estimation::Estimate Probe Object Locs (Relative size of this object. SQL column relative_size.).
- each Project::Estimation::Estimate Probe Object Loc (of type) links to exactly one Probe Type; each Probe Type may link to any number of Project::Estimation::Estimate Probe Object Locs (PROBE type of this object. SQL column type.).

