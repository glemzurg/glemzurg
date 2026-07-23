# Repository Agent Instructions

## User edits take precedence

When the user changes code the assistant wrote earlier, treat the user's version as authoritative unless it is clearly broken (won't compile, fails an explicit requirement they just gave, etc.). Do not revert, "fix," or reintroduce assistant-written patterns the user removed — their edit is most often intentional and correct. Read the current file first, infer intent from the diff, and build on what is there now.

## Confirmation / Autonomy Policy

- Be decisive and proactive. After a plan is approved, execute all routine, low-risk, and standard actions without asking for confirmation.
- Routine actions include: file edits, running tests/builds, git operations, standard `bd` commands, shell commands that are not destructive, reading files, etc.
- **Executing beads** (issues the user has instructed me to create as a plan) is one of those routine actions — start work on a freshly-created bead without asking "should I begin?". The act of asking the assistant to draft an issue is itself the go-ahead to execute it. Only stop if the user explicitly says "just plan, don't execute" or similar.
- ONLY ask for human confirmation on HIGH-RISK actions: deleting files/folders, running potentially destructive commands, major architectural changes, or anything truly ambiguous.
- For everything else, just do the work and use Beads tasks to track progress. Do not interrupt the user with "do you want me to work?" style questions on innocuous tasks.

## Beads for everything

**Use `bd` (beads) for all work — always.** Every user request, follow-up, bugfix, refactor, doc change, and test addition gets tracked in beads. There is no "too small for a bead."

**Mandatory workflow:**

