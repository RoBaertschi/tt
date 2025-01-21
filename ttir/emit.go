package ttir

import "robaertschi.xyz/robaertschi/tt/tast"

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
	instructions = append(instructions, &Ret{op: value})
	return &Function{
		Name:         function.Name,
		Instructions: instructions,
	}
}

func emitExpression(expr tast.Expression) (Operand, []Instruction) {
	switch expr := expr.(type) {
	case *tast.IntegerExpression:
		return &Constant{Value: expr.Value}, []Instruction{}
	}
	panic("unhandled tast.Expression case in ir emitter")
}
