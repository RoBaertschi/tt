# tt

## Syntax

```tt
fn main() = {
    i := 3;
    i
};
```

## Reference

### Basic Datatypes

#### Numbers

There is currently only one number type, `i64`, `i64` is a signed integer of the size of 64 bits.

#### Booleans

The boolean type `bool` can be either true or false, nothing else, it's size is implementation dependend and is only guaranteed to be 1 bit big.

### Expressions

There are many types of expression, tt is expression oriented.

#### Integer Expression

A Integer Expression contains an untyped, non-floating point, integer.
```tt
1234567890
100000
```
The Integer Expression must at minimum support the largest number type.

#### Boolean Expression
Is either the keyword `true` or `false`.
```tt
true
false
```

#### Binary Expression

A Binary Expression is a expression with two expression and an operator between them. A Operator has a precedence, that deteirmines, which way they have to be parsed.

##### Operators
- `+` Adds two numbers with the same type together
- `-` Subtracts the left expression with the right expression, they have the same type
- `*`
