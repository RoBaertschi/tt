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

	arguments := []string{}

	for _, arg := range function.Parameters {
		arguments = append(arguments, arg.Name)
	}

	f := &Function{
		Name:           function.Name,
		Instructions:   instructions,
		Arguments:      arguments,
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
		if expr.Else != nil {
			if !expr.ReturnType.IsSameType(types.Unit) {
				instructions = append(instructions, &Copy{Src: thenDst, Dst: dst})
			}
			instructions = append(instructions, Jump(endOfIfLabel))
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
	case *tast.AssignmentExpression:
		ident := expr.Lhs.(*tast.VariableReference)

		rhsDst, instructions := emitExpression(expr.Rhs)

		instructions = append(instructions, &Copy{Src: rhsDst, Dst: &Var{Value: ident.Identifier}})

		return nil, instructions
	case *tast.VariableDeclaration:
		rhsDst, instructions := emitExpression(expr.InitializingExpression)

		instructions = append(instructions, &Copy{Src: rhsDst, Dst: &Var{Value: expr.Identifier}})

		return nil, instructions
	case *tast.VariableReference:
		return &Var{Value: expr.Identifier}, []Instruction{}
	case *tast.FunctionCall:
		var dst Operand
		if !expr.ReturnType.IsSameType(types.Unit) {
			dst = &Var{Value: temp()}
		}
		args := []Operand{}

		instructions := []Instruction{}

		for _, arg := range expr.Arguments {
			dst, argInstructions := emitExpression(arg)

			instructions = append(instructions, argInstructions...)
			args = append(args, dst)
		}

		instructions = append(instructions, &Call{FunctionName: expr.Identifier, Arguments: args, ReturnValue: dst})
		return dst, instructions
	default:
		panic(fmt.Sprintf("unexpected tast.Expression: %#v", expr))
	}
}
