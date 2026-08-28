# Development Process

A process-definition and estimation catalog used to test the requirements tool.

Surrogate primary keys are object identity. Foreign keys are associations. Remaining columns are attributes.

## Live schema mapping

Tables and columns from `development-process-schema.sql` (uncommented `CREATE TABLE` only).

| Table | Column | Model |
| --- | --- | --- |
| family | family_id | Family identity |
| family | name | Family.name (unique index) |
| family | description | Family.description |
| phase | phase_id | Phase identity |
| phase | family_id | Family Has Phases |
| phase | num | Phase.num (unique per family) |
| phase | name | Phase.name (unique per family) |
| phase | description | Phase.description |
| defect_type | defect_type_id | Defect Type identity |
| defect_type | family_id | Family Has Defect Types |
| defect_type | num | Defect Type.num (unique per family) |
| defect_type | name | Defect Type.name (unique per family) |
| defect_type | description | Defect Type.description |
| defect_type | base_num | Defect Type.base_num |
| language | language_id | Language identity |
| language | family_id | Family Has Languages |
| language | name | Language.name (unique per family) |
| language | description | Language.description |
| language | INDEX (language_id, family_id) | Supports Language composite FK; not a second identity |
| process | process_id | Process identity |
| process | family_id | Family Has Processes |
| process | name, version, version_minor | unique per family |
| process | purpose | Process.purpose |
| process | entry_criteria | Process.entry_criteria |
| process | exit_criteria | Process.exit_criteria |
| process | script_lock | Process.script_lock |
| process | ancestor_id | Process Has Ancestor |
| script | script_id | Script identity |
| script | process_id | Process Has Scripts |
| script | num | Script.num (unique per process) |
| script | name | Script.name (unique per process) |
| script | task_summary | Script.task_summary |
| script | purpose | Script.purpose |
| script | entry_criteria | Script.entry_criteria |
| script | exit_criteria | Script.exit_criteria |
| script | cycle | Script.cycle |
| step | step_id | Step identity |
| step | script_id | Script Has Steps |
| step | num | Step.num (unique per script) |
| step | name | Step.name (unique per script) |
| step | phase_id | Step Occurs In Phase |
| step | tasks | Step.tasks |
| estimate | estimate_id | Estimate identity |
| estimate | family_id | Family Has Estimates |
| estimate | language_id | Estimate Uses Language |
| estimate | axis | Estimate.axis |
| estimate | scope | Estimate.scope |
| estimate | version | Estimate.version |
| estimate | mean | Estimate.mean |
| estimate | variance | Estimate.variance |
| estimate | low | Estimate.low |
| estimate | high | Estimate.high |
| estimate | actual | Estimate.actual |
| estimate | method | Estimate.method |
| estimate | prediction_interval | Estimate.prediction_interval |
| estimate | comment | Estimate.comment |
| estimate | estimation_time | Estimate.estimation_time |
| estimate | guess_lowest_80pred | Estimate.guess_lowest_80pred |
| estimate | guess_highest_80pred | Estimate.guess_highest_80pred |
| estimate | sum_count | Estimate.sum_count |
| estimate | sum_mean | Estimate.sum_mean |
| estimate | sum_variance | Estimate.sum_variance |
| estimate | portion_portion | Estimate.portion_portion |
| estimate | portion_mean | Estimate.portion_mean |
| estimate | portion_variance | Estimate.portion_variance |
| estimate | INDEX (actual) | Lookup aid, not an identity |
| estimate_historic | estimate_id | Estimate Has History |
| estimate_historic | family_id | Estimate Historic Belongs To Family |
| estimate_historic | language_id | Estimate Historic Uses Language |
| estimate_historic | version | Estimate Historic.version (unique per estimate) |
| estimate_historic | remaining columns | Same attributes as Estimate |
## Actors

The actors of this model.





## Domains

The domains of this model.

```mermaid
graph TD
domain_domain_process["Process"]

```

- **[Process](domain-domain.process.md).** Process families, scripts, and the estimates recorded against them.


## Invariants



## Global Functions

*No global functions*

## Named Sets

*No named sets*