1. **Start of session** — run `bd prime` and `bd ready` before doing anything else.
2. **Before writing code** — create or claim a bead (`bd create` or `bd update <id> --claim`). If the user's request maps to existing open work, claim that issue instead of starting untracked.
3. **While working** — the active bead is the source of truth for what you are doing. Multi-step work gets an epic + child beads (see `.claude/skills/bead-epic/SKILL.md`).
4. **When done** — close the bead (`bd close <id>`). File new beads for anything left unfinished.
5. **After `req/` changes** — when any bead (task, bugfix, or epic child) changes code in or below `apps/requirements/req/`, add a **build-gate closing bead** that runs last (see [Quality gate](#quality-gate-appsrequirementsreqbuildsh)). Implementation beads close on their feature; the closing bead owns `build.sh` and any complexity fixes.
6. **Never substitute** — do not use TodoWrite, TaskCreate, markdown TODO lists, or informal mental tracking. Beads only.

Skipping beads is a workflow failure, not a time-saver.

## File permission discipline

All generated and written files should be owned by user vscode for easy manipulation in IDE.

## Branch discipline

Do not switch git branches, git add, git commit, git push, git pull.

## Core model refactor staging

`apps/requirements/req/internal/core` is the source of truth for model data structures. When a refactor changes shapes or invariants at the model core, work in this order and do not start the next stage until the current one compiles and its tests pass:

1. **`internal/core`** — types, constructors, setters, validation, and unit tests.
2. **`internal/test_helper`** — fixture models and prune helpers that mirror core shapes.
3. **`internal/database`** — schema, persistence, load/assemble paths, and database tests.
4. **`internal/parser_ai`** — JSON tree read/write, conversion to/from core, and **JSON schemas** under `json_schemas/` that must stay aligned with core shapes. When a core field changes type or semantics, update the matching `*.schema.json` (descriptions, `type`, cross-field references) in the same stage — not only Go conversion code.

**Parser AI input types mirror core structurally.** `inputClass`, `inputAttribute`, and sibling `input*` types should match core layout (ordered slices where core uses slices, unique keys where core requires them). Their role is to ease JSON authoring and schema validation — small JSON-friendly adjustments (string keys instead of `identity.Key`, omitted empty fields) are fine; structural deviation (e.g. core `[]Attribute` vs input map) should be rare and temporary. Example: class `attributes` in `class.schema.json` is an **array** of attribute objects, each with a `key` field; array order is declaration order, matching core `[]Attribute`.
5. **`internal/parser_human`** — YAML read/write and conversion to/from core.
6. **Everything else** — generate, simulator, notation, cmd, and any other packages that consume the model; add or extend tests for behavior touched in each package.

Each stage gets thorough tests for the changes made in that stage. Staged refactors still end with the build-gate closing bead that runs `apps/requirements/req/build.sh` (see [Quality gate](#quality-gate-appsrequirementsreqbuildsh)).

## Go member mutation

An object may always assign into its **own** fields (`useCase.Level = "mud"` is fine). The restriction is **member's member** — a nested field on a value object the parent owns. Reads of nested fields are fine anywhere; writes to nested fields go through a method on that owned value.

Bad: `subdomain.Classes[classKey].Attributes[attrKey] = attr` or `useCase.Actors[actorKey] = actor`

Good: `class.SetAttributes(attrs)` or `useCase.SetActors(actors)` (or a parent method that delegates to the member)

This applies in constructors too. A parent constructor that needs to set up a member must call that member's constructor or other methods — not assign into the member's fields directly.

**Tests are the exception.** `_test.go` files may assign into a member's fields directly when that keeps a test focused on the behavior under test.

## Go constructors

Every Go object must be instantiated through a constructor in its package (`NewModel`, `NewClass`, `NewUseCase`, etc.) — not a bare struct literal at the call site.

Bad: `class := model_class.Class{Key: key, Name: "Order", Details: details}`

Good: `class := model_class.NewClass(key, "Order", details, "", nil, nil, nil, "")`

When introducing a new type that holds state or behavior, add a constructor that wires valid initial values. Callers use that constructor; they do not assemble the struct themselves.

**Tests are the exception.** `_test.go` files may use struct literals for invalid-state cases or minimal setup when the test target is a constructor, validator, or single field.

## Documentation discipline

Comments and commit-style prose must explain the **why** of the current code, not its history. Write for two audiences: the human maintainer and a future AI session that inherits the repo without this conversation.

- Frame rationale in the present tense, describing the system as it stands now.
- Do not reference removed code, prior versions, or past behavior. A future reader will not have that context.
- Do not reference plans, prompts, tickets, or design documents used to produce the change. A future reader will not have them either.
- Add comments where a decision is not obvious from the code alone — invariants, parser marker ordering, rejected alternatives, coupling constraints, or non-local contracts.
- Keep comments **succinct**. State the decision and the reason; one or two sentences is usually enough.
- Do **not** narrate the **how** — control flow, field names, and step-by-step mechanics belong in the code, not the comment.
- Prefer comments on functions, types, and non-obvious branches over line-by-line restatement of what the next statement does.

Example. Instead of "we removed `PRM_NO_KYC` because Topgolf doesn't use it", write "Topgolf players are presumed KYC-cleared at the partner side; the partner-service capability gate therefore distinguishes only `POINTS_ONLY` (jurisdiction / junior / guest) from `REAL_MONEY` (everything else)".

Bad (how): `// Split on each marker, then trim whitespace from each section.`

Good (why): `// ⚠ must be parsed before ◆ so unfinished notes stay out of UML and YAML sections.`

## UML stereotypes

UML stereotypes in this repo use **guillemets** (`«»`), not ASCII angle brackets. When the user writes a stereotype as `<>`, `<<`, `<<>>`, or similar bracket shorthand, they mean the guillemet form — e.g. `<<association>>` → `«association»`, `<<actor>>` → `«actor»`. Rendered diagrams must show `«»`. Mermaid class diagrams encode stereotypes in source as `<<name>>` inside the class body; Mermaid draws them as `«name»` above the class title.

## Model-agnostic code (always)

**Production code in this repo must always be model-agnostic.** It must not hardcode knowledge of any particular model (evenplay, a sandbox fixture domain, a partner name, wallet class names, association labels like `Adjusts`/`Amounts`, attribute spellings unique to one domain, and so on). Tools, parsers, simulators, generators, and shared libraries behave correctly for **any** model that parses and validates — they must not special-case one authored model.

This applies to Go under `apps/requirements/req` and to any other production code that consumes or walks models. Domain content lives in model files (`data_sandbox/model/…`, fixtures, samples); engine code reads it at runtime.

- Derive names, labels, multiplicities, field lists, and human-readable text from model data (class `name`, association `name`, attribute keys, event parameter names, guarantees, and so on) — never from a fixed vocabulary of one domain.
- Do not hardcode model-specific strings in `switch` tables, constants, cascade paths, formatters, or control flow to special-case one model.
- Prefer generic mechanisms driven by model structure and authored guarantees (e.g. association bulk-create from a set-map over a parameter set) over embedding domain cascade rules in the engine.
- **Tests and sample models are the exception:** `_test.go` files, fixtures, and sandbox models may use concrete domain names and paths. Production packages must still treat those only as examples of arbitrary valid models.

## Simulation surface (what is being tested)

The **simulation surface** is the set of **external drivers** the exercise simulator may choose at the top level. Its purpose is to tell a **human tester what is actually being exercised** on a given run—not to dump every class that happens to be loaded in scope.

**Surface drivers (report these):**

- External **creation events** (`_new` / «new» and other create transitions with no in-scope sender)
- External **state-transition events** (and their actions) with no in-scope sender
- Surface **do-actions** on a state
- External **queries**
- External **derived attributes** (readable at top level)

**Not surface (do not list as drivers even if in the simulation):**

- Classes (or events) that only participate as **peers**—created or updated by another class’s guarantees (e.g. association set-add/set-map, cascade Delete, association-class reify)
- **Association classes** when they only materialize via a host association (even if included in the include-list so reify is enabled)
- Classes that are in scope for liveness or cascade but have **no external events, queries, or derived attributes**
- Derived attributes / queries that depend on **out-of-scope** classes (listed separately as off-surface, not as drivers)

**Include-list vs surface report:** `-include-class` / sandbox `SIMULATE_CLASSES` (and subdomain includes) define **scope**—which classes may exist in the run. The surface **drivers** section is narrower: only hooks the selector can fire. Peer-only scoped classes still run (instances, links, invariants) but must not appear as empty driver rows.

**Scope report (also shown to testers):** summarize participation separately from drivers.

- If **every** class of a subdomain is in the run, list the **subdomain** path only (`finance/wallet`)—do not enumerate its classes.
- If **only some** classes of a subdomain are in the run, list each **class** path (`finance/wallet/transaction`, …).

**Output order (text CLI):** completion summary → step trace / final state → **simulation scope** → **simulation surface (drivers)** → violations.

When changing surface reporting or selection, preserve this contract: scope shows what is loaded; surface shows what is driven at top level.

## Simulator `schema` package (sole model home for a run)

`apps/requirements/req/internal/simulator/schema` is the **sole home of model facts for one simulation run**. It owns the (typically surface-filtered) `*core.Model` and run-facing indexes (classes in scope, attributes, associations).

**Data-flow gate**

```text
core.Model  ──schema.NewFromModel──►  *schema.Schema  ──►  instance / engine / checkers / bindings
```

- **Intake:** `*core.Model` may be used only to build `*schema.Schema` (and one-shot surface resolution before that).
- **Run:** simulator components must not hold a second authoritative `*core.Model` for the same run. Prefer Schema methods. During migration, `Schema.CoreModel()` is the only legitimate model pointer—use it to construct catalog/checkers, then drop it.
- Do not mutate the model after `NewFromModel`.
- Do not own: live instances, links, SM positions (`instance`).

**Why:** clear gates so static rules cannot drift across parallel model copies (schema vs catalog vs checkers each re-walking a free model pointer).

## Simulator `instance` package

`apps/requirements/req/internal/simulator/instance` holds **all mutable state for one simulation run**: class instances, binary association links, association-class host rows, state-machine positions, and identity mappings used with that world.

**Boundary**

- Own: create/update/delete instances, association links, SM current state, clone of the run world.
- Read static model facts only through the attached `*schema.Schema` (never a free `*core.Model`).
- Do not own: action execution, expression evaluation, model loading, surface selection, or TLA bindings construction (`state.BindingsBuilder` adapts `instance.State` into evaluator bindings).
- Production callers depend on this package for run data; this package must not import `engine`, `actions`, `invariants`, or `trace`.

**API discipline**

- Prefer a small exported protocol (`State`, `Instance`, `ID`, association types and methods). Keep maps, locks, and ID counters unexported.
- Construct via `NewState(sch *schema.Schema)` and `State` methods in production code (aligns with [Go constructors](#go-constructors)).
- Godoc on the package and exported types is the contract; do not grow hidden public side doors without documenting them as temporary.

**Joint test curation (AI + human)**

Unit tests for this package are **jointly written and curated by AI and human** maintainers.

- AI may draft and extend tests when implementing the package.
- A human reviews, edits, and owns the suite’s intent (coverage gaps, naming, fragile cases, intentional non-coverage).
- A green AI-only test pass is not “test design done” for this package — human curation is part of done for test changes here.
- Prefer protocol-level tests of the exported API over white-box tests of unexported maps/locks unless a bug requires it.

Unit tests for `schema` follow the same joint AI/human curation expectation when the suite grows beyond bootstrap coverage.

## Go `_test.go` files

- Use the [testify](https://github.com/stretchr/testify) framework (`require` for fatal assertions, `assert` for non-fatal).
- Prefer table-driven tests. Each row is a named case (`name string` field) and the test body uses `t.Run(tc.name, ...)`.
- Use table-driven tests even for two cases when more cases are likely to be added.

## Complexity linter (`go-complexity-lint`)

These rules apply on the **build-gate closing bead** (see [Quality gate](#quality-gate-appsrequirementsreqbuildsh)) when `./apps/requirements/req/build.sh` reports `go-complexity-lint` findings. They govern how that bead fixes parameter-count and other complexity metrics — not how feature or staging beads shape code before the gate runs.

They apply **only** to [go-complexity-lint](https://github.com/glemzurg/go-complexity-lint). Do not use `//complexity:...` comments to silence **golangci-lint** or any other linter — for those, surface the warning and let the human decide (including `//nolint:...`).

### Never add complexity exceptions

**AI must never add `//complexity:...` directives** — not for constructors, routing switches, parameter counts, fan-out, or any other metric. That includes scoped overrides such as `//complexity:params:warn=8,fail=8` or `//complexity:fanout:warn=9,fail=9`. Do not add them proactively, do not add them to make `apps/requirements/req/build.sh` pass, and do not add them even if the human asks you to "just suppress it." Only a human maintainer may add or approve inline `go-complexity-lint` exceptions in this repo.

When `go-complexity-lint` reports a finding:

1. Run the gate and surface the warning with the affected function and the lint's suggested counts.
2. Fix the code by following the linter's remediation hint for that metric and avoid any anti-pattern described — do not invent a different strategy.
3. If a scoped override or project-wide threshold change is the right outcome, describe the trade-off and leave the decision to the human.

Existing `//complexity:...` comments in the codebase are human-approved; do not extend that pattern or copy it into new code.

## Quality gate: `apps/requirements/req/build.sh`

This script is the single source of truth for "is the requirements tooling in shippable shape?". It runs `go fmt`, `golangci-lint`, `go-complexity-lint`, the full `go test ./...` suite (including database tests), and `go install`, exiting 0 only when every stage is clean.

**Create a build-gate closing bead whenever this session changed code in or below `apps/requirements/req/`.** Doc-only edits, changes elsewhere in the monorepo, or read-only investigation do not require one. The closing bead runs **last** — after every implementation, staging, or doc bead for that work — and is the only bead that runs `./apps/requirements/req/build.sh` and fixes what it reports (including complexity findings per [Complexity linter](#complexity-linter-go-complexity-lint)).

**A block of work is not done until the closing bead closes with the script exiting 0** (when `apps/requirements/req/` code changed). "Almost passing" doesn't count. If the script is red, the work is not finished.

### Build-gate closing bead

Every bead, epic, or session that touched `apps/requirements/req/` ends with a **build-gate closing bead** whose job is to run `./apps/requirements/req/build.sh`, fix every problem it reports, and verify a clean exit. Work of any size gets this bead — not only epics.

- Create the closing bead when you know `req/` will change; claim it only after preceding beads are closed so it truly runs last.
- If the script exits 0, close the closing bead (and the parent epic, if any).
- If the script surfaces problems, create child beads under the closing bead (or parent epic) for each problem class and address them. **Do not close the closing bead until the script is clean.** Keeping it open is the visible signal that the work is not yet shippable.
- Re-use the same closing bead across fix iterations — don't open a fresh "verify again" bead each round. The current open closing bead IS the verification gate.
- Complexity refactors belong on this bead: follow the [Complexity linter](#complexity-linter-go-complexity-lint) rules there, not on feature beads that precede it.

Epic layout and dependency wiring for multi-step work are documented in `.claude/skills/bead-epic/SKILL.md`.

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

- **Beads is mandatory** — see [Beads for everything](#beads-for-everything) above. No exceptions for "quick" tasks.
- Use `bd` for ALL task tracking — do NOT use TodoWrite, TaskCreate, or markdown TODO lists
- Run `bd prime` at session start for detailed command reference and session-close protocol
- Use `bd remember` for persistent knowledge — do NOT use MEMORY.md files

**Architecture in one line:** issues live in a local Dolt DB; sync uses `refs/dolt/data` on your git remote; `.beads/issues.jsonl` is a passive export. See https://github.com/gastownhall/beads/blob/main/docs/SYNC_CONCEPTS.md for details and anti-patterns.

## Session Completion

**When ending a work session**, you MUST complete ALL steps below. Work is NOT complete until all quality gates succeed.

**MANDATORY WORKFLOW:**

1. **File issues for remaining work** - Create issues for anything that needs follow-up
2. **Close the build-gate bead** (if code changed under `apps/requirements/req/`) — its `./apps/requirements/req/build.sh` run must exit 0; fix every issue it reports before closing that bead
3. **Update issue status** - Close finished work, update in-progress items
5. **Clean up** - Clear stashes
7. **Hand off** - Provide context for next session

**CRITICAL RULES:**
- Work is NOT complete until all quality gates succeed
<!-- END BEADS INTEGRATION -->