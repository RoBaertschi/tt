package typechecker

import (
	"errors"
	"fmt"

	"robaertschi.xyz/robaertschi/tt/ast"
	"robaertschi.xyz/robaertschi/tt/tast"
	"robaertschi.xyz/robaertschi/tt/types"
)

func (c *Checker) inferTypes(program *ast.Program) (*tast.Program, error) {
	decls := []tast.Declaration{}
	errs := []error{}

	for _, decl := range program.Declarations {
		decl, err := c.inferDeclaration(decl)
		if err == nil {
			decls = append(decls, decl)
		} else {
			errs = append(errs, err)
		}
	}

	return &tast.Program{Declarations: decls}, errors.Join(errs...)
}

func (c *Checker) inferDeclaration(decl ast.Declaration) (tast.Declaration, error) {
	switch decl := decl.(type) {
	case *ast.FunctionDeclaration:
		body, err := c.inferExpression(decl.Body)

		if err != nil {
			return nil, err
		}

		return &tast.FunctionDeclaration{Token: decl.Token, Body: body, ReturnType: body.Type(), Name: decl.Name}, nil
	}
	return nil, errors.New("unhandled declaration in type inferer")
}

func (c *Checker) inferExpression(expr ast.Expression) (tast.Expression, error) {
	switch expr := expr.(type) {
	case *ast.IntegerExpression:
		return &tast.IntegerExpression{Token: expr.Token, Value: expr.Value}, nil
	case *ast.BooleanExpression:
		return &tast.BooleanExpression{Token: expr.Token, Value: expr.Value}, nil
	case *ast.ErrorExpression:
		return nil, c.error(expr.InvalidToken, "invalid expression")
	case *ast.BinaryExpression:
		lhs, lhsErr := c.inferExpression(expr.Lhs)
		rhs, rhsErr := c.inferExpression(expr.Rhs)
		var resultType types.Type
		if lhsErr == nil && rhsErr == nil {
			if expr.Operator.IsBooleanOperator() {
				resultType = types.Bool
			} else {
				resultType = lhs.Type()
			}
		}

		return &tast.BinaryExpression{Lhs: lhs, Rhs: rhs, Operator: expr.Operator, Token: expr.Token, ResultType: resultType}, errors.Join(lhsErr, rhsErr)
	}
	return nil, fmt.Errorf("unhandled expression in type inferer")
}
