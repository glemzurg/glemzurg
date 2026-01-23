# Model Name Empty (E1002)

The model.json file has a `name` field that is empty or contains only whitespace.

## Solution

Provide a meaningful name for your model:

```json
{
    "name": "Your Model Name",
    "details": "Optional description"
}
```

## Invalid Examples

```json
{"name": ""}           // Empty string
{"name": "   "}        // Whitespace only
{"name": "\t\n"}       // Only whitespace characters
```

## Valid Examples

```json
{"name": "Web Store"}
{"name": "Order Management System"}
{"name": "Customer Portal"}
```
