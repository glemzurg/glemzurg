# Logic Association Class Spec Required (E14014)

An association-class reify guarantee needs a `specification` with the association-class creation event call (e.g. `_new(r.amount)` or a set-map of `_new`).

## How to Fix

Add `specification` containing the AC `_new(...)` (or `«new»(...)`) call, or `{ _new(...) : r \in Domain }`.
