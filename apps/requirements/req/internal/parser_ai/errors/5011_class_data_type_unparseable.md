# Attribute Data Type Unparseable (E5011)

The `data_type_rules` value for a class attribute could not be parsed.

## What Went Wrong

The `data_type_rules` field contains a string that does not match any valid data type syntax.

## Valid Data Type Syntax

```
unconstrained                          (or empty string / omit the field)
enum of active, inactive, pending      enumeration of named values
ordered enum of low, medium, high      ordered enumeration
[1 .. 100] at 1 unit                   numeric span with precision
(0 .. unconstrained] at 0.01 dollars   open-ended span

ordered [unique] [1-10] of <type>      ordered collection
unordered [unique] [1-10] of <type>    unordered collection (set)
stack of <type>                        LIFO stack
queue of <type>                        FIFO queue

{                                      record type
  field_name: <type>;
  other_field: <type>
}
```

Where `<type>` is any atomic type (unconstrained, enum, span).

## How to Fix

Check spelling and syntax of `data_type_rules`. Common mistakes:
- Using `string`/`integer`/`boolean` (these are TLA+ type specs, not data_type_rules)
- Missing `of` keyword in collections or enums
- Missing `at` keyword and precision in spans
- Invalid span bracket syntax
