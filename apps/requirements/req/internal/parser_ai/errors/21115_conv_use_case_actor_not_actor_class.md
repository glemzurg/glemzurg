# Use Case References Non-Actor Class (E21115)

A use case is defined under a class that is not an actor class.

## What Went Wrong

Use cases can only be defined under classes that have an `actor_key` set. The class this use case belongs to does not reference any actor.

## How to Fix

Either:
1. Add an `actor_key` to the class's `class.json` that references a defined actor
2. Move the use case to a class that already has an `actor_key`

## Related Errors

- **E5007**: Class actor not found
- **E18001**: Use case name required
