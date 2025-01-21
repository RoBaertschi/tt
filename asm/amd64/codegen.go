package amd64

import (
	"fmt"

	"robaertschi.xyz/robaertschi/tt/ttir"
)

func toAsmOperand(op ttir.Operand) Operand {
	switch op := op.(type) {
	case *ttir.Constant:
		return Imm(op.Value)
	default:
		panic(fmt.Sprintf("unkown operand %T", op))
	}
}

func CgProgram(prog *ttir.Program) Program {
	funcs := make([]Function, 0)

	for _, f := range prog.Functions {
		funcs = append(funcs, cgFunction(f))
	}

	return Program{
		Functions: funcs,
	}
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
	}

	return []Instruction{}
}
