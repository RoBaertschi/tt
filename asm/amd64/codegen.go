package amd64

import (
	"fmt"

	"robaertschi.xyz/robaertschi/tt/ast"
	"robaertschi.xyz/robaertschi/tt/ttir"
)

func toAsmOperand(op ttir.Operand) Operand {
	switch op := op.(type) {
	case *ttir.Constant:
		return Imm(op.Value)
	case *ttir.Var:
		return Pseudo(op.Value)
	default:
		panic(fmt.Sprintf("unkown operand %T", op))
	}
}

func CgProgram(prog *ttir.Program) Program {
	funcs := make([]Function, 0)

	for _, f := range prog.Functions {
		funcs = append(funcs, cgFunction(f))
	}

	newProgram := Program{
		Functions: funcs,
	}

	newProgram = replacePseudo(newProgram)
	newProgram = instructionFixup(newProgram)

	return newProgram
}

func cgFunction(f ttir.Function) Function {
	newInstructions := []Instruction{}

	for _, inst := range f.Instructions {
		newInstructions = append(newInstructions, cgInstruction(inst)...)
	}

	return Function{
		Name:         f.Name,
		Instructions: newInstructions,
	}
}

func cgInstruction(i ttir.Instruction) []Instruction {
	switch i := i.(type) {
	case *ttir.Ret:
		return []Instruction{
			{
				Opcode: Mov,
				Lhs:    AX,
				Rhs:    toAsmOperand(i.Op),
			},
			{
				Opcode: Ret,
			},
		}
	case *ttir.Binary:
		return cgBinary(i)
	}

	return []Instruction{}
}

func cgBinary(b *ttir.Binary) []Instruction {
	switch b.Operator {
	case ast.Add, ast.Subtract, ast.Multiply:
		var opcode Opcode
		switch b.Operator {
		case ast.Add:
			opcode = Add
		case ast.Subtract:
			opcode = Sub
		case ast.Multiply:
			opcode = Imul
		}

		return []Instruction{
			{Opcode: Mov, Lhs: toAsmOperand(b.Dst), Rhs: toAsmOperand(b.Lhs)},
			{Opcode: opcode, Lhs: toAsmOperand(b.Dst), Rhs: toAsmOperand(b.Rhs)},
		}
	case ast.Divide:
		return []Instruction{
			{Opcode: Mov, Lhs: Register(AX), Rhs: toAsmOperand(b.Lhs)},
			{Opcode: Cdq},
			{Opcode: Idiv, Lhs: toAsmOperand(b.Rhs)},
			{Opcode: Mov, Lhs: toAsmOperand(b.Dst), Rhs: Register(AX)},
		}
	}

	panic(fmt.Sprintf("unknown binary operator, %v", b))
}

// Second pass, replace all the pseudos with stack addresses
type replacePseudoPass struct {
	identToOffset map[string]int64
	currentOffset int64
}

func replacePseudo(prog Program) Program {
	newFunctions := make([]Function, 0)

	for _, f := range prog.Functions {
		newFunctions = append(newFunctions, rpFunction(f))
	}

	return Program{Functions: newFunctions}
}

func rpFunction(f Function) Function {
	newInstructions := make([]Instruction, 0)

	r := &replacePseudoPass{
		identToOffset: make(map[string]int64),
	}
	for _, i := range f.Instructions {
		newInstructions = append(newInstructions, rpInstruction(i, r))
	}

	return Function{Instructions: newInstructions, Name: f.Name, StackOffset: r.currentOffset}
}

func rpInstruction(i Instruction, r *replacePseudoPass) Instruction {

	newInstruction := Instruction{Opcode: i.Opcode}
	if i.Lhs != nil {
		newInstruction.Lhs = pseudoToStack(i.Lhs, r)
	}
	if i.Rhs != nil {
		newInstruction.Rhs = pseudoToStack(i.Rhs, r)
	}

	return newInstruction
}

func pseudoToStack(op Operand, r *replacePseudoPass) Operand {
	if pseudo, ok := op.(Pseudo); ok {
		if offset, ok := r.identToOffset[string(pseudo)]; ok {
			return Stack(offset)
		} else {
			r.currentOffset -= 4
			r.identToOffset[string(pseudo)] = r.currentOffset
			return Stack(r.currentOffset)
		}
	}
	return op
}

// Third pass, fixup invalid instructions

func instructionFixup(prog Program) Program {
	newFuncs := make([]Function, 0)

	for _, f := range prog.Functions {
		newFuncs = append(newFuncs, fixupFunction(f))
	}

	return Program{Functions: newFuncs}
}

func fixupFunction(f Function) Function {
	// The function will at minimum require the same amount of instructions, but never less
	newInstructions := make([]Instruction, 0)

	for _, i := range f.Instructions {
		newInstructions = append(newInstructions, fixupInstruction(i)...)
	}

	return Function{Name: f.Name, Instructions: newInstructions, StackOffset: f.StackOffset}
}

func fixupInstruction(i Instruction) []Instruction {

	switch i.Opcode {
	case Mov:
		if lhs, ok := i.Lhs.(Stack); ok {
			if rhs, ok := i.Rhs.(Stack); ok {
				return []Instruction{
					{Opcode: Mov, Lhs: Register(R10), Rhs: rhs},
					{Opcode: Mov, Lhs: lhs, Rhs: Register(R10)},
				}
			}
		}
	case Imul:
		if lhs, ok := i.Lhs.(Stack); ok {
			return []Instruction{
				{Opcode: Mov, Lhs: Register(R11), Rhs: lhs},
				{Opcode: Imul, Lhs: Register(R11), Rhs: i.Rhs},
				{Opcode: Mov, Lhs: lhs, Rhs: Register(R11)},
			}
		}
		fallthrough
	case Add, Sub, Idiv /* Imul (fallthrough) */ :
		if lhs, ok := i.Lhs.(Stack); ok {
			if rhs, ok := i.Rhs.(Stack); ok {
				return []Instruction{
					{Opcode: Mov, Lhs: Register(R10), Rhs: rhs},
					{Opcode: i.Opcode, Lhs: lhs, Rhs: Register(R10)},
				}
			}
		} else if lhs, ok := i.Lhs.(Imm); ok && i.Opcode == Idiv {
			return []Instruction{
				{Opcode: Mov, Lhs: Register(R10), Rhs: lhs},
				{Opcode: Idiv, Lhs: Register(R10)},
			}
		}
	}

	return []Instruction{i}
}
