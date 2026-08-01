# Association-class `_new`: accepted design (surface = data + link)

Short contract for authors and engine work. Audience: someone new to the simulator.

Related: [Association navigations](design-association-relations-id-data.md), AGENTS simulation surface rules.

---

## Goal

Creating an **association class** instance is one surface step:

1. Create the association-class **data** (the row: attributes, state machine).
2. Create the **host association link** (the two endpoints are now related *through* this row).

That link construction is an **engine capability**, not something TLA+ guarantees spell out as “add binary edge then create peer.”

Testers exercise **association-class `_new` on the surface**. Hosts do **not** create the association class via Configure / reify / set-add of the AC.

---

## Accepted decisions

| ID | Choice | Meaning |
|----|--------|---------|
| **A1** | Endpoints as `_new` params | Author lists the two ends (and any extra args) on `_new` / Initialize |
| **B1** | Engine-only link | TLA Initialize sets AC attrs and deeper cascades only; engine builds the host link |
| **C1** | No host Configure-create | Remove host create/reify paths that invent AC rows |
| **D2** | Derived host image | `self.HostAssoc` is **only** a view over AC rows (no separate “link without row”) |
| **E1** | AC `_new` on surface | Association-class creation is a top-level driver; not peer-sent from host |
| **F1** | Nested create in Initialize | After the link exists, AC Initialize may multi set-add / type:events as normal TLA |
| Package | **P-Surface-AC** | Stack above |

---

## Mental model (one picture)

```text
Surface:  CurrencyWalletDefinition._new(JWD, Currency, ExistingJWs)
              │
              ├─ engine (outside TLA): materialize host row
              │     JWD ──«Is Subdivided Into»── Currency
              │              └── AC instance (this)
              │
              └─ TLA Initialize: attrs + optional Defines ∪ { _new(w) : … }
```

Navigation later:

```text
jwd.IsSubdividedInto              → set of Currency endpoints (derived from AC rows)
jwd.IsSubdividedInto.CurrencyWalletDefinition
                                 → AC row(s) for those links
```

There is **no** stored host edge that is not implied by an association-class instance.

---

## Authoring rules

### Association class `_new` signature

1. **First two parameters** (or clearly named parameters) identify the **from** and **to** endpoints of the host association.
2. Types must match the host association’s endpoint classes (schema / catalog checks).
3. Further parameters are ordinary event args (sets, numbers, …) for Initialize only.

Example shape (illustrative):

```yaml
# On the association-class class (e.g. Jurisdictional Wallet Definition)
events:
  _new:
    parameters:
      - Partner          # host from-end
      - Jurisdiction     # host to-end
      - ExistingPlayers  # F1: used only in Initialize TLA
actions:
  Initialize:
    parameters: … same …
    guarantees:
      # B1: do NOT write host link guarantees here
      # F1: nested materialization OK, e.g.
      # - target: Defines
      #   specification: 'Defines \union { _new(p) : p \in ExistingPlayers }'
```

### What authors must not write

- Host `endpoint_selector` + reify `_new` to create the AC.
- Host `Configures… \union {FarEnd}` as the *only* way to introduce an AC pair (image is derived).
- TLA that “manually” creates a host link for an association that has `association_class_key`.

### Host class role after this design

- **Create:** none for the AC (C1).
- **Still allowed:** events on the host that *use* existing links (queries, cascades Delete/Recover over `HostAssoc.ACMember`, business events).
- Multiplicity / uniqueness still enforce “at most one AC row per endpoint pair” (and uniqueness attrs) on create.

### Surface (E1)

| On surface as creation driver | Not a surface creation driver |
|-------------------------------|--------------------------------|
| Association-class `_new` | Host Configure* that used to reify |
| | AC `_new` “sent by” host set-add/reify |

Scope may still include host and AC classes so cascades and navigation run; **drivers** for creating AC rows are AC `_new` only.

---

## Engine contract (implementers)

### On association-class creation (`_new`)

1. Resolve **from** and **to** instance IDs from the designated endpoint parameters (A1).
2. Reject if either endpoint missing, wrong class, or pair already has an AC row (uniqueness).
3. Create AC instance (attrs, state) as today.
4. Register **one** association-class host row: `(hostAssoc, from, to, acInstance)` (B1).
5. Do **not** add a plain binary link for that host association (D2: image comes from AC rows only).
6. Run Initialize **after** the row is visible to navigation (so F1 set-adds can see ends / reverse links if needed).

### Host association navigation (D2)

- `GetRelatedRecords` / host field access for an AC host: **derive endpoints from association-class rows** for that host key and anchor.
- `GetAssociationClassLinksByEndpoint`: same rows; keep extent-id resolution for set clones.
- Projecting to relation context: build endpoint image **from** AC rows, not from a parallel binary link table for that host.

### Remove / stop using

- Association-class **reify** guarantee path as the way to create AC rows from host actions.
- Host create actions in models that only existed to reify (evenplay: `ConfigureJurisdiction`, `ConfigureCurrency`, and any remaining reify-only Configure*).

Plain associations (no `association_class_key`) are unchanged: set-add / state_change still create binary links.

---

## Evenplay migration sketch

| Class | Today | Target |
|-------|--------|--------|
| Partner | `ConfigureJurisdiction` reifies JWD | Drop create; **JWD `_new(Partner, Jurisdiction, ExistingPlayers)`** on surface |
| JWD | `ConfigureCurrency` reifies CWD | Drop create; **CWD `_new(JWD, Currency, ExistingJWs)`** on surface |
| CWD | `ConfigureAccount` set-adds AccountDefinition | AccountDefinition is **not** an AC today — leave as plain host set-add **or** fold later; not required by this AC design |
| Transaction | reify Account Balance Change | **ABC `_new`…** with ends as params (surface), or keep only if product still wants host-driven posts — align with E1 when touched |
| Nested F1 | JWD/CWD Initialize multi set-add wallets | Keep on AC Initialize after link exists |

Account surface `_new(Definition, Wallet)` stays a **plain** reverse-link pattern; it is not an association-class `_new` unless Account is later remodelled as AC.

---

## Test expectations (acceptance)

1. Surface report lists AC class **creation: `_new`**, not host Configure-create.
2. After surface AC `_new(from, to, …)`, host navigation shows `to` in the from-end image and AC member projects the row.
3. No host binary link row for that association without an AC instance (D2).
4. Duplicate `_new` for same pair fails uniqueness / pair-exists.
5. F1: Initialize multi set-add still runs and can navigate host/AC if needed.
6. Set-cloned anchors still resolve AC rows (existing extent-id fix).

---

## Out of scope (for follow-ups)

- Making AccountDefinition an association class.
- Engine cascade tables keyed by domain class names (rejected as F3).
- Changing plain association set-add semantics.

---

## Implementation order (suggested)

1. **Engine D2 + B1:** create path materializes AC row only; host image derived from rows; drop dual binary write for AC hosts.
2. **Surface E1:** classify AC `_new` as external creation; stop treating host reify as sender of AC `_new`.
3. **Retire reify-create (C1)** in actions + tests.
4. **Model evenplay** to A1 signatures and remove host Configure-create.
5. **Build-gate** for `apps/requirements/req/`.
