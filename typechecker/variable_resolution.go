package typechecker

import (
	"errors"
	"fmt"

	"robaertschi.xyz/robaertschi/tt/ast"
	"robaertschi.xyz/robaertschi/tt/token"
)

type Var struct {
	Name             string
	FromCurrentScope bool
}

type Scope struct {
	Variables map[string]Var
	UniqueId  int64
}

func errorf(t token.Token, format string, args ...any) error {
	return fmt.Errorf("%s:%d:%d %s", t.Loc.File, t.Loc.Line, t.Loc.Col, fmt.Sprintf(format, args...))
}

func copyScope(s *Scope) Scope {
	newVars := make(map[string]Var)

	for k, v := range s.Variables {
		newVars[k] = Var{Name: v.Name, FromCurrentScope: false}
	}

	return Scope{Variables: newVars}
}

func (s *Scope) Get(name string) (Var, bool) {
	v, ok := s.Variables[name]

	if ok {
		return v, true
	}

	return Var{}, false
}

func (s *Scope) Set(name string, uniqName string) {
	s.Variables[name] = Var{Name: uniqName, FromCurrentScope: true}
}

func (s *Scope) Has(name string) bool {
	_, ok := s.Variables[name]
	return ok
}

func (s *Scope) HasInCurrent(name string) bool {
	v, ok := s.Variables[name]
	if !ok {
		return false
	}
	return v.FromCurrentScope
}

func VarResolve(p *ast.Program) (map[string]Scope, error) {
	functionToScope := make(map[string]Scope)

	for _, d := range p.Declarations {
		switch d := d.(type) {
		case *ast.FunctionDeclaration:
			s := Scope{Variables: make(map[string]Var)}
			err := VarResolveExpr(&s, d.Body)
			functionToScope[d.Name] = s
			if err != nil {
				return functionToScope, err
			}
		}
	}

	return functionToScope, nil
}

func VarResolveExpr(s *Scope, e ast.Expression) error {
	switch e := e.(type) {
	case *ast.ErrorExpression:
		// NOTE: The Checker will take care of this
		return nil
	case *ast.AssignmentExpression:
		err := VarResolveExpr(s, e.Lhs)
		if err != nil {
			return err
		}
		err = VarResolveExpr(s, e.Rhs)
		if err != nil {
			return err
		}

	case *ast.BinaryExpression:
		err := VarResolveExpr(s, e.Lhs)
		if err != nil {
			return err
		}
		err = VarResolveExpr(s, e.Rhs)
		if err != nil {
			return err
		}
	case *ast.BlockExpression:
		newS := copyScope(s)
		errs := []error{}
		for _, e := range e.Expressions {
			errs = append(errs, VarResolveExpr(&newS, e))
		}
		errs = append(errs, VarResolveExpr(&newS, e.ReturnExpression))

		return errors.Join(errs...)
	case *ast.IfExpression:
		err := VarResolveExpr(s, e.Condition)
		if err != nil {
			return err
		}

		err = VarResolveExpr(s, e.Then)
		if err != nil {
			return err
		}

		if e.Else != nil {
			err = VarResolveExpr(s, e.Else)
			if err != nil {
				return err
			}
		}
	case *ast.VariableDeclaration:
		if s.HasInCurrent(e.Identifier) {
			return errorf(e.Token, "variable %q redifinded", e.Identifier)
		}

		uniqName := fmt.Sprintf("%s.%d", e.Identifier, s.UniqueId)
		s.UniqueId += 1
		s.Set(e.Identifier, uniqName)
	case *ast.VariableReference:
		v, ok := s.Get(e.Identifier)
		if !ok {
			return errorf(e.Token, "variable %q is not declared", e.Identifier)
		}

		e.Identifier = v.Name
	case *ast.BooleanExpression:
	case *ast.IntegerExpression:
	default:
		panic(fmt.Sprintf("unexpected ast.Expression: %#v", e))
	}

	return nil
}
