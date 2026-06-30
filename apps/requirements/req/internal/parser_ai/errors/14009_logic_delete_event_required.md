# Logic Delete Event Required (E14009)

A logic specification of type `delete` is missing the required `delete_event` field.

## How to Fix

Add `delete_event` with the peer delete event call. Keep the set-filter selection in `specification`:

```json
{
    "type": "delete",
    "description": "Peer _delete events for removed peers",
    "target": "AppliesSocialCurrencyLogic",
    "notation": "tla_plus",
    "specification": "{ b \\in AppliesSocialCurrencyLogic : TRUE }",
    "delete_event": "_delete(b)"
}
```