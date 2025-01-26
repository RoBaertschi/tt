package typechecker

import (
	"errors"
	"fmt"

	"robaertschi.xyz/robaertschi/tt/ast"
	"robaertschi.xyz/robaertschi/tt/tast"
	"robaertschi.xyz/robaertschi/tt/token"
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
	}
	return fmt.Errorf("unhandled expression in type checker")
}
