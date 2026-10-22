package lexer

type Lexer struct {
	input        string
	position     int  // текущая позиция в input (указывает на текущий символ)
	readPosition int  // следующая позиция для чтения
	ch           byte // текущий символ
	line         int
	column       int
}

func New(input string) *Lexer {
	l := &Lexer{input: input, line: 1, column: 0}
	l.readChar()
	return l
}

func (l *Lexer) readChar() {
	if l.readPosition >= len(l.input) {
		l.ch = 0 // ASCII код NUL
	} else {
		l.ch = l.input[l.readPosition]
	}
	l.position = l.readPosition
	l.readPosition++
	l.column++
}

func (l *Lexer) NextToken() Token {
	var tok Token

	l.skipWhitespace()

	line, col := l.line, l.column

	switch l.ch {
	case '=':
		tok = newToken(ASSIGN, l.ch, line, col)
	case ';':
		tok = newToken(SEMICOLON, l.ch, line, col)
	case '(':
		tok = newToken(LPAREN, l.ch, line, col)
	case ')':
		tok = newToken(RPAREN, l.ch, line, col)
	case '{':
		tok = newToken(LBRACE, l.ch, line, col)
	case '}':
		tok = newToken(RBRACE, l.ch, line, col)
	case '[':
		tok = newToken(LBRACKET, l.ch, line, col)
	case ']':
		tok = newToken(RBRACKET, l.ch, line, col)
	case ':':
		tok = newToken(COLON, l.ch, line, col)
	case ',':
		tok = newToken(COMMA, l.ch, line, col)
	case '+':
		tok = newToken(PLUS, l.ch, line, col)
	case '-':
		tok = newToken(MINUS, l.ch, line, col)
	case '!':
		tok = newToken(BANG, l.ch, line, col)
	case '*':
		tok = newToken(ASTERISK, l.ch, line, col)
	case '/':
		tok = newToken(SLASH, l.ch, line, col)
	case '<':
		tok = newToken(LT, l.ch, line, col)
	case '>':
		tok = newToken(GT, l.ch, line, col)
	case '"':
		tok.Type = STRING
		tok.Literal = l.readString()
		tok.Line = line
		tok.Column = col
	case 0:
		tok.Literal = ""
		tok.Type = EOF
		tok.Line = line
		tok.Column = col
	default:
		if isLetter(l.ch) {
			tok.Literal = l.readIdentifier()
			tok.Type = LookupIdent(tok.Literal)
			tok.Line = line
			tok.Column = col
			return tok
		} else if isDigit(l.ch) {
			tok.Type = INT
			tok.Literal = l.readNumber()
			tok.Line = line
			tok.Column = col
			return tok
		} else {
			tok = newToken(ILLEGAL, l.ch, line, col)
		}
	}

	l.readChar()
	return tok
}

func newToken(tokenType TokenType, ch byte, line, col int) Token {
	return Token{Type: tokenType, Literal: string(ch), Line: line, Column: col}
}

func (l *Lexer) readIdentifier() string {
	position := l.position
	for isLetter(l.ch) || isDigit(l.ch) || l.ch == '_' {
		l.readChar()
	}
	return l.input[position:l.position]
}

func (l *Lexer) readNumber() string {
	position := l.position
	for isDigit(l.ch) {
		l.readChar()
	}
	return l.input[position:l.position]
}

func (l *Lexer) readString() string {
	position := l.position + 1
	for {
		l.readChar()
		if l.ch == '"' || l.ch == 0 {
			break
		}
	}
	return l.input[position:l.position]
}

func (l *Lexer) skipWhitespace() {
	for l.ch == ' ' || l.ch == '\t' || l.ch == '\n' || l.ch == '\r' {
		if l.ch == '\n' {
			l.line++
			l.column = 0
		}
		l.readChar()
	}
}

func isLetter(ch byte) bool {
	return 'a' <= ch && ch <= 'z' || 'A' <= ch && ch <= 'Z' || ch == '_'
}

func isDigit(ch byte) bool {
	return '0' <= ch && ch <= '9'
}
