package lexer

import (
	"fmt"
	"testing"

	"robaertschi.xyz/robaertschi/tt/token"
)

type lexerTest struct {
	input         string
	expectedToken []token.Token
}

func runLexerTest(t *testing.T, test lexerTest) {
	t.Helper()

	l, err := New(test.input, "test.tt")
	l.WithErrorCallback(func(l token.Loc, s string, a ...any) {
		format := fmt.Sprintf(s, a)
		t.Errorf("Lexer error callback called: %s:%d:%d %s", l.File, l.Line, l.Col, format)
	})
	if err != nil {
		t.Errorf("creating lexer failed: %v", err)
	}

	for i, expectedToken := range test.expectedToken {
		actualToken := l.NextToken()
		t.Logf("expected: %v, got: %v", expectedToken, actualToken)

		if expectedToken.Literal != actualToken.Literal {
			t.Errorf("%d: expected literal %q, got %q", i, expectedToken.Literal, actualToken.Literal)
		}

		if expectedToken.Type != actualToken.Type {
			t.Errorf("%d: expected type %q, got %q", i, expectedToken.Type, actualToken.Type)
		}
	}
}

func TestBasicFunctionality(t *testing.T) {
	runLexerTest(t, lexerTest{
		input: "fn main() = 0;",
		expectedToken: []token.Token{
			{Type: token.FN, Literal: "fn"},
			{Type: token.IDENT, Literal: "main"},
			{Type: token.OPEN_PAREN, Literal: "("},
			{Type: token.CLOSE_PAREN, Literal: ")"},
			{Type: token.EQUAL, Literal: "="},
			{Type: token.INT, Literal: "0"},
			{Type: token.SEMICOLON, Literal: ";"},
			{Type: token.EOF, Literal: ""},
		},
	})
}
