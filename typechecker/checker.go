package typechecker

import (
	"errors"
	"fmt"

	"robaertschi.xyz/robaertschi/tt/ast"
	"robaertschi.xyz/robaertschi/tt/tast"
	"robaertschi.xyz/robaertschi/tt/token"
)

type Checker struct{}

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

	return &tast.Program{Declarations: decls}, errors.Join(errs...)
}

func (c *Checker) checkDeclaration(decl ast.Declaration) (tast.Declaration, error) {
	switch decl := decl.(type) {
	case *ast.FunctionDeclaration:
		body, err := c.checkExpression(decl.Body)

		if err != nil {
			return nil, err
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
	}
	return nil, fmt.Errorf("unhandled expression in type checker")
}
