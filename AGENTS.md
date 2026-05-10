# Repository Agent Instructions

## Branch discipline

Do not switch git branches, git add, git commit, git push, git pull.

## Documenation discipline

Comments and commit-style prose must explain the **why** of the current code, not its history.

- Frame rationale in the present tense, describing the system as it stands now.
- Do not reference removed code, prior versions, or past behavior. A future reader will not have that context.
- Do not reference plans, prompts, tickets, or design documents used to produce the change. A future reader will not have them either.

Example. Instead of "we removed `PRM_NO_KYC` because Topgolf doesn't use it", write "Topgolf players are presumed KYC-cleared at the partner side; the partner-service capability gate therefore distinguishes only `POINTS_ONLY` (jurisdiction / junior / guest) from `REAL_MONEY` (everything else)".

## Go `_test.go` files

- Use the [testify](https://github.com/stretchr/testify) framework (`require` for fatal assertions, `assert` for non-fatal).
- Prefer table-driven tests. Each row is a named case (`name string` field) and the test body uses `t.Run(tc.name, ...)`.
- Use table-driven tests even for two cases when more cases are likely to be added.

<!-- BEGIN BEADS INTEGRATION v:1 profile:minimal hash:7510c1e2 -->
## Beads Issue Tracker

This project uses **bd (beads)** for issue tracking. Run `bd prime` to see full workflow context and commands.

### Quick Reference

```bash
bd ready              # Find available work
bd show <id>          # View issue details
bd update <id> --claim  # Claim work
bd close <id>         # Complete work
```

### Rules

- Use `bd` for ALL task tracking — do NOT use TodoWrite, TaskCreate, or markdown TODO lists
- Run `bd prime` for detailed command reference and session close protocol
- Use `bd remember` for persistent knowledge — do NOT use MEMORY.md files

**Architecture in one line:** issues live in a local Dolt DB; sync uses `refs/dolt/data` on your git remote; `.beads/issues.jsonl` is a passive export. See https://github.com/gastownhall/beads/blob/main/docs/SYNC_CONCEPTS.md for details and anti-patterns.

## Session Completion

**When ending a work session**, you MUST complete ALL steps below. Work is NOT complete until all quality gates succee.

**MANDATORY WORKFLOW:**

1. **File issues for remaining work** - Create issues for anything that needs follow-up
2. **Run and pass all quality gates** (if code changed) - Tests, linters, builds
3. **Update issue status** - Close finished work, update in-progress items
5. **Clean up** - Clear stashes
7. **Hand off** - Provide context for next session

**CRITICAL RULES:**
- Work is NOT complete until all quality gates succeed
<!-- END BEADS INTEGRATION -->
