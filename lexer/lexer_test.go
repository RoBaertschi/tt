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
		format := fmt.Sprintf(s, a...)
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
		input: "fn main() = 0 + 3;",
		expectedToken: []token.Token{
			{Type: token.Fn, Literal: "fn"},
			{Type: token.Ident, Literal: "main"},
			{Type: token.OpenParen, Literal: "("},
			{Type: token.CloseParen, Literal: ")"},
			{Type: token.Equal, Literal: "="},
			{Type: token.Int, Literal: "0"},
			{Type: token.Plus, Literal: "+"},
			{Type: token.Int, Literal: "3"},
			{Type: token.Semicolon, Literal: ";"},
			{Type: token.Eof, Literal: ""},
		},
	})
}
