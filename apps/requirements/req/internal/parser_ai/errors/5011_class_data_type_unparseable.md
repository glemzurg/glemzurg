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
- Using `integer`/`float`/`number` — there are no primitive numeric types; use a **span** instead, e.g. `[0 .. unconstrained] at 1 count` for integers or `[0 .. unconstrained] at 0.01 dollars` for decimals
- Using `boolean` — there is no boolean type; use `enum of true, false` instead
- Using `string` — there is no string type; use `unconstrained` for free text, `enum of x, y, z` for a fixed set of values, or `ref from Source Name` for externally documented values (e.g. ISO country codes)
- Missing `of` keyword in collections or enums
- Missing `at` keyword and precision in spans
- Invalid span bracket syntax
