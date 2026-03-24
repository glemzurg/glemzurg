# Use Case Shared Share Type Required (E20001)

The use case shared relationship is missing the required `share_type` field.

## How to Fix

Add a `share_type` field. Valid values: `"include"`, `"extend"`.

```json
{
    "share_type": "include",
    "shared_use_case_key": "validate_payment"
}
```
