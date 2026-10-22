package lexer

// TokenType определяет тип токена
type TokenType string

const (
	ILLEGAL = "ILLEGAL"
	EOF     = "EOF"

	// Идентификаторы + литералы
	IDENT  = "IDENT"  // msg, chat_input
	INT    = "INT"    // 1, 2, 3
	STRING = "STRING" // "alert text"

	// Операторы
	ASSIGN   = "="
	PLUS     = "+"
	MINUS    = "-"
	BANG     = "!"
	ASTERISK = "*"
	SLASH    = "/"
	LT       = "<"
	GT       = ">"

	// Разделители
	COMMA     = ","
	SEMICOLON = ";"
	LPAREN    = "("
	RPAREN    = ")"
	LBRACE    = "{"
	RBRACE    = "}"
	LBRACKET  = "["
	RBRACKET  = "]"
	COLON     = ":"

	// Ключевые слова Vigil
	STREAM   = "STREAM"
	ON       = "ON"
	AS       = "AS"
	IF       = "IF"
	LEVEL    = "LEVEL"
	MASK     = "MASK"
	ALERT    = "ALERT"
	TRANSMIT = "TRANSMIT"
	TO       = "TO"

	// Новые ключевые слова для InfoSec
	TAGS          = "TAGS"
	ALLOWED_ROLES = "ALLOWED_ROLES"
)

// Token представляет собой одну лексему
type Token struct {
	Type    TokenType
	Literal string
	Line    int
	Column  int
}

// LookupIdent проверяет, является ли идентификатор ключевым словом
var keywords = map[string]TokenType{
	"stream":        STREAM,
	"on":            ON,
	"as":            AS,
	"if":            IF,
	"level":         LEVEL,
	"mask":          MASK,
	"alert":         ALERT,
	"transmit":      TRANSMIT,
	"to":            TO,
	"tags":          TAGS,
	"allowed_roles": ALLOWED_ROLES,
}

func LookupIdent(ident string) TokenType {
	if tok, ok := keywords[ident]; ok {
		return tok
	}
	return IDENT
}
