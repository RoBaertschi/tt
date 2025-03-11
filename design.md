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
