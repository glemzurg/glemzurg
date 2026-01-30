# State Machine Schema Violation (E7002)

The state machine JSON file contains valid JSON but does not conform to the expected schema.

## What Went Wrong

The parser successfully read your state machine file as valid JSON, but its structure or content violates the schema rules. This typically means:

- A required field is missing
- A field has the wrong type
- An unknown field is present
- A field value doesn't meet constraints

## File Location

State machine files are located alongside their class files:

```
your_model/
├── model.json
└── order_management/
    ├── domain.json
    ├── order.class.json
    └── order.state_machine.json    <-- This file violates the schema
```

## Schema Requirements

### States

Each state must have:
- `name` (required, minLength: 1)
- `details` (optional)
- `uml_comment` (optional)
- `actions` (optional array)

### Events

Each event must have:
- `name` (required, minLength: 1)
- `details` (optional)
- `parameters` (optional array)

### Guards

Each guard must have:
- `name` (required, minLength: 1)
- `details` (required, minLength: 1)

### Transitions

Each transition must have:
- `event_key` (required, minLength: 1)
- `from_state_key` (optional, null for initial)
- `to_state_key` (optional, null for final)
- `guard_key` (optional)
- `action_key` (optional)
- `uml_comment` (optional)

## Common Schema Violations

### 1. State Missing Name

```json
// WRONG
{
    "states": {
        "pending": {
            "details": "Order is pending"
        }
    }
}

// CORRECT
{
    "states": {
        "pending": {
            "name": "Pending",
            "details": "Order is pending"
        }
    }
}
```

### 2. Event Missing Name

```json
// WRONG
{
    "events": {
        "submit": {
            "details": "User submits"
        }
    }
}

// CORRECT
{
    "events": {
        "submit": {
            "name": "Submit",
            "details": "User submits"
        }
    }
}
```

### 3. Guard Missing Details

```json
// WRONG
{
    "guards": {
        "is_valid": {
            "name": "Is Valid"
        }
    }
}

// CORRECT
{
    "guards": {
        "is_valid": {
            "name": "Is Valid",
            "details": "Checks all validation rules pass"
        }
    }
}
```

### 4. Invalid Action When Value

```json
// WRONG: "when" must be "entry", "exit", or "do"
{
    "states": {
        "pending": {
            "name": "Pending",
            "actions": [
                {
                    "action_key": "notify",
                    "when": "always"
                }
            ]
        }
    }
}

// CORRECT
{
    "states": {
        "pending": {
            "name": "Pending",
            "actions": [
                {
                    "action_key": "notify",
                    "when": "entry"
                }
            ]
        }
    }
}
```

### 5. Unknown Field

```json
// WRONG: "type" is not in the schema
{
    "states": {
        "pending": {
            "name": "Pending",
            "type": "initial"
        }
    }
}

// CORRECT
{
    "states": {
        "pending": {
            "name": "Pending"
        }
    }
}
```

## Valid Example

```json
{
    "states": {
        "pending": {
            "name": "Pending",
            "details": "Order awaiting approval"
        },
        "approved": {
            "name": "Approved"
        }
    },
    "events": {
        "approve": {
            "name": "Approve"
        }
    },
    "guards": {
        "is_valid": {
            "name": "Is Valid",
            "details": "Order passes validation"
        }
    },
    "transitions": [
        {
            "from_state_key": "pending",
            "to_state_key": "approved",
            "event_key": "approve",
            "guard_key": "is_valid"
        }
    ]
}
```

## Related Errors

- **E7001**: Invalid JSON syntax
- **E7003**: State name is missing
- **E7009**: Event name is missing
- **E7014**: Guard name is missing
