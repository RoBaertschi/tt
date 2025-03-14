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
	PrecComparison
	PrecSum
	PrecProduct
	PrecAssignment
)

var precedences = map[token.TokenType]precedence{
	token.Plus:             PrecSum,
	token.Minus:            PrecSum,
	token.Asterisk:         PrecProduct,
	token.Slash:            PrecProduct,
	token.DoubleEqual:      PrecComparison,
	token.NotEqual:         PrecComparison,
	token.GreaterThan:      PrecComparison,
	token.GreaterThanEqual: PrecComparison,
	token.LessThan:         PrecComparison,
	token.LessThanEqual:    PrecComparison,
	token.Equal:            PrecAssignment,
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
	p.registerPrefixFn(token.True, p.parseBooleanExpression)
	p.registerPrefixFn(token.False, p.parseBooleanExpression)
	p.registerPrefixFn(token.OpenParen, p.parseGroupedExpression)
	p.registerPrefixFn(token.OpenBrack, p.parseBlockExpression)
	p.registerPrefixFn(token.If, p.parseIfExpression)
	p.registerPrefixFn(token.Ident, p.parseVariable)

	p.infixParseFns = make(map[token.TokenType]infixParseFn)
	p.registerInfixFn(token.Plus, p.parseBinaryExpression)
	p.registerInfixFn(token.Minus, p.parseBinaryExpression)
	p.registerInfixFn(token.Asterisk, p.parseBinaryExpression)
	p.registerInfixFn(token.Slash, p.parseBinaryExpression)
	p.registerInfixFn(token.DoubleEqual, p.parseBinaryExpression)
	p.registerInfixFn(token.NotEqual, p.parseBinaryExpression)
	p.registerInfixFn(token.GreaterThan, p.parseBinaryExpression)
	p.registerInfixFn(token.GreaterThanEqual, p.parseBinaryExpression)
	p.registerInfixFn(token.LessThan, p.parseBinaryExpression)
	p.registerInfixFn(token.LessThanEqual, p.parseBinaryExpression)

	p.registerInfixFn(token.Equal, p.parseAssignmentExpression)

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

func (p *Parser) expect(tt token.TokenType) (bool, ast.Expression) {
	if p.curToken.Type != tt {
		p.error(p.curToken, "expected %q, got %q", tt, p.curToken.Type)
		return false, &ast.ErrorExpression{InvalidToken: p.curToken}
	}
	return true, nil
}

func (p *Parser) expectPeek(tt token.TokenType) (bool, ast.Expression) {
	if p.peekToken.Type != tt {
		p.error(p.peekToken, "expected %q, got %q", tt, p.peekToken.Type)
		p.nextToken()
		return false, nil
	}
	p.nextToken()
	return true, &ast.ErrorExpression{InvalidToken: p.curToken}
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

func (p *Parser) parseType() (t ast.Type, ok bool) {
	if ok, _ := p.expect(token.Ident); !ok {
		return "", false
	}

	return ast.Type(p.curToken.Literal), true
}

func (p *Parser) parseParameterList() ([]ast.Parameter, bool) {
	parameters := []ast.Parameter{}

	for p.peekTokenIs(token.Ident) {
		p.nextToken()
		name := p.curToken.Literal
		if ok, _ := p.expectPeek(token.Colon); !ok {
			return parameters, false
		}
		p.nextToken()
		t, ok := p.parseType()

		if !ok {
			return parameters, false
		}

		parameters = append(parameters, ast.Parameter{Type: t, Name: name})

		if !p.peekTokenIs(token.Comma) {
			break
		}
		p.nextToken()
	}

	return parameters, true
}

func (p *Parser) parseDeclaration() ast.Declaration {
	if ok, _ := p.expect(token.Fn); !ok {
		return nil
	}
	tok := p.curToken
	if ok, _ := p.expectPeek(token.Ident); !ok {
		return nil
	}

	name := p.curToken.Literal
	if ok, _ := p.expectPeek(token.OpenParen); !ok {
		return nil
	}

	params, ok := p.parseParameterList()

	if !ok {
		return nil
	}

	if ok, _ := p.expectPeek(token.CloseParen); !ok {
		return nil
	}
	if ok, _ := p.expectPeek(token.Colon); !ok {
		return nil
	}
	p.nextToken()
	t, ok := p.parseType()
	if !ok {
		return nil
	}
	if ok, _ := p.expectPeek(token.Equal); !ok {
		return nil
	}

	p.nextToken()
	expr := p.parseExpression(PrecLowest)
	if ok, _ := p.expectPeek(token.Semicolon); !ok {
		return nil
	}

	return &ast.FunctionDeclaration{
		Token:      tok,
		Name:       name,
		Body:       expr,
		Parameters: params,
		ReturnType: t,
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
	if ok, errExpr := p.expect(token.Int); !ok {
		return errExpr
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

func (p *Parser) parseBooleanExpression() ast.Expression {
	var value bool
	switch p.curToken.Type {
	case token.True:
		value = true
	case token.False:
		value = false
	default:
		return p.exprError(p.curToken, "invalid token for boolean expression %s", p.curToken.Type)
	}

	return &ast.BooleanExpression{
		Token: p.curToken,
		Value: value,
	}
}

func (p *Parser) parseGroupedExpression() ast.Expression {
	p.expect(token.OpenParen)

	p.nextToken()
	expr := p.parseExpression(PrecLowest)

	if ok, errExpr := p.expectPeek(token.CloseParen); !ok {
		return errExpr
	}

	return expr
}

func (p *Parser) parseBlockExpression() ast.Expression {
	if ok, errExpr := p.expect(token.OpenBrack); !ok {
		return errExpr
	}
	block := &ast.BlockExpression{Token: p.curToken}

	p.nextToken()
	for !p.curTokenIs(token.CloseBrack) {
		expr := p.parseExpression(PrecLowest)
		if p.peekTokenIs(token.Semicolon) {
			block.Expressions = append(block.Expressions, expr)
			p.nextToken()
			p.nextToken()
		} else if p.peekTokenIs(token.CloseBrack) {
			block.ReturnExpression = expr
			p.nextToken()
		} else {
			return p.exprError(p.peekToken, "expected a ';' or '}' to either end the current expression or block, but got %q instead.", p.peekToken.Type)
		}
	}

	return block
}

func (p *Parser) parseIfExpression() ast.Expression {
	if ok, errExpr := p.expect(token.If); !ok {
		return errExpr
	}

	ifExpr := &ast.IfExpression{Token: p.curToken}

	p.nextToken()
	ifExpr.Condition = p.parseExpression(PrecLowest)

	if p.peekTokenIs(token.OpenBrack) {
		p.nextToken()
		ifExpr.Then = p.parseBlockExpression()
	} else {
		if ok, errExpr := p.expectPeek(token.In); !ok {
			return errExpr
		}

		p.nextToken()
		ifExpr.Then = p.parseExpression(PrecLowest)
	}

	if p.peekTokenIs(token.Else) {
		p.nextToken()
		p.nextToken()
		ifExpr.Else = p.parseExpression(PrecLowest)
	} else {
		ifExpr.Else = nil
	}

	return ifExpr
}

func (p *Parser) parseVariable() ast.Expression {
	if ok, errExpr := p.expect(token.Ident); !ok {
		return errExpr
	}

	switch p.peekToken.Type {
	case token.Colon:
		return p.parseVariableDeclaration()
	case token.OpenParen:
		return p.parseFunctionCall()
	default:
		return &ast.VariableReference{
			Token:      p.curToken,
			Identifier: p.curToken.Literal,
		}
	}

	// FIXME(Robin): Add variable references
	// Lets panic about deez nuts of yours
	panic("deez nuts")
}

func (p *Parser) parseVariableDeclaration() ast.Expression {
	if ok, errExpr := p.expect(token.Ident); !ok {
		return errExpr
	}

	variable := &ast.VariableDeclaration{Token: p.curToken, Identifier: p.curToken.Literal}
	if ok, errExpr := p.expectPeek(token.Colon); !ok {
		return errExpr
	}

	if p.peekTokenIs(token.Ident) {
		p.nextToken()
		variable.Type = ast.Type(p.curToken.Literal)
	}

	if ok, errExpr := p.expectPeek(token.Equal); !ok {
		return errExpr
	}

	p.nextToken()
	variable.InitializingExpression = p.parseExpression(PrecLowest)

	return variable
}

func (p *Parser) parseFunctionCall() ast.Expression {
	if ok, errExpr := p.expect(token.Ident); !ok {
		return errExpr
	}

	funcCall := &ast.FunctionCall{Token: p.curToken, Identifier: p.curToken.Literal}
	if ok, errExpr := p.expectPeek(token.OpenParen); !ok {
		return errExpr
	}

	args := []ast.Expression{}

	for !p.peekTokenIs(token.CloseParen) {
		p.nextToken()

		args = append(args, p.parseExpression(PrecLowest))
		if !p.peekTokenIs(token.Comma) {
			break
		}
		p.nextToken()
	}

	// Move onto the ')'
	p.nextToken()

	funcCall.Arguments = args

	return funcCall
}

// Binary

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
	case token.LessThan:
		op = ast.LessThan
	case token.LessThanEqual:
		op = ast.LessThanEqual
	case token.GreaterThan:
		op = ast.GreaterThan
	case token.GreaterThanEqual:
		op = ast.GreaterThanEqual
	default:
		return p.exprError(p.curToken, "invalid token for binary expression %s", p.curToken.Type)
	}
	tok := p.curToken

	precedence := p.curPrecedence()
	p.nextToken()
	rhs := p.parseExpression(precedence)

	return &ast.BinaryExpression{Lhs: lhs, Rhs: rhs, Operator: op, Token: tok}
}

func (p *Parser) parseAssignmentExpression(lhs ast.Expression) ast.Expression {
	if ok, errExpr := p.expect(token.Equal); !ok {
		return errExpr
	}

	varAss := &ast.AssignmentExpression{
		Token: p.curToken,
		Lhs:   lhs,
	}

	p.nextToken()

	varAss.Rhs = p.parseExpression(PrecLowest)

	return varAss
}
