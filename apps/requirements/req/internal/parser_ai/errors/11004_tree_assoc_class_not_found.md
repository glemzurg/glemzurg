# Association Class Not Found (E11004)

An association references an association class (`association_class_key`) that does not exist.

## What Went Wrong

An association file has an `association_class_key` field that references a class to serve as the association class, but that class does not exist at the expected location.

## What is an Association Class?

An association class is a class that provides additional attributes or behavior for an association. For example, a "Registration" class might be the association class between "Student" and "Course", holding attributes like "grade" and "enrollment_date".

```
Student ----<Registration>---- Course
              - grade
              - enrollment_date
```

## Key Formats by Scope

| Scope | Key Format | Example |
|-------|-----------|---------|
| Subdomain | `class_name` | `registration` |
| Domain | `subdomain/class` | `enrollments/registration` |
| Model | `domain/subdomain/class` | `academic/enrollments/registration` |

## How to Fix

### Option 1: Create the Missing Association Class

Create the class that will serve as the association class:

```
domains/{domain}/subdomains/{subdomain}/classes/{class_name}/
└── class.json
```

### Option 2: Fix the Reference

Update the association to reference an existing class:

```json
{
    "name": "Enrollment",
    "from_class_key": "student",
    "to_class_key": "course",
    "association_class_key": "existing_class"
}
```

### Option 3: Remove the Association Class

If no association class is needed, remove the field or set it to null:

```json
{
    "name": "Enrollment",
    "from_class_key": "student",
    "to_class_key": "course",
    "association_class_key": null
}
```

## Troubleshooting Checklist

1. **Check spelling**: Ensure the class key matches exactly
2. **Check key format**: Ensure the scope matches the association's location
3. **Check class exists**: The association class must be a valid class in the model

## Important: Resolve Each Issue Individually

Do not attempt to fix multiple association errors in a single bulk operation. Each association is distinct — a bulk move or rename will often fix some associations correctly while breaking others, creating more errors to fix later. Address each association error one at a time, verifying correctness before moving to the next.

## Related Errors

- **E11002**: Association from_class_key not found
- **E11003**: Association to_class_key not found
