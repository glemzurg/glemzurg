# Logic Delete Event Not Allowed (E14010)

A logic specification that is not type `delete` declares `delete_event`. Peer `_delete` calls are only allowed on `delete` guarantees (or surface actions), not inline on other logic kinds.

## How to Fix

Either change the logic `type` to `delete` and move the selection into `specification`, or remove `delete_event` and use a separate `delete` guarantee for peer removal.