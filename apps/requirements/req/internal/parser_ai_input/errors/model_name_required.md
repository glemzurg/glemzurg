# Model Name Required (E1001)

The model.json file is missing the required `name` field.

## Solution

Add a `name` field to your model.json file:

```json
{
    "name": "Your Model Name",
    "details": "Optional description"
}
```

## Why This Field is Required

The model name is used to identify the model throughout the system. It appears in:

- Generated documentation
- UML diagrams
- Error messages
- Database references
