// Typed AST
// Almost identical to the AST, but contains types and contains only correct types.
// Also, it does not contain a Error Expression, because, that an previous error

package tast

import (
	"fmt"
	"strings"

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

func (fd *FunctionDeclaration) declarationNode()     {}
func (fd *FunctionDeclaration) TokenLiteral() string { return fd.Token.Literal }
func (fd *FunctionDeclaration) String() string {
	return fmt.Sprintf("fn %v(): %v = %v;", fd.Name, fd.ReturnType.Name(), fd.Body.String())
}

type IntegerExpression struct {
	Token token.Token // The token.INT
	Value int64
}

func (ie *IntegerExpression) expressionNode() {}
func (ie *IntegerExpression) Type() types.Type {
	return types.I64
}
func (ie *IntegerExpression) TokenLiteral() string { return ie.Token.Literal }
func (ie *IntegerExpression) String() string       { return ie.Token.Literal }
