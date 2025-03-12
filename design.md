# Designs

Playground for language design dessisions

## Function Calls

```tt
fn hi(arg1: i32, arg2: i32) = {
    arg1 + arg2
    // Hi
};

fn main() = {

    // Args

    arg1 := 2;
    arg2 := 2;

    //hi(arg1, arg2) |> hi(arg2);
    //hi(arg1, arg2) |> hi(arg1, |);
};

```

## ABI

ABI Considarations:

Each file is a module. Everything inside a module is prefixed with the module name. For Example:

```asm
# - In module1 folder
# mod module1;
# fn test(): i64 = ...
module1_test:
# - In module1/module2 folder
# mod module2;
# fn test(): i64 = ...
module1_module2_test:
```

Except the main module, all members of the main module are exposed with their concrete name.

If we ever add something like generics, that will have to be encoded in the function name too.
