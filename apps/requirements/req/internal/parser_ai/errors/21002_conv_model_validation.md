# Internal Validation Error (E21002)

An unexpected internal validation error occurred that cannot be fixed by changing inputs.

## What Went Wrong

After converting all parsed JSON input into the internal model representation, the final model validation produced an error that does not have a specific error mapping. This indicates either an unmapped core validation code or an unexpected error type within the validation system.

## What To Do

**Do not continue changing inputs — this error will not be corrected by modifying the model files.** Have a human review the error message and the tool's error mapping to determine what went wrong.

The error message and hint will contain the internal error code that was not mapped, which a developer can use to add the missing mapping.
