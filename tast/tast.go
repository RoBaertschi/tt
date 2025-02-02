// Typed AST
// Almost identical to the AST, but contains types and contains only correct types.
// Also, it does not contain a Error Expression, because, that an previous error

package tast

import (
	"fmt"
	"strings"

	"robaertschi.xyz/robaertschi/tt/ast"
	"robaertschi.xyz/robaertschi/tt/token"
	"robaertschi.xyz/robaertschi/tt/types"
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
	Type() types.Type
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
	Token      token.Token // The token.FN
	Body       Expression
	Name       string
	ReturnType types.Type
}

var _ Declaration = &FunctionDeclaration{}

func (fd *FunctionDeclaration) declarationNode()     {}
func (fd *FunctionDeclaration) TokenLiteral() string { return fd.Token.Literal }
func (fd *FunctionDeclaration) String() string {
	return fmt.Sprintf("fn %v(): %v = %v;", fd.Name, fd.ReturnType.Name(), fd.Body.String())
}

type IntegerExpression struct {
	Token token.Token // The token.INT
	Value int64
}

var _ Expression = &IntegerExpression{}

func (ie *IntegerExpression) expressionNode() {}
func (ie *IntegerExpression) Type() types.Type {
	return types.I64
}
func (ie *IntegerExpression) TokenLiteral() string { return ie.Token.Literal }
func (ie *IntegerExpression) String() string       { return ie.Token.Literal }

type BooleanExpression struct {
	Token token.Token // The token.TRUE or token.FALSE
	Value bool
}

var _ Expression = &BooleanExpression{}

func (be *BooleanExpression) expressionNode() {}
func (ie *BooleanExpression) Type() types.Type {
	return types.Bool
}
func (be *BooleanExpression) TokenLiteral() string { return be.Token.Literal }
func (be *BooleanExpression) String() string       { return be.Token.Literal }

type BinaryExpression struct {
	Token      token.Token // The operator
	Lhs, Rhs   Expression
	Operator   ast.BinaryOperator
	ResultType types.Type
}

var _ Expression = &BinaryExpression{}

func (be *BinaryExpression) expressionNode() {}
func (be *BinaryExpression) Type() types.Type {
	return be.ResultType
}
func (be *BinaryExpression) TokenLiteral() string { return be.Token.Literal }
func (be *BinaryExpression) String() string {
	return fmt.Sprintf("(%s %s %s :> %s)", be.Lhs, be.Operator.SymbolString(), be.Rhs, be.ResultType.Name())
}

type BlockExpression struct {
	Token            token.Token // The '{'
	Expressions      []Expression
	ReturnExpression Expression // A expression that does not end with a semicolon, there can only be one of those and it hast to be at the end
	ReturnType       types.Type
}

func (be *BlockExpression) expressionNode() {}
func (be *BlockExpression) Type() types.Type {
	return be.ReturnType
}
func (be *BlockExpression) TokenLiteral() string { return be.Token.Literal }
func (be *BlockExpression) String() string {
	var builder strings.Builder

	builder.WriteString("({\n")
	for _, expr := range be.Expressions {
		builder.WriteString("\t")
		builder.WriteString(expr.String())
		builder.WriteString(";\n")
	}
	if be.ReturnExpression != nil {
		builder.WriteString(fmt.Sprintf("\t%s\n", be.ReturnExpression.String()))
	}
	builder.WriteString("})")

	return builder.String()
}

type IfExpression struct {
	Token     token.Token // The 'if' token
	Condition Expression
	Then      Expression
	// Can be nil
	Else       Expression
	ReturnType types.Type
}

func (ie *IfExpression) expressionNode() {}
func (ie *IfExpression) Type() types.Type {
	return ie.ReturnType
}
func (ie *IfExpression) TokenLiteral() string { return ie.Token.Literal }
func (ie *IfExpression) String() string {
	var builder strings.Builder

	builder.WriteString("(if\n\t")
	builder.WriteString(ie.Then.String())

	if ie.Else != nil {
		builder.WriteString(" else in ")
		builder.WriteString(ie.Else.String())
	}
	builder.WriteString(")")

	return builder.String()
}
