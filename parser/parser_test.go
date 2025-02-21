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
		expectDeclaration(t, decl, actual.Declarations[i])
	}
}

func expectDeclaration(t *testing.T, expected ast.Declaration, actual ast.Declaration) {
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

	if expected == nil {
		if actual != nil {
			t.Errorf("expected a nil expression but got %v", actual)
		}
		return
	}

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
	case *ast.BlockExpression:
		blockExpr, ok := actual.(*ast.BlockExpression)
		if !ok {
			t.Errorf("expected %T, got %T", expected, actual)
			return
		}

		if len(expected.Expressions) != len(blockExpr.Expressions) {
			t.Errorf("expected block with %d expressions, got %d", len(expected.Expressions), len(blockExpr.Expressions))
			return
		}
		for i, expectedExpression := range expected.Expressions {
			expectExpression(t, expectedExpression, blockExpr.Expressions[i])
		}
		expectExpression(t, expected.ReturnExpression, blockExpr.ReturnExpression)
	case *ast.VariableDeclaration:
		varDecl, ok := actual.(*ast.VariableDeclaration)
		if !ok {
			t.Errorf("expected %T, got %T", expected, actual)
			return
		}

		if expected.Identifier != varDecl.Identifier {
			t.Errorf("expected variable identifier to be %q, got %q", expected.Identifier, varDecl.Identifier)
		}

		if expected.Type != varDecl.Type {
			t.Errorf("expected variable type to be %q, got %q", expected.Type, varDecl.Type)
		}

		expectExpression(t, expected.InitializingExpression, varDecl.InitializingExpression)
	case *ast.VariableReference:
		varRef, ok := actual.(*ast.VariableReference)

		if !ok {
			t.Errorf("expected %T, got %T", expected, actual)
			return
		}

		if expected.Identifier != varRef.Identifier {
			t.Errorf("expected variable reference identifier to be %q but got %q", expected.Identifier, varRef.Identifier)
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

func TestBlockExpression(t *testing.T) {
	test := parserTest{
		input: "fn main() = {\n3;\n{ 3+2 }\n}\n;",
		expectedProgram: ast.Program{
			Declarations: []ast.Declaration{
				&ast.FunctionDeclaration{
					Name: "main",
					Body: &ast.BlockExpression{
						Expressions: []ast.Expression{
							&ast.IntegerExpression{Value: 3},
						},
						ReturnExpression: &ast.BlockExpression{
							Expressions: []ast.Expression{},
							ReturnExpression: &ast.BinaryExpression{
								Lhs:      &ast.IntegerExpression{Value: 3},
								Rhs:      &ast.IntegerExpression{Value: 2},
								Operator: ast.Add,
							},
						},
					},
				},
			},
		},
	}
	runParserTest(test, t)
}

func TestGroupedExpression(t *testing.T) {
	test := parserTest{
		input: "fn main() = (3);",
		expectedProgram: ast.Program{
			Declarations: []ast.Declaration{
				&ast.FunctionDeclaration{
					Name: "main",
					Body: &ast.IntegerExpression{Value: 3},
				},
			},
		},
	}
	runParserTest(test, t)
}

func TestVariableExpression(t *testing.T) {
	test := parserTest{
		input: "fn main() = { x : u32 = 3; x };",
		expectedProgram: ast.Program{
			Declarations: []ast.Declaration{
				&ast.FunctionDeclaration{
					Name: "main",
					Body: &ast.BlockExpression{
						Expressions: []ast.Expression{
							&ast.VariableDeclaration{
								InitializingExpression: &ast.IntegerExpression{Value: 3},
								Identifier:             "x",
								Type:                   "u32",
							},
						},
						ReturnExpression: &ast.VariableReference{Identifier: "x"},
					},
				},
			},
		},
	}
	runParserTest(test, t)
}
