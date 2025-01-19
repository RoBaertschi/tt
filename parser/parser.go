package parser

import (
	"robaertschi.xyz/robaertschi/tt/lexer"
	"robaertschi.xyz/robaertschi/tt/token"
)

type ErrorCallback func(token.Token, string, ...any)

type Parser struct {
	lexer lexer.Lexer

	curToken  token.Token
	peekToken token.Token

	errors        int
	errorCallback ErrorCallback
}
