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
	newProgram, err := c.inferTypes(program)
	if err != nil {
		return nil, err
	}

	errs := []error{}

	for _, decl := range newProgram.Declarations {
		err := c.checkDeclaration(decl)
		if err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		return nil, errors.Join(errs...)
	}

	if !c.foundMain {
		// TODO(Robin): Add support for libraries
		errs = append(errs, errors.New("no function called 'main' found"))
	}

	return newProgram, errors.Join(errs...)
}

func (c *Checker) checkDeclaration(decl tast.Declaration) error {
	switch decl := decl.(type) {
	case *tast.FunctionDeclaration:
		err := c.checkExpression(decl.Body)

		if err != nil {
			return err
		}

		if decl.Name == "main" {
			c.foundMain = true
		}

		return nil
	}
	return errors.New("unhandled declaration in type checker")
}

func (c *Checker) checkExpression(expr tast.Expression) error {
	switch expr := expr.(type) {
	case *tast.IntegerExpression:
		return nil
	case *tast.BooleanExpression:
		return nil
	case *tast.BinaryExpression:
		lhsErr := c.checkExpression(expr.Lhs)
		rhsErr := c.checkExpression(expr.Rhs)
		var operandErr error
		if lhsErr == nil && rhsErr == nil {
			if !expr.Lhs.Type().IsSameType(expr.Rhs.Type()) {
				operandErr = c.error(expr.Token, "the lhs of the expression does not have the same type then the rhs, lhs=%q, rhs=%q", expr.Lhs.Type().Name(), expr.Rhs.Type().Name())
			} else if !expr.Lhs.Type().SupportsBinaryOperator(expr.Operator) {
				operandErr = c.error(expr.Token, "the operator %q is not supported by the type %q", expr.Operator, expr.Lhs.Type().Name())
			}
		}

		return errors.Join(lhsErr, rhsErr, operandErr)
	case *tast.BlockExpression:
		errs := []error{}

		for _, expr := range expr.Expressions {
			errs = append(errs, c.checkExpression(expr))
		}
		if expr.ReturnExpression != nil {
			errs = append(errs, c.checkExpression(expr.ReturnExpression))
		}
		return errors.Join(errs...)
	case *tast.IfExpression:
		condErr := c.checkExpression(expr.Condition)
		if condErr == nil {
			if !expr.Condition.Type().IsSameType(types.Bool) {
				condErr = c.error(expr.Token, "the condition in the if should be a boolean, but got %q", expr.Condition.Type().Name())
			}
		}
		thenErr := c.checkExpression(expr.Then)

		if expr.Else == nil {
			return errors.Join(condErr, thenErr)
		}

		elseErr := c.checkExpression(expr.Else)
		if thenErr == nil && elseErr == nil {
			if !expr.Then.Type().IsSameType(expr.Else.Type()) {
				thenErr = c.error(expr.Token, "the then branch of type %q does not match with the else branch of type %q", expr.Then.Type().Name(), expr.Else.Type().Name())
			}
		}
		return errors.Join(condErr, thenErr, elseErr)
	}
	return fmt.Errorf("unhandled expression %T in type checker", expr)
}
