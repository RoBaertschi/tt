package qbe

import (
	"fmt"
	"io"
	"strings"

	"robaertschi.xyz/robaertschi/tt/ast"
	"robaertschi.xyz/robaertschi/tt/ttir"

	_ "embed"
)

var extraLabelId int64 = 0

func extraLabel() string {
	extraLabelId += 1
	return fmt.Sprintf("qbe.extra.%d", extraLabelId)
}

//go:embed qbe_stub.asm
var Stub string

func emitf(w io.Writer, format string, args ...any) error {
	_, err := w.Write(fmt.Appendf(nil, format, args...))
	return err
}

func emit(w io.Writer, content string) error {
	_, err := w.Write(fmt.Append(nil, content))
	return err
}

func Emit(output io.Writer, input *ttir.Program) error {
	if input.MainFunction.HasReturnValue {
		emitf(output, `
export function $_start() {
@start
    %%result =l call $main()
    call $syscall1(l 60, l %%result)
    hlt
}
            `,
		)
	} else {
		emitf(output, `
export function $_start() {
@start
    call $main()
    call $syscall1(l 60, l 0)
    hlt
}
`,
		)
	}

	for _, f := range input.Functions {
		err := emitFunction(output, f)
		if err != nil {
			return err
		}
	}
	return nil
}

func emitFunction(w io.Writer, f *ttir.Function) error {
	emitf(w, "export function ")
	if f.HasReturnValue {
		if err := emitf(w, "l "); err != nil {
			return err
		}
	}

	b := strings.Builder{}

	for i, arg := range f.Arguments {
		if i > 0 {
			b.WriteString(", ")
		}
		b.WriteString("l %")
		b.WriteString(arg)
	}

	if err := emitf(w, "$%s(%v) {\n@start\n", f.Name, b.String()); err != nil {
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
			return emitf(w, "\tret %s\n", emitOperand(i.Op))
		} else {
			return emitf(w, "\tret\n")
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
		case ast.Equal:
			inst = "ceql"
		case ast.NotEqual:
			inst = "cnel"
		case ast.GreaterThan:
			inst = "csgtl"
		case ast.GreaterThanEqual:
			inst = "csgel"
		case ast.LessThan:
			inst = "csltl"
		case ast.LessThanEqual:
			inst = "cslel"
		}
		return emitf(w, "\t%s =l %s %s, %s\n", emitOperand(i.Dst), inst, emitOperand(i.Lhs), emitOperand(i.Rhs))
	case *ttir.Copy:
		emitf(w, "\t%s =l copy %s\n", emitOperand(i.Dst), emitOperand(i.Src))
	case ttir.Label:
		return emitf(w, "@%s\n", string(i))
	case ttir.Jump:
		return emitf(w, "\tjmp @%s\n", string(i))
	case *ttir.JumpIfNotZero:
		after := extraLabel()
		return emitf(w, "\tjnz %s, @%s, @%s\n@%s\n", emitOperand(i.Value), i.Label, after, after)
	case *ttir.JumpIfZero:
		after := extraLabel()
		return emitf(w, "\tjnz %s, @%s, @%s\n@%s\n", emitOperand(i.Value), after, i.Label, after)
	case *ttir.Call:
		b := strings.Builder{}
		b.WriteRune('\t')
		if i.ReturnValue != nil {
			b.WriteString(emitOperand(i.ReturnValue) + " =l ")
		}

		b.WriteString("call $" + i.FunctionName + "(")
		for j, arg := range i.Arguments {
			b.WriteString("l " + emitOperand(arg))
			if j < (len(i.Arguments) - 1) {
				b.WriteString(", ")
			}
		}

		b.WriteString(")\n")
		return emit(w, b.String())
	default:
		panic("unkown instruction")
	}

	return nil
}
