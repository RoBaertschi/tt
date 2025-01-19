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
	"fn": FN,
}

const (
	ILLEGAL TokenType = "ILLEGAL"
	EOF     TokenType = "EOF"

	IDENT TokenType = "IDENT"
	INT   TokenType = "INT"

	SEMICOLON   = ";"
	EQUAL       = "="
	OPEN_PAREN  = "("
	CLOSE_PAREN = ")"

	// Keywords
	FN = "FN"
)

func LookupKeyword(literal string) TokenType {
	if value, ok := keywords[literal]; ok {
		return value
	}
	return IDENT
}
