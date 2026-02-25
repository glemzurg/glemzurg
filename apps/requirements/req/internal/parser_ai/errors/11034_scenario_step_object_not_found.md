# ErrTreeScenarioStepObjectNotFound (11034)

Description

- A scenario step references an `object` key (via `from_object_key` or `to_object_key`) that is not defined in the scenario's `objects` map.

Cause

- The step uses an object identifier that does not exist in the scenario file (typo, refactor mismatch, or missing object definition).

Why this matters

- Steps in scenarios reference objects to resolve classes, events, and queries. If the object is missing, subsequent validation cannot determine which class the step refers to.

How to fix

- Add the missing object definition to the scenario's `objects` section with the referenced key.
- Fix typos in the step's `from_object_key` / `to_object_key`.

Example

- Invalid: step `from_object_key: buyer` but `objects` contains only `customer`.
- Fix: rename or add the appropriate object entry.
