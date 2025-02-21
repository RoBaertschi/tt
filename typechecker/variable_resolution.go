package typechecker

import (
	"fmt"

	"robaertschi.xyz/robaertschi/tt/ast"
	"robaertschi.xyz/robaertschi/tt/types"
)

type Variable struct {
	Name string
	Type types.Type
}

type Scope struct {
	Variables   map[string]Variable
	ParentScope *Scope
}

func (s *Scope) Get(name string) (Variable, bool) {
	v, ok := s.Variables[name]

	if ok {
		return v, true
	}

	if s.ParentScope != nil {
		return s.ParentScope.Get(name)
	}

	return Variable{}, false
}

func (s *Scope) Set(name string, t types.Type) {
	s.Variables[name] = Variable{Name: name, Type: t}
}

func (s *Scope) Has(name string) bool {
	_, ok := s.Variables[name]

	if !ok && s.ParentScope != nil {
		return s.ParentScope.Has(name)
	}

	return ok
}

func VarResolve(p *ast.Program) (Scope, error) {
	s := Scope{Variables: make(map[string]Variable)}

	for _, d := range p.Declarations {
		switch d := d.(type) {
		case *ast.FunctionDeclaration:
			err := VarResolveExpr(&s, d.Body)
			if err != nil {
				return s, err
			}
		}
	}

	return s, nil
}

func VarResolveExpr(s *Scope, e ast.Expression) error {
	switch e := e.(type) {
	case *ast.ErrorExpression:
	case *ast.AssignmentExpression:
	case *ast.BinaryExpression:
	case *ast.BlockExpression:
	case *ast.BooleanExpression:
	case *ast.IfExpression:
	case *ast.IntegerExpression:
	case *ast.VariableDeclaration:
	case *ast.VariableReference:
	default:
		panic(fmt.Sprintf("unexpected ast.Expression: %#v", e))
	}

	return nil
}
