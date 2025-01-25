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
func (e *ErrorExpression) String() string       { return "<ERROR EXPR>" }

type IntegerExpression struct {
	Token token.Token // The token.INT
	Value int64
}

func (ie *IntegerExpression) expressionNode()      {}
func (ie *IntegerExpression) TokenLiteral() string { return ie.Token.Literal }
func (ie *IntegerExpression) String() string       { return ie.Token.Literal }

type BooleanExpression struct {
	Token token.Token // The token.TRUE or token.FALSE
	Value bool
}

func (be *BooleanExpression) expressionNode()      {}
func (be *BooleanExpression) TokenLiteral() string { return be.Token.Literal }
func (be *BooleanExpression) String() string       { return be.Token.Literal }

//go:generate stringer -type=BinaryOperator
type BinaryOperator int

const (
	Add BinaryOperator = iota
	Subtract
	Multiply
	Divide
	Equal
	NotEqual
)

func (bo BinaryOperator) IsBooleanOperator() bool {
	return bo == Equal || bo == NotEqual
}

func (bo BinaryOperator) SymbolString() string {
	switch bo {
	case Add:
		return "+"
	case Subtract:
		return "-"
	case Multiply:
		return "*"
	case Divide:
		return "/"
	case Equal:
		return "=="
	case NotEqual:
		return "!="
	}
	return "<INVALID BINARY OPERATOR>"
}

type BinaryExpression struct {
	Token    token.Token // The operator
	Lhs, Rhs Expression
	Operator BinaryOperator
}

func (be *BinaryExpression) expressionNode()      {}
func (be *BinaryExpression) TokenLiteral() string { return be.Token.Literal }
func (be *BinaryExpression) String() string {
	return fmt.Sprintf("(%s %s %s)", be.Lhs, be.Operator.SymbolString(), be.Rhs)
}
