package ast

import (
	"fmt"
	"strings"

	"robaertschi.xyz/robaertschi/tt/token"
)

type Node interface {
	TokenLiteral() string
	Tok() token.Token
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

func (p *Program) Tok() token.Token {
	if len(p.Declarations) > 0 {
		return p.Declarations[0].Tok()
	}
	return token.Token{}
}

func (p *Program) String() string {
	var builder strings.Builder

	for _, decl := range p.Declarations {
		builder.WriteString(decl.String())
		builder.WriteRune('\n')
	}

	return builder.String()
}

type Type string

type Argument struct {
	Name string
	Type Type
}

type FunctionDeclaration struct {
	Token token.Token // The token.FN
	Body  Expression
	Name  string
	Args  []Argument
}

func ArgsToString(args []Argument) string {
	var b strings.Builder

	for _, arg := range args {
		b.WriteString(fmt.Sprintf("%s %s,", arg.Name, arg.Type))
	}

	return b.String()
}

func (fd *FunctionDeclaration) declarationNode()     {}
func (fd *FunctionDeclaration) TokenLiteral() string { return fd.Token.Literal }
func (fd *FunctionDeclaration) Tok() token.Token     { return fd.Token }
func (fd *FunctionDeclaration) String() string {
	return fmt.Sprintf("fn %v(%v) = %v;", fd.Name, ArgsToString(fd.Args), fd.Body.String())
}

// Represents a Expression that we failed to parse
type ErrorExpression struct {
	InvalidToken token.Token
}

func (e *ErrorExpression) expressionNode()      {}
func (e *ErrorExpression) TokenLiteral() string { return e.InvalidToken.Literal }
func (e *ErrorExpression) Tok() token.Token     { return e.InvalidToken }
func (e *ErrorExpression) String() string       { return "<ERROR EXPR>" }

type IntegerExpression struct {
	Token token.Token // The token.INT
	Value int64
}

func (ie *IntegerExpression) expressionNode()      {}
func (ie *IntegerExpression) TokenLiteral() string { return ie.Token.Literal }
func (ie *IntegerExpression) Tok() token.Token     { return ie.Token }
func (ie *IntegerExpression) String() string       { return ie.Token.Literal }

type BooleanExpression struct {
	Token token.Token // The token.TRUE or token.FALSE
	Value bool
}

func (be *BooleanExpression) expressionNode()      {}
func (be *BooleanExpression) TokenLiteral() string { return be.Token.Literal }
func (be *BooleanExpression) Tok() token.Token     { return be.Token }
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
	LessThan
	LessThanEqual
	GreaterThan
	GreaterThanEqual
)

func (bo BinaryOperator) IsBooleanOperator() bool {
	return bo == Equal || bo == NotEqual || bo == LessThan || bo == LessThanEqual || bo == GreaterThan || bo == GreaterThanEqual
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
	case LessThan:
		return "<"
	case LessThanEqual:
		return "<="
	case GreaterThan:
		return ">"
	case GreaterThanEqual:
		return ">="
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
func (be *BinaryExpression) Tok() token.Token     { return be.Token }
func (be *BinaryExpression) String() string {
	return fmt.Sprintf("(%s %s %s)", be.Lhs, be.Operator.SymbolString(), be.Rhs)
}

type BlockExpression struct {
	Token       token.Token // The '{'
	Expressions []Expression
	// NOTE: Nullable
	//
	// A expression that does not end with a semicolon, there can only be one of those and it has to be at the end
	ReturnExpression Expression
}

func (be *BlockExpression) expressionNode()      {}
func (be *BlockExpression) TokenLiteral() string { return be.Token.Literal }
func (be *BlockExpression) Tok() token.Token     { return be.Token }
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
	// NOTE: Can be nil
	Else Expression
}

func (ie *IfExpression) expressionNode()      {}
func (ie *IfExpression) TokenLiteral() string { return ie.Token.Literal }
func (ie *IfExpression) Tok() token.Token     { return ie.Token }
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

type VariableDeclaration struct {
	Token                  token.Token // The Identifier token
	InitializingExpression Expression
	Type                   Type
	Identifier             string
}

func (vd *VariableDeclaration) expressionNode()      {}
func (vd *VariableDeclaration) TokenLiteral() string { return vd.Token.Literal }
func (vd *VariableDeclaration) Tok() token.Token     { return vd.Token }
func (vd *VariableDeclaration) String() string {
	return fmt.Sprintf("%s : %v = %s", vd.Identifier, vd.Type, vd.InitializingExpression)
}

type VariableReference struct {
	Token      token.Token // The identifier token
	Identifier string
}

func (vr *VariableReference) expressionNode()      {}
func (vr *VariableReference) TokenLiteral() string { return vr.Token.Literal }
func (vr *VariableReference) Tok() token.Token     { return vr.Token }
func (vr *VariableReference) String() string {
	return fmt.Sprintf("%s", vr.Identifier)
}

type AssignmentExpression struct {
	Token token.Token // The Equal
	Lhs   Expression
	Rhs   Expression
}

func (ae *AssignmentExpression) expressionNode()      {}
func (ae *AssignmentExpression) TokenLiteral() string { return ae.Token.Literal }
func (ae *AssignmentExpression) Tok() token.Token     { return ae.Token }
func (ae *AssignmentExpression) String() string {
	return fmt.Sprintf("%s = %s", ae.Lhs.String(), ae.Rhs.String())
}
