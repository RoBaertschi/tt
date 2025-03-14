package typechecker

import (
	"errors"
	"fmt"

	"robaertschi.xyz/robaertschi/tt/ast"
	"robaertschi.xyz/robaertschi/tt/tast"
	"robaertschi.xyz/robaertschi/tt/types"
)

func (c *Checker) inferTypes(program *ast.Program) (*tast.Program, error) {
	c.functionVariables = make(map[string]Variables)
	decls := []tast.Declaration{}
	errs := []error{}
	vars := make(Variables)

	funcToParams := make(map[string][]tast.Parameter)

	for _, decl := range program.Declarations {
		switch decl := decl.(type) {
		case *ast.FunctionDeclaration:
			parameters := []tast.Parameter{}
			for _, param := range decl.Parameters {
				t, ok := types.From(param.Type)
				if !ok {
					return nil, c.error(decl.Token, "could not find the type %q for argument %q", param.Type, param.Name)
				}
				vars[param.Name] = t
				parameters = append(parameters, tast.Parameter{Name: param.Name, Type: t})
			}

			t, ok := types.From(decl.ReturnType)
			if !ok {
				return nil, c.error(decl.Token, "invalid type %q", decl.ReturnType)
			}

			parameterTypes := []types.Type{}

			for _, param := range parameters {
				parameterTypes = append(parameterTypes, param.Type)
			}

			vars[decl.Name] = &types.FunctionType{ReturnType: t, Parameters: parameterTypes}
			funcToParams[decl.Name] = parameters
		}
	}

	for _, decl := range program.Declarations {
		decl, err := c.inferDeclaration(funcToParams, copyVars(vars), decl)
		if err == nil {
			decls = append(decls, decl)
		} else {
			errs = append(errs, err)
		}
	}

	return &tast.Program{Declarations: decls}, errors.Join(errs...)
}

func (c *Checker) inferDeclaration(funcToParams map[string][]tast.Parameter, vars Variables, decl ast.Declaration) (tast.Declaration, error) {
	switch decl := decl.(type) {
	case *ast.FunctionDeclaration:
		body, err := c.inferExpression(vars, decl.Body)
		c.functionVariables[decl.Name] = vars

		if err != nil {
			return nil, err
		}

		return &tast.FunctionDeclaration{Token: decl.Token, Parameters: funcToParams[decl.Name], Body: body, ReturnType: vars[decl.Name], Name: decl.Name}, nil
	}
	return nil, errors.New("unhandled declaration in type inferer")
}

