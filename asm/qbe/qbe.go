package qbe

import (
	"fmt"
	"io"

	"robaertschi.xyz/robaertschi/tt/ast"
	"robaertschi.xyz/robaertschi/tt/ttir"

	_ "embed"
)

//go:embed qbe_stub.asm
var Stub string

func emitf(w io.Writer, format string, args ...any) error {
	_, err := w.Write([]byte(fmt.Sprintf(format, args...)))
	return err
}

func Emit(output io.Writer, input *ttir.Program) error {
	for _, f := range input.Functions {
		err := emitFunction(output, f)
		if err != nil {
			return err
		}
	}
	return nil
}

func emitFunction(w io.Writer, f ttir.Function) error {
	emitf(w, "export function ")
	if f.HasReturnValue {
		if err := emitf(w, "l "); err != nil {
			return err
		}
	}
	if err := emitf(w, "$%s() {\n@start\n", f.Name); err != nil {
		return err
	}
	for _, i := range f.Instructions {
		if err := emitInstruction(w, i); err != nil {
			return err
		}
	}
	return emitf(w, "}\n")
}

func emitOperand(op ttir.Operand) string {
	switch op := op.(type) {
	case *ttir.Constant:
		return fmt.Sprintf("%d", op.Value)
	case *ttir.Var:
		return "%" + op.Value
	}
	panic(fmt.Sprintf("invalid operand %T", op))
}

func emitInstruction(w io.Writer, i ttir.Instruction) error {
	switch i := i.(type) {
	case *ttir.Ret:
		if op := i.Op; op != nil {
			if err := emitf(w, "\tret %s\n", emitOperand(i.Op)); err != nil {
				return err
			}
		} else {
			if err := emitf(w, "\tret\n"); err != nil {
				return err
			}

		}
	case *ttir.Binary:
		var inst string
		switch i.Operator {
		case ast.Add:
			inst = "add"
		case ast.Subtract:
			inst = "sub"
		case ast.Multiply:
			inst = "mul"
		case ast.Divide:
			inst = "div"
		case ast.GreaterThan:
			inst = "csgtl"
		case ast.GreaterThanEqual:
			inst = "csgel"
		case ast.LessThan:
			inst = "csltl"
		case ast.LessThanEqual:
			inst = "cslel"
		}
		if err := emitf(w, "\t%s =l %s %s, %s\n", emitOperand(i.Dst), inst, emitOperand(i.Lhs), emitOperand(i.Rhs)); err != nil {
			return err
		}

	}

	return nil
}
