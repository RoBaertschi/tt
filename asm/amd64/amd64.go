package amd64

import (
	"fmt"
	"strings"
)

type Program struct {
	Functions []Function
}

func (p *Program) Emit() string {
	var builder strings.Builder

	for _, function := range p.Functions {
		builder.WriteString(function.Emit())
		builder.WriteString("\n")
	}

	return builder.String()
}

type Function struct {
	Name         string
	Instructions []Instruction
}

// This calls the main function and uses it's return value to exit
const executableAsmHeader = "format ELF64 executable\n" +
	"segment readable executable\n" +
	"entry _start\n" +
	"_start:\n" +
	"  call main\n" +
	"  mov rdi, rax\n" +
	"  mov rax, 60\n" +
	"  syscall\n"

func (f *Function) Emit() string {
	var builder strings.Builder

	builder.WriteString(executableAsmHeader)
	builder.WriteString(fmt.Sprintf("%s:\n", f.Name))

	for _, inst := range f.Instructions {
		builder.WriteString(fmt.Sprintf("  %s\n", inst.InstructionString()))
	}

	return builder.String()
}

type Opcode string

const (
	Mov   Opcode = "mov"
	Ret   Opcode = "ret"
	Add   Opcode = "add"
	Sub   Opcode = "sub"
	Imull Opcode = "imul"
	Idiv  Opcode = "idiv"
)

type Instruction struct {
	Opcode Opcode
	Lhs    Operand
	Rhs    Operand
}

func (i *Instruction) InstructionString() string {
	if i.Lhs == nil {
		return fmt.Sprintf("%s", i.Opcode)
	}

	return fmt.Sprintf("%s %s, %s", i.Opcode, i.Lhs.OperandString(Eight), i.Rhs.OperandString(Eight))
}

type OperandSize int

type Operand interface {
	OperandString(OperandSize) string
}

type Register int

const (
	AX Register = iota
	R10

	One OperandSize = iota
	Four
	Eight
)

func (r Register) OperandString(size OperandSize) string {
	switch r {
	case AX:
		switch size {
		case One:
			return "al"
		case Four:
			return "eax"
		}
		return "rax"
	case R10:
		switch size {
		case One:
			return "r10b"
		case Four:
			return "r10d"
		}
		return "r10"
	}
	return "INVALID_REGISTER"
}

type Imm int64

func (i Imm) OperandString(size OperandSize) string {
	return fmt.Sprintf("%d", i)
}

type Stack int64

func (s Stack) OperandString(size OperandSize) string {
	return fmt.Sprintf("rbp(%d)", s)
}
