tt Programming Language Backend Architecture

# Goals
- Easy support for different architectures and OSs
- Easily Optimisable on most levels
- Good Performance

# Architecture

AST --> Type Checking --> TAST --> IR Emission --> TTIR --> Codegen --> TTASM --> Emit --> FASM -> Binary

TTIR: TT Intermediate Representation is the Representation that the AST gets turned into. This will be mostly be used for optimissing and abstracting away from Assembly
TAST: Typed Ast

## Type Checking
Passes:
- Type Inference
- Type Checking
