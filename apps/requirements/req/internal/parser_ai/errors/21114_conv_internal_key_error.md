# Internal Key Validation Error (E21114)

An internal key validation check failed during model validation.

## What Went Wrong

This error indicates an internal consistency failure in the model's key structure. It should not normally occur if the model files are well-formed. If you see this error, it likely indicates a bug in the model conversion process rather than a problem with your input files.

## How to Fix

Verify that all entity keys (filenames, directory names) follow the required `snake_case` format and that no keys are empty. If the error persists, report it as it may indicate an internal issue.

## Related Errors

- **E11026**: Key has invalid format
- **E21001**: Identity key construction failed
