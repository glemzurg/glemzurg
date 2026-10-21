# Actor Filename Invalid (E2008)

The actor filename does not follow the expected pattern `<key>.actor.json`.

## How to Fix

Rename the file to match the pattern:

```
actors/
├── customer.actor.json           <- Valid
├── inventory_system.actor.json   <- Valid
└── BadName.json                  <- Invalid: must end with .actor.json
```

Keys must be lowercase snake_case: `^[a-z][a-z0-9]*(_[a-z0-9]+)*$`
