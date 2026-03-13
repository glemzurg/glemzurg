# ErrTreeScenarioStepQueryNotFound (11036)

Description

- A scenario step references a `query_key` that does not exist on the resolved class.

Cause

- The step's `query_key` is incorrect or the resolved class does not define the referenced query in its `queries` map.

Why this matters

- Queries are part of the class API used in scenarios. If a query is missing, the scenario cannot be validated or executed by tooling.

How to fix

- Add the missing query to the referenced class's `queries` (create `queries/<key>.json` and ensure parsing), or correct the `query_key` in the step.

Example

- Invalid: step references `query_key: get_total` but the class defines `calculate_total`.
- Fix: change to `calculate_total` or add `get_total` to the class queries.

Important: Resolve each issue individually

Do not attempt to fix multiple scenario step errors in a single bulk operation. Each step references distinct objects, events, and queries — a bulk change will often fix some steps correctly while breaking others, creating more errors to fix later. Address each step error one at a time, verifying correctness before moving to the next.
