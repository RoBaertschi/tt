package parser

import (
	"fmt"
	"testing"

	"robaertschi.xyz/robaertschi/tt/ast"
	"robaertschi.xyz/robaertschi/tt/lexer"
	"robaertschi.xyz/robaertschi/tt/token"
)

type parserTest struct {
	input           string
	expectedProgram ast.Program
}

func runParserTest(test parserTest, t *testing.T) {
	t.Helper()
	l, err := lexer.New(test.input, "test.tt")
	l.WithErrorCallback(func(l token.Loc, s string, a ...any) {
		format := fmt.Sprintf(s, a...)
		t.Errorf("Lexer error callback called: %s:%d:%d %s", l.File, l.Line, l.Col, format)
	})

	if err != nil {
		t.Errorf("creating lexer failed: %v", err)
	}

	p := New(l)
	p.WithErrorCallback(func(tok token.Token, s string, a ...any) {
		format := fmt.Sprintf(s, a...)
		t.Errorf("Parser error callback called: %s:%d:%d %s", tok.Loc.File, tok.Loc.Line, tok.Loc.Col, format)
	})

	actual := p.ParseProgram()

	if p.errors > 0 {
		t.Fatalf("parser errors: %d", p.errors)
	}

	if len(actual.Declarations) != len(test.expectedProgram.Declarations) {
		t.Fatalf("expected %d declarations, got %d", len(test.expectedProgram.Declarations), len(actual.Declarations))
	}

	for i, decl := range test.expectedProgram.Declarations {
		expectDeclarationSame(t, decl, actual.Declarations[i])
	}
}

func expectDeclarationSame(t *testing.T, expected ast.Declaration, actual ast.Declaration) {
	t.Helper()

	switch expected := expected.(type) {
	case *ast.FunctionDeclaration:
		actual, ok := actual.(*ast.FunctionDeclaration)
		if !ok {
			t.Errorf("expected function declaration, got %T", actual)
			return
		}
		if actual.Name != expected.Name {
			t.Errorf("expected function name %s, got %s", expected.Name, actual.Name)
		}

		expectExpression(t, expected.Body, actual.Body)
	}
}

func expectExpression(t *testing.T, expected ast.Expression, actual ast.Expression) {
	t.Helper()

	switch expected := expected.(type) {
	case *ast.ErrorExpression:
		actual, ok := actual.(*ast.ErrorExpression)
		if !ok {
			t.Errorf("expected error expression, got %T", actual)
			return
		}
		if actual.InvalidToken != expected.InvalidToken {
			t.Errorf("expected invalid token %v, got %v", expected.InvalidToken, actual.InvalidToken)
		}
	case *ast.IntegerExpression:
		integerExpr, ok := actual.(*ast.IntegerExpression)
		if !ok {
			t.Errorf("expected *ast.IntegerExpression, got %T", actual)
			return
		}
		if integerExpr.Value != expected.Value {
			t.Errorf("expected integer value %d, got %d", expected.Value, integerExpr.Value)
		}
	case *ast.BinaryExpression:
		binaryExpr, ok := actual.(*ast.BinaryExpression)
		if !ok {
			t.Errorf("expected %T, got %T", expected, actual)
			return
		}

		if binaryExpr.Operator != expected.Operator {
			t.Errorf("expected %q operator for binary expression, got %q", expected.Operator.SymbolString(), binaryExpr.Operator.SymbolString())
		}
		expectExpression(t, expected.Lhs, binaryExpr.Lhs)
		expectExpression(t, expected.Rhs, binaryExpr.Rhs)
	case *ast.BooleanExpression:
		booleanExpr, ok := actual.(*ast.BooleanExpression)
		if !ok {
			t.Errorf("expected %T, got %T", expected, actual)
			return
		}

		if booleanExpr.Value != expected.Value {
			t.Errorf("expected boolean %v, got %v", expected.Value, booleanExpr.Value)
		}
	default:
		t.Fatalf("unknown expression type %T", expected)
	}
}

func TestFunctionDeclaration(t *testing.T) {
	test := parserTest{
		input: "fn main() = 0;",
		expectedProgram: ast.Program{
			Declarations: []ast.Declaration{
				&ast.FunctionDeclaration{
					Name: "main",
					Body: &ast.IntegerExpression{Value: 0, Token: token.Token{Type: token.Int, Literal: "0"}},
				},
			},
		},
	}
	runParserTest(test, t)
}

func TestBinaryExpressions(t *testing.T) {
	test := parserTest{
		input: "fn main() = true == true == true;",
		expectedProgram: ast.Program{
			Declarations: []ast.Declaration{
				&ast.FunctionDeclaration{
					Name: "main",
					Body: &ast.BinaryExpression{
						Lhs: &ast.BinaryExpression{
							Lhs:      &ast.BooleanExpression{Value: true},
							Rhs:      &ast.BooleanExpression{Value: true},
							Operator: ast.Equal,
						},
						Rhs:      &ast.BooleanExpression{Value: true},
						Operator: ast.Equal,
					},
				},
			},
		},
	}

	runParserTest(test, t)
}
