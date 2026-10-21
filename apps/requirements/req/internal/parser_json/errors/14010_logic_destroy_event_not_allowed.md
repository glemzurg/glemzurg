# Logic Destroy Event Not Allowed (E14010)

A logic specification that is not type `destroy` declares `destroy_event`. Peer `_destroy` calls are only allowed on `destroy` guarantees (or surface actions), not inline on other logic kinds.

## How to Fix

Either change the logic `type` to `destroy` and move the selection into `specification`, or remove `destroy_event` and use a separate `destroy` guarantee for peer removal.