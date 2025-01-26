package ttir

import (
	"fmt"

	"robaertschi.xyz/robaertschi/tt/tast"
	"robaertschi.xyz/robaertschi/tt/types"
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
		Name:           function.Name,
		Instructions:   instructions,
		HasReturnValue: !function.ReturnType.IsSameType(types.Unit),
	}
}

func emitExpression(expr tast.Expression) (Operand, []Instruction) {
	switch expr := expr.(type) {
	case *tast.IntegerExpression:
		return &Constant{Value: expr.Value}, []Instruction{}
	case *tast.BooleanExpression:
		value := int64(0)
		if expr.Value {
			value = 1
		}
		return &Constant{Value: value}, []Instruction{}
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
	case *tast.BlockExpression:
		instructions := []Instruction{}

		for _, expr := range expr.Expressions {
			_, insts := emitExpression(expr)
			instructions = append(instructions, insts...)
		}

		var value Operand
		if expr.ReturnExpression != nil {
			dst, insts := emitExpression(expr.ReturnExpression)
			value = dst
			instructions = append(instructions, insts...)
		}

		return value, instructions
	}
	panic("unhandled tast.Expression case in ir emitter")
}
