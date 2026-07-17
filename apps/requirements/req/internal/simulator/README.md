# Exercise simulator

The simulator exercises a scoped slice of a requirements model by randomly firing
eligible surface actions, then reports correctness violations and coverage gaps.

## Surface selection

Each step picks one eligible action uniformly at random. The pool is built from
every simulatable class in the surface:

| Kind | Eligibility |
|------|-------------|
| **Initial transition** | External creation events only — not sent by another simulatable in-scope class, and not driven by a mandatory outbound association |
| **State transition** | External events on an **existing** instance in the matching source state |
| **Query** | External queries on an **existing** instance (no simulatable in-scope caller) |
| **Do-action** | Always surface-level on an **existing** instance while in the matching state |

SentBy and CalledBy metadata is derived from use-case scenario steps and mandatory
association creation chains. Actor-only classes (no state machine) are recorded as
senders but do not suppress external firing, because they cannot participate in the
simulation loop.

Creation never invents instances for classes without initial transitions. Cascaded
creation (mandatory associations) runs inside a creation step, not as a top-level pick.

## Liveness (coverage)

After the run, liveness checks the **whole scoped subdomain** — every class and
association in the surface, including stateless classes that cannot be simulated.
Violations are intentional development signals: they show where the subdomain still
lacks exercised logic. Only simulatable classes participate in the run itself.

| Check | What it means |
|-------|----------------|
| Class not instantiated | No instance of this class was created |
| Association not linked | No link was created for this association |
| Event not sent | A declared event never fired |
| Query not run | A declared query never executed |
| Action not executed | A declared action never ran (transition, entry, exit, or do) |
| Attribute not written | A non-derived attribute never received a primed write |

A clean simulation run is rare for a partially modeled subdomain. Treat liveness
output as a coverage map for what to model or wire next.

## Class extents in TLA+ (id vs data)

Keep **representation** and **author-facing access** distinct:

| | |
|--|--|
| **Internal** | Instances are id → attribute data; associations link id → id |
| **Class name in TLA** (e.g. `Account`) | A **set of ordinary records** `[id |-> N, data |-> attrs]` — not a special Map type |
| **`self` on a live instance** | The **data** record only (`self.amount`), so instance specs stay flat |

Patterns:

```tla
_FiniteSets!Cardinality(Account) >= 3
CHOOSE a \in Account : TRUE          \* a = [id |-> …, data |-> …]
a.id
a.data._state
```

See `docs/design-association-relations-id-data.md` for associations and Approach A quantifiers over navigations.

## CLI

See `cmd/simulate` and `scripts/simulate.sh` for `-include-subdomain`,
`-include-class`, seeds, and trace output. Clean runs (no violations) print the
full step trace by default; use `-trace` to force it when violations are present.