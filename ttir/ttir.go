package ttir

import (
	"fmt"
	"strings"

	"robaertschi.xyz/robaertschi/tt/ast"
)

type Program struct {
	Functions    []*Function
	MainFunction *Function
}

func (p *Program) String() string {
	var builder strings.Builder
	for _, f := range p.Functions {
		builder.WriteString(f.String())
	}
	return builder.String()
}

type Function struct {
	Name           string
	Arguments      []string
	Instructions   []Instruction
	HasReturnValue bool
}

func (f *Function) String() string {
	var builder strings.Builder
	builder.WriteString(fmt.Sprintf("fn %s", f.Name))
	for _, arg := range f.Arguments {
		builder.WriteString(" " + arg)
	}
	builder.WriteRune('\n')
	for _, i := range f.Instructions {
		builder.WriteString("  ")
		builder.WriteString(i.String())
	}
	return builder.String()
}

type Instruction interface {
	String() string
	instruction()
}

type Ret struct {
	// Nullable, if it does not return anything
	Op Operand
}

func (r *Ret) String() string {
	if r.Op != nil {
		return fmt.Sprintf("ret %s\n", r.Op)
	} else {
		return "ret\n"
	}
}
func (r *Ret) instruction() {}

type Binary struct {
	Operator ast.BinaryOperator
	Lhs      Operand
	Rhs      Operand
	Dst      Operand
}

func (b *Binary) String() string {
	return fmt.Sprintf("%s = %s %s, %s\n", b.Dst, b.Operator, b.Lhs, b.Rhs)
}
func (b *Binary) instruction() {}

type Copy struct {
	Src Operand
	Dst Operand
}

func (c *Copy) String() string {
	return fmt.Sprintf("%s = copy %s\n", c.Dst, c.Src)
}
func (c *Copy) instruction() {}

type JumpIfZero struct {
	Value Operand
	Label string
}

func (jiz *JumpIfZero) String() string {
	return fmt.Sprintf("jz %v, %v\n", jiz.Value, jiz.Label)
}
func (jiz *JumpIfZero) instruction() {}

type JumpIfNotZero struct {
	Value Operand
	Label string
}

func (jiz *JumpIfNotZero) String() string {
	return fmt.Sprintf("jnz %v, %v\n", jiz.Value, jiz.Label)
}
func (jiz *JumpIfNotZero) instruction() {}

type Jump string

func (j Jump) String() string {
	return fmt.Sprintf("jmp %v\n", string(j))
}
func (j Jump) instruction() {}

type Label string

func (l Label) String() string {
	return fmt.Sprintf("%v:\n", string(l))
}
func (l Label) instruction() {}

type Call struct {
	Label       string
	Arguments   []Operand
	ReturnValue Operand
}

func (c *Call) String() string {
	b := strings.Builder{}

	b.WriteString(c.ReturnValue.String() + " = call " + c.Label + " ")

	for i, arg := range c.Arguments {
		b.WriteString(arg.String())

		if i < (len(c.Arguments) - 1) {
			b.WriteString(", ")
		}
	}

	return b.String()
}
func (c *Call) instruction() {}

type Operand interface {
	String() string
	operand()
}

type Constant struct {
	Value int64
}

func (c *Constant) String() string {
	return fmt.Sprintf("%d", c.Value)
}
func (c *Constant) operand() {}

type Var struct {
	Value string
}

func (v *Var) String() string {
	return v.Value
}
func (v *Var) operand() {}
