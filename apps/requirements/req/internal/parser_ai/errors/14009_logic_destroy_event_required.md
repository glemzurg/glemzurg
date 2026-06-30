# Logic Destroy Event Required (E14009)

A logic specification of type `destroy` is missing the required `destroy_event` field.

## How to Fix

Add `destroy_event` with the peer destroy event call. Keep the set-filter selection in `specification`:

```json
{
    "type": "destroy",
    "description": "Peer _destroy events for removed peers",
    "target": "AppliesSocialCurrencyLogic",
    "notation": "tla_plus",
    "specification": "{ b \\in AppliesSocialCurrencyLogic : TRUE }",
    "destroy_event": "_destroy(b)"
}
```