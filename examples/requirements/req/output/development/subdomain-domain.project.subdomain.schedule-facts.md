[⇦ Development](model.md) / [Project](domain-domain.project.md) / [Schedule](subdomain-domain.project.subdomain.schedule.md)

# Model Facts — Schedule

Association multiplicity, association invariant, and index uniqueness constraints for this subdomain.

## Associations

- each Core::Project (has schedule) may link to at most one Schedule; each Schedule links to exactly one Core::Project (Schedule for recording planned work.).
- each Schedule (has weeks) links to any number of Schedule Weeks; each Schedule Week links to exactly one Schedule; each Schedule–Schedule Week pairing has the uniqueness → Num (Ordered weeks (or day slots) on this schedule.).
- each Schedule Week (for project) links to exactly one Core::Project; each Core::Project may link to any number of Schedule Weeks (Project this week belongs to, copied from the schedule.).
- each Task::Task (on week) links to exactly one Schedule Week; each Schedule Week may link to any number of Task::Tasks (Schedule week this task is planned for.).

