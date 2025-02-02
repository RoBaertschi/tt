package ttir

import (
	"fmt"

	"robaertschi.xyz/robaertschi/tt/tast"
	"robaertschi.xyz/robaertschi/tt/types"
)

var uniqueTempId int64

func temp() string {
	uniqueTempId += 1
	return fmt.Sprintf("temp.%d", uniqueTempId)
}

var uniqueLabelId int64

func tempLabel() string {
	uniqueLabelId += 1
	return fmt.Sprintf("lbl.%d", uniqueLabelId)
}

func EmitProgram(program *tast.Program) *Program {
	functions := make([]*Function, 0)
	var mainFunction *Function
	for _, decl := range program.Declarations {
		switch decl := decl.(type) {
		case *tast.FunctionDeclaration:
			f := emitFunction(decl)
			functions = append(functions, f)
			if f.Name == "main" {
				mainFunction = f
			}
		}
	}

	return &Program{
		Functions:    functions,
		MainFunction: mainFunction,
	}
}

func emitFunction(function *tast.FunctionDeclaration) *Function {
	value, instructions := emitExpression(function.Body)
	instructions = append(instructions, &Ret{Op: value})
	f := &Function{
		Name:           function.Name,
		Instructions:   instructions,
		HasReturnValue: !function.ReturnType.IsSameType(types.Unit),
	}

	return f
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
	case *tast.IfExpression:
		// if (cond -> false jump to "else") {
		//     ...
		// } jump to end of if
		// else: else {
		//     ...
		// } endOfIf:
		elseLabel := tempLabel()
		endOfIfLabel := tempLabel()
		var dst Operand = &Var{Value: temp()}

		condDst, instructions := emitExpression(expr.Condition)

		instructions = append(instructions, &JumpIfZero{Value: condDst, Label: elseLabel})
		thenDst, thenInstructions := emitExpression(expr.Then)
		instructions = append(instructions, thenInstructions...)
		if !expr.ReturnType.IsSameType(types.Unit) {
			instructions = append(instructions, &Copy{Src: thenDst, Dst: dst}, Jump(endOfIfLabel))
		} else {
			dst = nil
		}

		instructions = append(instructions, Label(elseLabel))
		if expr.Else != nil {
			elseDst, elseInstructions := emitExpression(expr.Else)
			instructions = append(instructions, elseInstructions...)
			if !expr.ReturnType.IsSameType(types.Unit) {
				instructions = append(instructions, &Copy{Src: elseDst, Dst: dst})
			}
		}
		instructions = append(instructions, Label(endOfIfLabel))
		return dst, instructions
	}
	panic("unhandled tast.Expression case in ir emitter")
}
