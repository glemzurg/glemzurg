[⇦ Development Process](model.md) / [Project](domain-domain.project.md) / [Task](subdomain-domain.project.subdomain.task.md)

# Model Facts — Task

Association multiplicity, association invariant, and index uniqueness constraints for this subdomain.

## Associations

- each Core::Project (has tasks) links to any number of Tasks; each Task links to exactly one Core::Project (Planned tasks on this project.).
- each Core::Project (has time logs) links to any number of Time Logs; each Time Log links to exactly one Core::Project (Recorded time log entries.).
- each Task (occurs in) links to exactly one Process::Family::Phase; each Process::Family::Phase may link to any number of Tasks (Phase this task is performed in.).
- each Task (on week) links to exactly one Schedule::Schedule Week; each Schedule::Schedule Week may link to any number of Tasks (Schedule week this task is planned for.).
- each Time Log (for task) links to exactly one Task; each Task may link to any number of Time Logs (Task this time was spent on.).
- each Time Log (occurs in) links to exactly one Process::Family::Phase; each Process::Family::Phase may link to any number of Time Logs (Phase this time was spent in.).

