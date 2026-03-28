# Action Duplicate Name (E8007)

Two action files in the same class have different keys but the same `name` value. This produces duplicate map keys in YAML output.

## How to Fix

Rename one of the actions so each has a unique `name` value within the class. Do not delete or remove either action to resolve this error — both actions exist for a reason and the model needs them. Instead, adjust the `name` field in one of the conflicting action files to distinguish them.
