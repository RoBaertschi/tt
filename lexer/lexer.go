package lexer

import (
	"fmt"
	"iter"
	"unicode"
	"unicode/utf8"

	"robaertschi.xyz/robaertschi/tt/token"
)

type ErrorCallback func(token.Loc, string, ...any)

type Lexer struct {
	input        string
	position     int
	readPosition int
	ch           rune

	linePosition int
	lineCount    int

	errors        int
	errorCallback ErrorCallback

	file string
}

func New(input string, file string) (*Lexer, error) {
	l := &Lexer{input: input, file: file}
	if err := l.readChar(); err != nil {
		return nil, err
	}
	return l, nil
}

func (l *Lexer) Iter() iter.Seq[token.Token] {
	return func(yield func(token.Token) bool) {
		for {
			if !yield(l.NextToken()) {
				return
			}
		}
	}
}

func (l *Lexer) WithErrorCallback(errorCallback ErrorCallback) {
	l.errorCallback = errorCallback
}

func (l *Lexer) loc() token.Loc {
	return token.Loc{
		Line: l.lineCount,
		Col:  l.position - l.linePosition,
		Pos:  l.position,
		File: l.file,
	}
}

func (l *Lexer) NextToken() token.Token {
	l.skipWhitespace()
	var tok token.Token
	tok.Loc = l.loc()

	switch l.ch {
	case ';':
		tok = l.newToken(token.Semicolon)
	case '=':
		if l.peekByte() == '=' {
			pos := l.position
			l.readChar()
			l.readChar()
			tok.Type = token.DoubleEqual
			tok.Literal = l.input[pos:l.position]
			return tok
		}
		tok = l.newToken(token.Equal)
	case '<':
		if l.peekByte() == '=' {
			pos := l.position
			l.readChar()
			l.readChar()
			tok.Type = token.LessThanEqual
			tok.Literal = l.input[pos:l.position]
			return tok
		}
		tok = l.newToken(token.LessThan)
	case '>':
		if l.peekByte() == '=' {
			pos := l.position
			l.readChar()
			l.readChar()
			tok.Type = token.GreaterThanEqual
			tok.Literal = l.input[pos:l.position]
			return tok
		}
		tok = l.newToken(token.GreaterThan)
	case '(':
		tok = l.newToken(token.OpenParen)
	case ')':
		tok = l.newToken(token.CloseParen)
	case '+':
		tok = l.newToken(token.Plus)
	case '-':
		tok = l.newToken(token.Minus)
	case '*':
		tok = l.newToken(token.Asterisk)
	case '/':
		tok = l.newToken(token.Slash)
	case '{':
		tok = l.newToken(token.OpenBrack)
	case '}':
		tok = l.newToken(token.CloseBrack)
	case '!':
		if l.peekByte() == '=' {
			pos := l.position
			l.readChar()
			l.readChar()
			tok.Type = token.NotEqual
			tok.Literal = l.input[pos:l.position]
			return tok
		}
		tok = l.newToken(token.Illegal)
	case -1:
		tok.Literal = ""
		tok.Type = token.Eof
	default:
		if isNumber(l.ch) {
			tok.Literal = l.readInteger()
			tok.Type = token.Int
			return tok
		} else if unicode.IsLetter(l.ch) {
			tok.Literal = l.readIdentifier()
			tok.Type = token.LookupKeyword(tok.Literal)
			return tok
		} else {
			if l.errorCallback != nil {
				l.errorCallback(tok.Loc, "Unknown character %r", l.ch)
			}
			tok = l.newToken(token.Illegal)
		}
	}
	if err := l.readChar(); err != nil {
		if l.errorCallback != nil {
			l.errorCallback(tok.Loc, "%v", err.Error())
		}
	}
	return tok
}

func (l *Lexer) newToken(t token.TokenType) token.Token {
	return token.Token{
		Type:    t,
		Literal: string(l.ch),
		Loc:     l.loc(),
	}
}

func (l *Lexer) readChar() (err error) {
	if l.readPosition < len(l.input) {
		l.position = l.readPosition
		if l.ch == '\n' {
			l.linePosition = l.position
			l.lineCount += 1
		}
		r, w := utf8.DecodeRuneInString(l.input[l.readPosition:])
		if r == utf8.RuneError && w == 1 {
			err = fmt.Errorf("Found illegal UTF-8 encoding")
		} else if r == '\uFEFF' && l.position > 0 {
			err = fmt.Errorf("Found illegal BOM")
		}
		l.readPosition += w
		l.ch = r
	} else {
		l.position = len(l.input)
		if l.ch == '\n' {
			l.linePosition = l.position
			l.lineCount += 1
		}
		l.ch = -1
	}
	return
}

func (l *Lexer) peekByte() byte {
	if l.readPosition < len(l.input) {
		return l.input[l.readPosition]
	} else {
		return 0
	}
}

func (l *Lexer) readIdentifier() string {
	startPos := l.position

	for unicode.IsLetter(l.ch) || isNumber(l.ch) || l.ch == '_' {
		l.readChar()
	}

	return l.input[startPos:l.position]
}

func (l *Lexer) readInteger() string {
	startPos := l.position

	for isNumber(l.ch) {
		l.readChar()
	}

	return l.input[startPos:l.position]
}

func isNumber(ch rune) bool {
	return '0' <= ch && ch <= '9'
}

func (l *Lexer) skipWhitespace() {
	for unicode.IsSpace(l.ch) {
		l.readChar()
	}
}

func (l *Lexer) error(loc token.Loc, format string, args ...any) {
	if l.errorCallback != nil {
		l.errorCallback(loc, format, args...)
	}

	l.errors += 1
}
