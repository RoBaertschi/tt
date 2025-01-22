package amd64

import (
	"fmt"
	"strings"
)

type Program struct {
	Functions []Function
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

func (p *Program) Emit() string {
	var builder strings.Builder
	builder.WriteString(executableAsmHeader)

	for _, function := range p.Functions {
		builder.WriteString(function.Emit())
		builder.WriteString("\n")
	}

	return builder.String()
}

type Function struct {
	StackOffset  int64
	Name         string
	Instructions []Instruction
}

func (f *Function) Emit() string {
	var builder strings.Builder

	builder.WriteString(fmt.Sprintf("%s:\n  push rbp\n  mov rbp, rsp\n  add rsp, %d\n", f.Name, f.StackOffset))

	for _, inst := range f.Instructions {
		builder.WriteString(fmt.Sprintf("  %s\n", inst.InstructionString()))
	}

	return builder.String()
}

type Opcode string

const (

	// Two operands
	Mov  Opcode = "mov" // Lhs: dst, Rhs: src, or better said intel syntax
	Add  Opcode = "add"
	Sub  Opcode = "sub"
	Imul Opcode = "imul"

	// One operand
	Idiv Opcode = "idiv"

	// No operands
	Ret Opcode = "ret"
	Cdq Opcode = "cdq"
)

type Instruction struct {
	Opcode Opcode
	// Dst
	Lhs Operand
	// Src
	Rhs Operand
}

func (i *Instruction) InstructionString() string {

	if i.Opcode == Ret {
		return fmt.Sprintf("mov rsp, rbp\n  pop rbp\n  ret\n")
	}

	// No operands
	if i.Lhs == nil {
		return fmt.Sprintf("%s", i.Opcode)
	}

	// One operand
	if i.Rhs == nil {
		return fmt.Sprintf("%s %s", i.Opcode, i.Lhs.OperandString(Eight))
	}

	// Two operands
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
	R11

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
	case R11:
		switch size {
		case One:
			return "r11b"
		case Four:
			return "r11d"
		}
		return "r11"
	}
	return "INVALID_REGISTER"
}

type Imm int64

func (i Imm) OperandString(size OperandSize) string {
	return fmt.Sprintf("%d", i)
}

type Stack int64

func (s Stack) OperandString(size OperandSize) string {

	var sizeString string
	switch size {
	case One:
		sizeString = "byte"
	case Four:
		sizeString = "dword"
	case Eight:
		sizeString = "qword"
	}

	return fmt.Sprintf("%s [rsp %+d]", sizeString, s)
}

type Pseudo string

func (s Pseudo) OperandString(size OperandSize) string {
	panic("Pseudo Operands cannot be represented in asm")
}
