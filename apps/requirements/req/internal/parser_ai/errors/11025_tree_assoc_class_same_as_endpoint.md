# Association Class Same As Endpoint (E11025)

An association's association_class_key cannot be the same as its from_class_key or to_class_key.

## What Went Wrong

An association has an `association_class_key` that references the same class as either `from_class_key` or `to_class_key`. The association class must be a distinct third class that represents additional data about the relationship between the two endpoint classes.

## Context

Association classes are used when a many-to-many relationship has its own attributes. The association class sits "on the line" between two classes and cannot be one of the endpoint classes itself.

```json
{
    "name": "Enrollment",
    "from_class_key": "student",
    "to_class_key": "course",
    "from_multiplicity": "*",
    "to_multiplicity": "*",
    "association_class_key": "enrollment"    <-- Must be different from student and course
}
```

## How to Fix

### Step 1: Identify the Role of Each Class

- **Endpoint classes** (`from_class_key`, `to_class_key`): The two classes being related
- **Association class** (`association_class_key`): A class that holds data about the relationship

### Step 2: Create a Separate Association Class

If you need an association class, create a distinct class:

```json
{
    "name": "Enrollment",
    "from_class_key": "student",
    "to_class_key": "course",
    "from_multiplicity": "*",
    "to_multiplicity": "*",
    "association_class_key": "enrollment"
}
```

Where `enrollment` is a separate class with attributes like `grade`, `enrollment_date`, etc.

### Step 3: Or Remove the Association Class

If you don't need additional data on the relationship, simply omit the association_class_key:

```json
{
    "name": "Takes",
    "from_class_key": "student",
    "to_class_key": "course",
    "from_multiplicity": "*",
    "to_multiplicity": "*"
}
```

## Common Mistakes

### Wrong: Association Class Same as From Class
```json
{
    "name": "Has Orders",
    "from_class_key": "customer",
    "to_class_key": "order",
    "association_class_key": "customer"    <-- ERROR: same as from_class_key
}
```

### Wrong: Association Class Same as To Class
```json
{
    "name": "Contains",
    "from_class_key": "order",
    "to_class_key": "line_item",
    "association_class_key": "line_item"    <-- ERROR: same as to_class_key
}
```

### Correct: Distinct Association Class
```json
{
    "name": "For",
    "from_class_key": "book_order",
    "to_class_key": "medium",
    "from_multiplicity": "*",
    "to_multiplicity": "1..*",
    "association_class_key": "book_order_line"    <-- Correct: distinct class
}
```

## When to Use Association Classes

Use an association class when:
- The relationship itself has attributes (e.g., quantity, date, status)
- You need to track history of the relationship
- The relationship has its own lifecycle or behavior

Don't use an association class when:
- The relationship is simple with no additional data
- One class already contains all necessary relationship information

## Related Errors

- **E11004**: Association class not found
- **E11002**: From class not found
- **E11003**: To class not found
