# Logic Delete Event Required (E14009)

A logic specification of type `delete` is missing the required `destroy_event` field.

## How to Fix

Add `destroy_event` with the peer delete event call. Keep the set-filter selection in `specification`:

```json
{
    "type": "delete",
    "description": "Peer _destroy events for removed peers",
    "target": "AppliesSocialCurrencyLogic",
    "notation": "tla_plus",
    "specification": "{ b \\in AppliesSocialCurrencyLogic : TRUE }",
    "destroy_event": "_destroy(b)"
}
```