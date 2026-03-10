# Scenario Step Invalid (E21111)

A scenario step has a structural rule violation.

## What Went Wrong

Scenario steps must follow specific structural rules depending on their type:

- **Sequence** steps must have at least 2 statements
- **Switch** steps must have at least 1 case
- **Loop** steps must have a condition and at least 1 statement
- **Case** steps must have a condition
- **Leaf** steps must have the correct fields for their leaf_type (event, query, scenario, delete)
- A scenario step cannot reference its own scenario

## How to Fix

Check the error message for the specific rule violated. Ensure:

- Event steps have `from_object_key`, `to_object_key`, and `event_key`
- Query steps have `from_object_key`, `to_object_key`, and `query_key`
- Scenario steps have `from_object_key`, `to_object_key`, and `scenario_key`
- Delete steps have `from_object_key` only
- Compound steps meet their minimum child count requirements

## Related Errors

- **E19004**: Scenario schema violation
