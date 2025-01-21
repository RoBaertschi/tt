package ttir

import (
	"fmt"
	"strings"
)

type Program struct {
	Functions []Function
}

type Function struct {
	Name         string
	Instructions []Instruction
}

func (f *Function) String() string {
	var builder strings.Builder
	builder.WriteString(fmt.Sprintf("fn %s\n", f.Name))
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
	Op Operand
}

func (r *Ret) String() string {
	return fmt.Sprintf("ret %s\n", r.Op)
}
func (r *Ret) instruction() {}

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
