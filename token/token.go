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
	"fn":    Fn,
	"true":  True,
	"false": False,
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
	OpenBrack  TokenType = "{"
	CloseBrack TokenType = "}"

	// Binary Operators
	Plus        TokenType = "+"
	Minus       TokenType = "-"
	Asterisk    TokenType = "*"
	Slash       TokenType = "/"
	DoubleEqual TokenType = "=="
	NotEqual    TokenType = "!="

	// Keywords
	Fn    TokenType = "FN"
	True  TokenType = "TRUE"
	False TokenType = "FALSE"
)

func LookupKeyword(literal string) TokenType {
	if value, ok := keywords[literal]; ok {
		return value
	}
	return Ident
}
