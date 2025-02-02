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
	"else":  Else,
	"false": False,
	"fn":    Fn,
	"if":    If,
	"in":    In,
	"true":  True,
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
	Plus             TokenType = "+"
	Minus            TokenType = "-"
	Asterisk         TokenType = "*"
	Slash            TokenType = "/"
	DoubleEqual      TokenType = "=="
	NotEqual         TokenType = "!="
	LessThan         TokenType = "<"
	LessThanEqual    TokenType = "<="
	GreaterThan      TokenType = ">"
	GreaterThanEqual TokenType = ">="

	// Keywords
	Else  TokenType = "ELSE"
	False TokenType = "FALSE"
	Fn    TokenType = "FN"
	If    TokenType = "IF"
	In    TokenType = "IN"
	True  TokenType = "TRUE"
)

func LookupKeyword(literal string) TokenType {
	if value, ok := keywords[literal]; ok {
		return value
	}
	return Ident
}
