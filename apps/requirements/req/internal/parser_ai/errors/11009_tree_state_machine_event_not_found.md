# State Machine Event Not Found (E11009)

A transition references an event that does not exist in the state machine.

## What Went Wrong

A `state_machine.json` file has a transition with an `event_key` that references an event, but no event with that key exists in the `events` map.

## How Events Work

Events trigger transitions between states. Every transition must reference a valid event defined in the `events` map.

```json
{
    "events": {
        "confirm": {
            "name": "confirm",
            "details": "Order payment confirmed"
        },
        "ship": {
            "name": "ship",
            "details": "Order shipped"
        }
    },
    "transitions": [
        {
            "from_state_key": "pending",
            "to_state_key": "confirmed",
            "event_key": "confirm"           // Must exist in events
        }
    ]
}
```

## How to Fix

### Option 1: Add the Missing Event

Add the event to the `events` map:

```json
{
    "events": {
        "missing_event": {
            "name": "missingEvent",
            "details": "Description of when this event occurs"
        }
    }
}
```

### Option 2: Fix the Reference

Update the transition to reference an existing event:

```json
{
    "transitions": [
        {
            "from_state_key": "pending",
            "to_state_key": "confirmed",
            "event_key": "existing_event"
        }
    ]
}
```

## Troubleshooting Checklist

1. **Check spelling**: Event keys are case-sensitive
2. **Check event exists**: The event must be defined in the `events` map
3. **Check key vs name**: Use the event's key, not its display name

## Common Mistakes

```json
// WRONG: Using event name instead of key
{
    "events": {
        "order_confirmed": {"name": "Order Confirmed"}
    },
    "transitions": [
        {"event_key": "Order Confirmed", ...}
    ]
}
// Should be:
{
    "transitions": [
        {"event_key": "order_confirmed", ...}
    ]
}

// WRONG: Typo in event key
{
    "transitions": [
        {"event_key": "confrim", ...}
    ]
}
```

## Related Errors

- **E11008**: Transition state not found
- **E11010**: Transition guard not found
- **E7009**: Event name is required
