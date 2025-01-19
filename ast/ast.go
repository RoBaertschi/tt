package ast

import (
	"fmt"
	"strings"

	"robaertschi.xyz/robaertschi/tt/token"
)

type Node interface {
	TokenLiteral() string
	String() string
}

type Declaration interface {
	Node
	declarationNode()
}

type Expression interface {
	Node
	expressionNode()
}

type Program struct {
	Declarations []Declaration
}

func (p *Program) TokenLiteral() string {
	if len(p.Declarations) > 0 {
		return p.Declarations[0].TokenLiteral()
	}
	return ""
}

func (p *Program) String() string {
	var builder strings.Builder

	for _, decl := range p.Declarations {
		builder.WriteString(decl.String())
		builder.WriteRune('\n')
	}

	return builder.String()
}

type FunctionDeclaration struct {
	Token token.Token // The token.FN
	Body  Expression
	Name  string
}

func (fd *FunctionDeclaration) declarationNode()     {}
func (fd *FunctionDeclaration) TokenLiteral() string { return fd.Token.Literal }
func (fd *FunctionDeclaration) String() string {
	return fmt.Sprintf("fn %v() = %v;", fd.Name, fd.Body.String())
}

// Represents a Expression that we failed to parse
type ErrorExpression struct {
	InvalidToken token.Token
}

func (e *ErrorExpression) expressionNode()      {}
func (e *ErrorExpression) TokenLiteral() string { return e.InvalidToken.Literal }
func (e *ErrorExpression) String() string       { return "<ERROR>" }

type IntegerExpression struct {
	Token token.Token // The token.INT
	Value int64
}

func (ie *IntegerExpression) expressionNode()      {}
func (ie *IntegerExpression) TokenLiteral() string { return ie.Token.Literal }
func (ie *IntegerExpression) String() string       { return ie.Token.Literal }
