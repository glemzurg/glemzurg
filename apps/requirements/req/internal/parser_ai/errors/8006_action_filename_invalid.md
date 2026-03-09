# Action Filename Invalid (E8006)

The action filename does not follow the expected `<key>.json` pattern.

## How to Fix

Action filenames must be lowercase snake_case with `.json` extension:

```
actions/
├── calculate_total.json    <- Valid
└── BadName.JSON            <- Invalid
```
