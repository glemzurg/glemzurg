# Actor Generalization Name Empty (E12002)

The actor generalization JSON file has a `name` field that is empty or contains only whitespace.

## What Went Wrong

The parser found a `name` field in your actor generalization file, but its value is either an empty string (`""`) or contains only whitespace characters. The actor generalization name must contain at least one visible character.

## File Location

Actor generalization files are located in the `actor_generalizations/` directory:

```
your_model/
├── model.json
└── actor_generalizations/
    └── user_types.actor_gen.json    <-- This file has an empty "name" value
```

## How to Fix

Provide a meaningful, non-empty name for your actor generalization:

```json
{
    "name": "User Types",
    "superclass_key": "customer",
    "subclass_keys": ["premium_customer", "guest"]
}
```

## Invalid Examples

These values will all trigger this error:

```json
{"name": ""}              // Empty string
{"name": "   "}           // Spaces only
{"name": "\t"}            // Tab only
```

## Valid Examples

The name must contain at least one non-whitespace character:

```json
{"name": "User Types"}           // Typical name
{"name": "Staff Hierarchy"}      // Descriptive name
{"name": "Staff Roles"}          // Classification name
```

## Choosing a Good Actor Generalization Name

A good actor generalization name should:

1. **Describe the classification**: What distinguishes the child actors?
2. **Be concise**: Aim for 2-4 words
3. **Use noun phrases**: e.g., "User Types", "Staff Hierarchy"
4. **Reflect the discriminator**: What attribute differentiates the child actors?

### Good Name Examples

| Name | Superclass | Subclasses |
|------|------------|------------|
| `"User Types"` | customer | premium_customer, guest |
| `"Staff Hierarchy"` | staff | admin, manager, support_agent |
| `"Staff Roles"` | employee | developer, designer, tester |
| `"System Actors"` | system | scheduler, notifier, auditor |

## Complete Schema

| Field | Type | Required | Constraints |
|-------|------|----------|-------------|
| `name` | string | **Yes** | `minLength: 1` |
| `superclass_key` | string | **Yes** | `minLength: 1` |
| `subclass_keys` | string[] | **Yes** | `minItems: 1`, each item `minLength: 1` |
| `details` | string | No | None |
| `is_complete` | boolean | No | Default: false |
| `is_static` | boolean | No | Default: false |
| `uml_comment` | string | No | None |

## Related Errors

- **E12001**: Actor generalization name field is missing entirely
- **E12003**: Invalid JSON syntax
- **E12004**: Schema violation