func (c *Checker) inferExpression(vars Variables, expr ast.Expression) (tast.Expression, error) {
	switch expr := expr.(type) {
	case *ast.IntegerExpression:
		return &tast.IntegerExpression{Token: expr.Token, Value: expr.Value}, nil
	case *ast.BooleanExpression:
		return &tast.BooleanExpression{Token: expr.Token, Value: expr.Value}, nil
	case *ast.ErrorExpression:
		return nil, c.error(expr.InvalidToken, "invalid expression")
	case *ast.BinaryExpression:
		lhs, lhsErr := c.inferExpression(vars, expr.Lhs)
		rhs, rhsErr := c.inferExpression(vars, expr.Rhs)
		var resultType types.Type
		if lhsErr == nil && rhsErr == nil {
			if expr.Operator.IsBooleanOperator() {
				resultType = types.Bool
			} else {
				resultType = lhs.Type()
			}
		}

		return &tast.BinaryExpression{Lhs: lhs, Rhs: rhs, Operator: expr.Operator, Token: expr.Token, ResultType: resultType}, errors.Join(lhsErr, rhsErr)
	case *ast.BlockExpression:
		expressions := []tast.Expression{}
		errs := []error{}

		for _, expr := range expr.Expressions {
			newExpr, err := c.inferExpression(vars, expr)
			if err != nil {
				errs = append(errs, err)
			} else {
				expressions = append(expressions, newExpr)
			}
		}

		var returnExpr tast.Expression
		var returnType types.Type
		if expr.ReturnExpression != nil {
			expr, err := c.inferExpression(vars, expr.ReturnExpression)
			returnExpr = expr
			if err != nil {
				errs = append(errs, err)
			} else {
				returnType = returnExpr.Type()
			}
		} else {
			returnType = types.Unit
		}

		return &tast.BlockExpression{
			Token:            expr.Token,
			Expressions:      expressions,
			ReturnType:       returnType,
			ReturnExpression: returnExpr,
		}, errors.Join(errs...)

	case *ast.IfExpression:
		cond, condErr := c.inferExpression(vars, expr.Condition)
		then, thenErr := c.inferExpression(vars, expr.Then)

		if expr.Else != nil {
			elseExpr, elseErr := c.inferExpression(vars, expr.Else)

			return &tast.IfExpression{Token: expr.Token, Condition: cond, Then: then, Else: elseExpr, ReturnType: then.Type()}, errors.Join(condErr, thenErr, elseErr)
		}
		return &tast.IfExpression{Token: expr.Token, Condition: cond, Then: then, Else: nil, ReturnType: types.Unit}, errors.Join(condErr, thenErr)
	case *ast.AssignmentExpression:
		varRef, ok := expr.Lhs.(*ast.VariableReference)
		if !ok {
			return &tast.AssignmentExpression{}, c.error(expr.Token, "not a valid assignment target")
		}

		rhs, err := c.inferExpression(vars, expr.Rhs)
		if err != nil {
			return &tast.AssignmentExpression{}, err
		}

		varRefT, err := c.inferExpression(vars, varRef)
		return &tast.AssignmentExpression{Lhs: varRefT, Rhs: rhs, Token: expr.Token}, err
	case *ast.VariableDeclaration:
		vd := &tast.VariableDeclaration{}
		var t types.Type
		var initializingExpr tast.Expression

		if expr.Type != "" {
			var ok bool
			t, ok = types.From(expr.Type)
			if !ok {
				return vd, c.error(expr.Token, "could not find the type %q", expr.Type)
			}
			var err error
			initializingExpr, err = c.inferExpression(vars, expr.InitializingExpression)
			if err != nil {
				return vd, err
			}
		} else {
			var err error
			initializingExpr, err = c.inferExpression(vars, expr.InitializingExpression)
			if err != nil {
				return vd, err
			}

			t = initializingExpr.Type()
		}

		vd.VariableType = t
		vars[expr.Identifier] = t

		vd.InitializingExpression = initializingExpr
		vd.Token = expr.Token
		vd.Identifier = expr.Identifier
		return vd, nil
	case *ast.VariableReference:
		vr := &tast.VariableReference{Identifier: expr.Identifier, Token: expr.Token}

		t, ok := vars[expr.Identifier]
		if !ok {
			return vr, c.error(expr.Token, "could not get type for variable %q", vr.Identifier)
		}

		vr.VariableType = t

		return vr, nil
	case *ast.FunctionCall:
		fc := &tast.FunctionCall{Identifier: expr.Identifier, Token: expr.Token}

		t, ok := vars[expr.Identifier]
		if !ok {
			return fc, c.error(expr.Token, "could not get type for function %q", fc.Identifier)
		}

		funcType, ok := t.(*types.FunctionType)
		if !ok {
			return fc, c.error(expr.Token, "tried to call non function variable %q with type %q", expr.Identifier, t.Name())
		}

		fc.ReturnType = funcType.ReturnType

		args := []tast.Expression{}
		errs := []error{}

		for _, arg := range expr.Arguments {
			inferredArg, err := c.inferExpression(vars, arg)
			errs = append(errs, err)

			if err == nil {
				args = append(args, inferredArg)
			}
		}

		fc.Arguments = args

		return fc, errors.Join(errs...)

	default:
		panic(fmt.Sprintf("unexpected ast.Expression: %#v", expr))
	}
	return nil, fmt.Errorf("unhandled expression in type inferer")
}
