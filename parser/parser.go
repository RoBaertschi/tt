package parser

import (
	"fmt"
	"strconv"

	"robaertschi.xyz/robaertschi/tt/ast"
	"robaertschi.xyz/robaertschi/tt/lexer"
	"robaertschi.xyz/robaertschi/tt/token"
)

type precedence int

const (
	PrecLowest precedence = iota
	PrecSum
	PrecProduct
)

var precedences = map[token.TokenType]precedence{
	token.Plus:     PrecSum,
	token.Minus:    PrecSum,
	token.Asterisk: PrecProduct,
	token.Slash:    PrecProduct,
}

type ErrorCallback func(token.Token, string, ...any)
type prefixParseFn func() ast.Expression
type infixParseFn func(ast.Expression) ast.Expression

type Parser struct {
	curToken  token.Token
	peekToken token.Token

	errors        int
	errorCallback ErrorCallback

	l              *lexer.Lexer
	prefixParseFns map[token.TokenType]prefixParseFn
	infixParseFns  map[token.TokenType]infixParseFn
}

func New(l *lexer.Lexer) *Parser {
	p := &Parser{l: l}

	p.prefixParseFns = make(map[token.TokenType]prefixParseFn)
	p.registerPrefixFn(token.Int, p.parseIntegerExpression)

	p.infixParseFns = make(map[token.TokenType]infixParseFn)
	p.registerInfixFn(token.Plus, p.parseBinaryExpression)
	p.registerInfixFn(token.Minus, p.parseBinaryExpression)
	p.registerInfixFn(token.Asterisk, p.parseBinaryExpression)
	p.registerInfixFn(token.Slash, p.parseBinaryExpression)
	p.registerInfixFn(token.Equal, p.parseBinaryExpression)
	p.registerInfixFn(token.NotEqual, p.parseBinaryExpression)

	p.nextToken()
	p.nextToken()

	return p
}

func (p *Parser) WithErrorCallback(errorCallback ErrorCallback) {
	p.errorCallback = errorCallback
}

func (p *Parser) Errors() int {
	return p.errors
}

func (p *Parser) registerInfixFn(tt token.TokenType, infix infixParseFn) {
	p.infixParseFns[tt] = infix
}

func (p *Parser) registerPrefixFn(tt token.TokenType, fn prefixParseFn) {
	p.prefixParseFns[tt] = fn
}

func (p *Parser) nextToken() {
	p.curToken = p.peekToken
	p.peekToken = p.l.NextToken()
}

func (p *Parser) curTokenIs(tt token.TokenType) bool {
	return p.curToken.Type == tt
}

func (p *Parser) peekTokenIs(tt token.TokenType) bool {
	return p.peekToken.Type == tt
}

func getPrecedence(tt token.TokenType) precedence {
	if prec, ok := precedences[tt]; ok {
		return prec
	}
	return PrecLowest
}

func (p *Parser) curPrecedence() precedence {
	return getPrecedence(p.curToken.Type)
}

func (p *Parser) peekPrecedence() precedence {
	return getPrecedence(p.peekToken.Type)
}

func (p *Parser) error(t token.Token, format string, args ...any) {
	if p.errorCallback != nil {
		p.errorCallback(t, format, args...)
	} else {
		fmt.Printf("%s:%d:%d ", t.Loc.File, t.Loc.Line, t.Loc.Col)
		fmt.Printf(format, args...)
		fmt.Println()
	}

	p.errors += 1
}

func (p *Parser) exprError(invalidToken token.Token, format string, args ...any) ast.Expression {
	p.error(invalidToken, format, args...)
	return &ast.ErrorExpression{
		InvalidToken: invalidToken,
	}
}

func (p *Parser) expect(tt token.TokenType) bool {
	if p.curToken.Type != tt {
		p.error(p.curToken, "expected %q, got %q", tt, p.curToken.Type)
		return false
	}
	return true
}

func (p *Parser) expectPeek(tt token.TokenType) bool {
	if p.peekToken.Type != tt {
		p.error(p.peekToken, "expected %q, got %q", tt, p.peekToken.Type)
		p.nextToken()
		return false
	}
	p.nextToken()
	return true
}

func (p *Parser) ParseProgram() *ast.Program {
	decls := []ast.Declaration{}

	for p.curToken.Type != token.Eof {
		decl := p.parseDeclaration()
		if decl != nil {
			decls = append(decls, decl)
		}
		p.nextToken()
	}

	return &ast.Program{
		Declarations: decls,
	}
}

func (p *Parser) parseDeclaration() ast.Declaration {
	if !p.expect(token.Fn) {
		return nil
	}
	tok := p.curToken
	if !p.expectPeek(token.Ident) {
		return nil
	}

	name := p.curToken.Literal
	if !p.expectPeek(token.OpenParen) {
		return nil
	}
	if !p.expectPeek(token.CloseParen) {
		return nil
	}
	if !p.expectPeek(token.Equal) {
		return nil
	}

	p.nextToken()
	expr := p.parseExpression(PrecLowest)
	if !p.expectPeek(token.Semicolon) {
		return nil
	}

	return &ast.FunctionDeclaration{
		Token: tok,
		Name:  name,
		Body:  expr,
	}
}

func (p *Parser) parseExpression(precedence precedence) ast.Expression {
	prefix := p.prefixParseFns[p.curToken.Type]
	if prefix == nil {
		return p.exprError(p.curToken, "could not parse invalid token in expression %s", p.curToken.Type)
	}

	leftExpr := prefix()

	for !p.peekTokenIs(token.Semicolon) && precedence < p.peekPrecedence() {
		infix := p.infixParseFns[p.peekToken.Type]
		if infix == nil {
			return leftExpr
		}

		p.nextToken()

		leftExpr = infix(leftExpr)
	}

	return leftExpr
}

func (p *Parser) parseIntegerExpression() ast.Expression {
	if !p.expect(token.Int) {
		return &ast.ErrorExpression{InvalidToken: p.curToken}
	}

	int := &ast.IntegerExpression{
		Token: p.curToken,
	}

	value, err := strconv.ParseInt(int.Token.Literal, 0, 64)
	if err != nil {
		return p.exprError(int.Token, "invalid integer literal: %v", err)
	}

	int.Value = value
	return int
}

func (p *Parser) parseBinaryExpression(lhs ast.Expression) ast.Expression {
	var op ast.BinaryOperator
	switch p.curToken.Type {
	case token.Plus:
		op = ast.Add
	case token.Minus:
		op = ast.Subtract
	case token.Asterisk:
		op = ast.Multiply
	case token.Slash:
		op = ast.Divide
	case token.DoubleEqual:
		op = ast.Equal
	case token.NotEqual:
		op = ast.NotEqual
	default:
		return p.exprError(p.curToken, "invalid token for binary expression %s", p.curToken.Type)
	}
	tok := p.curToken

	precedence := p.curPrecedence()
	p.nextToken()
	rhs := p.parseExpression(precedence)

	return &ast.BinaryExpression{Lhs: lhs, Rhs: rhs, Operator: op, Token: tok}
}
