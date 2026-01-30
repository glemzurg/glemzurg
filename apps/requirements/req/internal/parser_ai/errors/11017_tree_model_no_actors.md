# Model Has No Actors (E11017)

The model must have at least one actor defined to describe who interacts with the system.

## What Went Wrong

The model's `actors/` directory is empty or does not contain any `.actor.json` files. Every model needs at least one actor to represent the users, systems, or external entities that interact with the application.

## Context

Actors represent the entities that interact with your system. They are essential for understanding:
- Who uses the system
- What systems integrate with yours
- External services that trigger behavior

```
your_model/
├── model.json
├── actors/                       <-- This directory must contain at least one actor
│   ├── customer.actor.json       <-- Example: A human user
│   └── payment_gateway.actor.json <-- Example: An external system
└── domains/
    └── ...
```

## How to Fix

### Create an Actor File

Add at least one actor file in the `actors/` directory:

```
actors/{actor_key}.actor.json
```

For a human user:

```json
{
    "name": "Customer",
    "type": "person",
    "details": "A user who browses products and places orders"
}
```

For an external system:

```json
{
    "name": "Payment Gateway",
    "type": "external_system",
    "details": "Third-party service that processes credit card payments"
}
```

For a time-based trigger:

```json
{
    "name": "Daily Scheduler",
    "type": "time",
    "details": "Triggers daily report generation at midnight"
}
```

## Actor Types

Valid actor types include:
- `person` - A human user of the system
- `external_system` - An external service or API
- `time` - Time-based triggers (cron jobs, scheduled tasks)

## Why Actors Matter

Actors help AI understand:
1. **Use cases**: What actions each type of user needs to perform
2. **Security boundaries**: Who has access to what functionality
3. **Integration points**: How external systems connect
4. **Event sources**: What triggers behavior in the system

## Common Actors to Consider

- **Customers**: End users who consume your service
- **Administrators**: Users with elevated privileges
- **Support Staff**: Internal users helping customers
- **Payment Systems**: Stripe, PayPal, etc.
- **Email Services**: SendGrid, SES, etc.
- **Analytics**: Data collection and reporting systems
- **Schedulers**: Background job triggers

## Related Errors

- **E2001**: Actor name is required
- **E2003**: Actor type is required
- **E11001**: Class references a non-existent actor
