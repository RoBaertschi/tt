package token

type Loc struct {
	Line int
	Col  int
	Pos  int
	File string
}

type TokenType string

type Token struct {
	Type    TokenType
	Literal string
	Loc     Loc
}

var keywords = map[string]TokenType{
	"fn": Fn,
}

const (
	Illegal TokenType = "ILLEGAL"
	Eof     TokenType = "EOF"

	Ident TokenType = "IDENT"
	Int   TokenType = "INT"

	Semicolon  TokenType = ";"
	Equal      TokenType = "="
	OpenParen  TokenType = "("
	CloseParen TokenType = ")"
	Plus       TokenType = "+"
	Minus      TokenType = "-"
	Asterisk   TokenType = "*"
	Slash      TokenType = "/"

	// Keywords
	Fn TokenType = "FN"
)

func LookupKeyword(literal string) TokenType {
	if value, ok := keywords[literal]; ok {
		return value
	}
	return Ident
}
