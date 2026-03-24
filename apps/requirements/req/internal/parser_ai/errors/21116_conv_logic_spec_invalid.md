# Logic Specification Invalid (E21116)

A logic item's specification (TLA+ expression) or type specification failed internal validation.

## What Went Wrong

After parsing, the TLA+ expression or type specification was found to have structural issues in its abstract syntax tree. This typically means the expression parsed but produced an invalid internal representation.

## How to Fix

Check the error message for details about which expression or type spec failed. Common causes:

- Missing required sub-expressions in complex TLA+ constructs
- Invalid operator usage
- Malformed set, tuple, or record expressions
- Data type validation failures in parsed data type rules

Review and simplify the TLA+ expression, then retry.

## Related Errors

- **E21102**: Logic type invalid for context
