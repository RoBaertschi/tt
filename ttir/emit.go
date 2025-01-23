package ttir

import (
	"fmt"

	"robaertschi.xyz/robaertschi/tt/tast"
)

var uniqueId int64

func temp() string {
	uniqueId += 1
	return fmt.Sprintf("temp.%d", uniqueId)
}

func EmitProgram(program *tast.Program) *Program {
	functions := make([]Function, 0)
	for _, decl := range program.Declarations {
		switch decl := decl.(type) {
		case *tast.FunctionDeclaration:
			functions = append(functions, *emitFunction(decl))
		}
	}

	return &Program{
		Functions: functions,
	}
}

func emitFunction(function *tast.FunctionDeclaration) *Function {
	value, instructions := emitExpression(function.Body)
	instructions = append(instructions, &Ret{Op: value})
	return &Function{
		Name:         function.Name,
		Instructions: instructions,
	}
}

func emitExpression(expr tast.Expression) (Operand, []Instruction) {
	switch expr := expr.(type) {
	case *tast.IntegerExpression:
		return &Constant{Value: expr.Value}, []Instruction{}
	case *tast.BinaryExpression:
		switch expr.Operator {
		default:
			lhsDst, instructions := emitExpression(expr.Lhs)
			rhsDst, rhsInstructions := emitExpression(expr.Rhs)
			instructions = append(instructions, rhsInstructions...)
			dst := &Var{Value: temp()}
			instructions = append(instructions, &Binary{Operator: expr.Operator, Lhs: lhsDst, Rhs: rhsDst, Dst: dst})
			return dst, instructions
		}
	}
	panic("unhandled tast.Expression case in ir emitter")
}
