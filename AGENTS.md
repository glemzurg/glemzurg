# Repository Agent Instructions

## Any file type

Comments and commit-style prose must explain the **why** of the current code, not its history.

- Frame rationale in the present tense, describing the system as it stands now.
- Do not reference removed code, prior versions, or past behavior. A future reader will not have that context.
- Do not reference plans, prompts, tickets, or design documents used to produce the change. A future reader will not have them either.

Example. Instead of "we removed `PRM_NO_KYC` because Topgolf doesn't use it", write "Topgolf players are presumed KYC-cleared at the partner side; the partner-service capability gate therefore distinguishes only `POINTS_ONLY` (jurisdiction / junior / guest) from `REAL_MONEY` (everything else)".

## Go `_test.go` files

- Use the [testify](https://github.com/stretchr/testify) framework (`require` for fatal assertions, `assert` for non-fatal).
- Prefer table-driven tests. Each row is a named case (`name string` field) and the test body uses `t.Run(tc.name, ...)`.
- Use table-driven tests even for two cases when more cases are likely to be added.
