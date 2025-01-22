package typechecker

import (
	"errors"
	"fmt"

	"robaertschi.xyz/robaertschi/tt/ast"
	"robaertschi.xyz/robaertschi/tt/tast"
	"robaertschi.xyz/robaertschi/tt/token"
	"robaertschi.xyz/robaertschi/tt/types"
)

type Checker struct {
	foundMain bool
}

func New() *Checker {
	return &Checker{}
}

func (c *Checker) error(t token.Token, format string, args ...any) error {
	return fmt.Errorf("%s:%d:%d %s", t.Loc.File, t.Loc.Line, t.Loc.Col, fmt.Sprintf(format, args...))
}

func (c *Checker) CheckProgram(program *ast.Program) (*tast.Program, error) {
	decls := []tast.Declaration{}
	errs := []error{}

	for _, decl := range program.Declarations {
		decl, err := c.checkDeclaration(decl)
		if err == nil {
			decls = append(decls, decl)
		} else {
			errs = append(errs, err)
		}
	}

	if !c.foundMain {
		// TODO: Add support for libraries
		errs = append(errs, errors.New("no function called 'main' found"))
	}

	return &tast.Program{Declarations: decls}, errors.Join(errs...)
}

func (c *Checker) checkDeclaration(decl ast.Declaration) (tast.Declaration, error) {
	switch decl := decl.(type) {
	case *ast.FunctionDeclaration:
		body, err := c.checkExpression(decl.Body)

		if err != nil {
			return nil, err
		}

		if decl.Name == "main" {
			c.foundMain = true
		}

		return &tast.FunctionDeclaration{Token: decl.Token, Body: body, ReturnType: body.Type(), Name: decl.Name}, nil
	}
	return nil, errors.New("unhandled declaration in type checker")
}

func (c *Checker) checkExpression(expr ast.Expression) (tast.Expression, error) {
	switch expr := expr.(type) {
	case *ast.IntegerExpression:
		return &tast.IntegerExpression{Token: expr.Token, Value: expr.Value}, nil
	case *ast.ErrorExpression:
		return nil, c.error(expr.InvalidToken, "invalid expression")
	case *ast.BinaryExpression:
		lhs, lhsErr := c.checkExpression(expr.Lhs)
		rhs, rhsErr := c.checkExpression(expr.Rhs)
		var operandErr error
		var resultType types.Type
		if lhsErr == nil && rhsErr == nil {
			if !lhs.Type().IsSameType(rhs.Type()) {
				operandErr = fmt.Errorf("the lhs of the expression does not have the same type then the rhs, lhs=%q, rhs=%q", lhs.Type(), rhs.Type())
			} else {
				resultType = lhs.Type()
			}
		}

		return &tast.BinaryExpression{Lhs: lhs, Rhs: rhs, Operator: expr.Operator, Token: expr.Token, ResultType: resultType}, errors.Join(lhsErr, rhsErr, operandErr)
	}
	return nil, fmt.Errorf("unhandled expression in type checker")
}
